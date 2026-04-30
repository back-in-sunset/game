package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"im/internal/auth"
	"im/internal/config"
	"im/internal/contracts"
	"im/internal/domain"
	"im/internal/goimx"
	"im/internal/mempool"
	"im/internal/pipeline"
	"im/internal/session"

	"github.com/gorilla/websocket"
)

type Server struct {
	addr      string
	nodeID    string
	auth      contracts.AuthProvider
	scope     contracts.ScopeResolver
	sessions  contracts.SessionManager
	messaging *pipeline.Messaging
	presence  interface {
		Bind(ctx context.Context, principal auth.Principal, nodeID string) error
		Unbind(ctx context.Context, principal auth.Principal, nodeID string) error
	}
	server      *http.Server
	nextID      atomic.Int64
	adapter     *goimx.Adapter
	framePool   *mempool.BufferPool
	ringSize    int
	heartbeat   time.Duration
	missLimit   int
	flushPeriod time.Duration
}

type loginRequest struct {
	Token  string `json:"token"`
	Domain string `json:"domain"`
	Scope  struct {
		TenantID    string `json:"tenant_id"`
		ProjectID   string `json:"project_id"`
		Environment string `json:"environment"`
	} `json:"scope"`
}

func New(
	addr string,
	nodeID string,
	authProvider contracts.AuthProvider,
	scopeResolver contracts.ScopeResolver,
	sessions contracts.SessionManager,
	messaging *pipeline.Messaging,
	presence interface {
		Bind(ctx context.Context, principal auth.Principal, nodeID string) error
		Unbind(ctx context.Context, principal auth.Principal, nodeID string) error
	},
	cfg config.Session,
) *Server {
	s := &Server{
		addr:        addr,
		nodeID:      nodeID,
		auth:        authProvider,
		scope:       scopeResolver,
		sessions:    sessions,
		messaging:   messaging,
		presence:    presence,
		adapter:     goimx.NewAdapter(),
		framePool:   mempool.NewBufferPool(cfg.FrameBufferSize),
		ringSize:    cfg.RingSize,
		heartbeat:   time.Duration(cfg.HeartbeatInterval) * time.Second,
		missLimit:   cfg.HeartbeatMisses,
		flushPeriod: time.Duration(cfg.WriteFlushInterval) * time.Millisecond,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	s.server = &http.Server{Addr: addr, Handler: mux}
	return s
}

func (s *Server) Start() error {
	go func() {
		_ = s.server.ListenAndServe()
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	principal, loginSeq, err := s.authenticateConn(conn)
	if err != nil {
		_ = s.writeErrorFrame(conn, 0, err.Error())
		return
	}

	c := newWSConn(s.nextConnID(), conn, s.ringSize, s.flushPeriod)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go c.writeLoop(ctx)

	s.sessions.Bind(principal, c)
	if s.presence != nil {
		_ = s.presence.Bind(ctx, principal, s.nodeID)
	}
	defer s.sessions.Unbind(principal, c.id)
	defer func() {
		if s.presence != nil {
			_ = s.presence.Unbind(context.Background(), principal, s.nodeID)
		}
		_ = c.Close()
	}()

	authReply, err := s.adapter.NewAuthReply(loginSeq, mustJSON(map[string]any{
		"type":    "login_ok",
		"user_id": principal.UserID,
		"domain":  principal.Domain,
		"scope":   principal.Scope,
	}), s.framePool.Get())
	if err == nil {
		_ = c.Send(ctx, authReply)
	}
	if offline, err := s.messaging.DrainOffline(ctx, principal); err == nil {
		if frame, ferr := s.adapter.NewPush(0, offline, s.framePool.Get()); ferr == nil {
			_ = c.Send(ctx, frame)
		}
	}

	conn.SetReadLimit(int64(goimx.HeaderSize + goimx.MaxBodySize))
	_ = conn.SetReadDeadline(time.Now().Add(s.heartbeat * time.Duration(s.missLimit)))

	for {
		_, wire, err := conn.ReadMessage()
		if err != nil {
			return
		}
		frame, err := s.adapter.Decode(bytes.NewReader(wire))
		if err != nil {
			_ = s.writeErrorReply(c, 0, err)
			continue
		}
		_ = conn.SetReadDeadline(time.Now().Add(s.heartbeat * time.Duration(s.missLimit)))
		switch frame.Op {
		case goimx.OpHeartbeat:
			reply, err := s.adapter.NewHeartbeatReply(frame.Seq, s.framePool.Get())
			if err == nil {
				_ = c.Send(ctx, reply)
			}
		case goimx.OpServerPush:
			reply, err := s.messaging.HandleCommand(ctx, principal, frame.Body)
			if err != nil {
				_ = s.writeErrorReply(c, frame.Seq, err)
				continue
			}
			wire, err := s.adapter.NewCommandReply(frame.Seq, reply, s.framePool.Get())
			if err == nil {
				_ = c.Send(ctx, wire)
			}
		default:
			_ = s.writeErrorReply(c, frame.Seq, errors.New("unsupported operation"))
		}
	}
}

func (s *Server) authenticateConn(conn *websocket.Conn) (auth.Principal, int32, error) {
	_, wire, err := conn.ReadMessage()
	if err != nil {
		return auth.Principal{}, 0, err
	}
	frame, err := s.adapter.Decode(bytes.NewReader(wire))
	if err != nil {
		return auth.Principal{}, 0, err
	}
	if frame.Op != goimx.OpAuth {
		return auth.Principal{}, frame.Seq, errors.New("auth frame required")
	}
	var login loginRequest
	if err := json.Unmarshal(frame.Body, &login); err != nil {
		return auth.Principal{}, frame.Seq, errors.New("invalid login payload")
	}
	scope, err := s.scope.Resolve(domainFromString(login.Domain), domainScope(login))
	if err != nil {
		return auth.Principal{}, frame.Seq, err
	}
	principal, err := s.auth.Authenticate(login.Token, domainFromString(login.Domain), scope)
	return principal, frame.Seq, err
}

func (s *Server) writeErrorFrame(conn *websocket.Conn, seq int32, message string) error {
	wire, err := s.adapter.NewError(seq, message, s.framePool.Get())
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.BinaryMessage, wire)
}

func (s *Server) writeErrorReply(conn *wsConn, seq int32, err error) error {
	wire, ferr := s.adapter.NewError(seq, err.Error(), s.framePool.Get())
	if ferr != nil {
		return ferr
	}
	return conn.Send(context.Background(), wire)
}

func (s *Server) nextConnID() string {
	return "ws-" + itoa(s.nextID.Add(1))
}

type wsConn struct {
	id          string
	conn        *websocket.Conn
	ring        *session.Ring
	mu          sync.Mutex
	signal      chan struct{}
	closed      atomic.Bool
	flushPeriod time.Duration
}

func newWSConn(id string, conn *websocket.Conn, ringSize int, flushPeriod time.Duration) *wsConn {
	return &wsConn{
		id:          id,
		conn:        conn,
		ring:        session.NewRing(ringSize),
		signal:      make(chan struct{}, 1),
		flushPeriod: flushPeriod,
	}
}

func (c *wsConn) ID() string {
	return c.id
}

func (c *wsConn) Send(_ context.Context, payload []byte) error {
	if c.closed.Load() {
		return errors.New("connection closed")
	}
	cp := append([]byte(nil), payload...)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ring.Push(cp); err != nil {
		return err
	}
	select {
	case c.signal <- struct{}{}:
	default:
	}
	return nil
}

func (c *wsConn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	select {
	case c.signal <- struct{}{}:
	default:
	}
	return c.conn.Close()
}

func (c *wsConn) writeLoop(ctx context.Context) {
	ticker := time.NewTicker(c.flushPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.signal:
		case <-ticker.C:
		}
		if c.closed.Load() {
			return
		}
		for {
			c.mu.Lock()
			payload, err := c.ring.Pop()
			c.mu.Unlock()
			if err != nil {
				break
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				_ = c.Close()
				return
			}
		}
	}
}

func domainFromString(v string) domain.IMDomain {
	return domain.IMDomain(v)
}

func domainScope(login loginRequest) domain.Scope {
	return domain.Scope{
		TenantID:    login.Scope.TenantID,
		ProjectID:   login.Scope.ProjectID,
		Environment: login.Scope.Environment,
	}
}

func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(errors.New("marshal response: " + err.Error()))
	}
	return data
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
