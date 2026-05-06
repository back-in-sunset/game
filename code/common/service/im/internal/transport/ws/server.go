package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

	authBody, err := json.Marshal(map[string]any{
		"type":    "login_ok",
		"user_id": principal.UserID,
		"domain":  principal.Domain,
		"scope":   principal.Scope,
	})
	if err == nil {
		s.encodeAndSend(ctx, c, goimx.OpAuthReply, loginSeq, authBody)
	}
	if offline, err := s.messaging.DrainOffline(ctx, principal); err == nil {
		s.encodeAndSend(ctx, c, goimx.OpServerPush, 0, offline)
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
			s.encodeAndSend(ctx, c, goimx.OpHeartbeatReply, frame.Seq, []byte(time.Now().UTC().Format(time.RFC3339)))
		case goimx.OpServerPush:
			reply, err := s.messaging.HandleCommand(ctx, principal, frame.Body)
			if err != nil {
				_ = s.writeErrorReply(c, frame.Seq, err)
				continue
			}
			s.encodeAndSend(ctx, c, goimx.OpCommandReply, frame.Seq, reply)
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

func (s *Server) encodeAndSend(ctx context.Context, c *wsConn, op int32, seq int32, body []byte) {
	buf := s.framePool.Get()
	wire, err := s.adapter.Encode(op, seq, body, buf)
	if err == nil {
		_ = c.Send(ctx, wire)
	}
	s.framePool.Put(buf)
}

func (s *Server) writeErrorFrame(conn *websocket.Conn, seq int32, message string) error {
	buf := s.framePool.Get()
	defer s.framePool.Put(buf)
	wire, err := s.adapter.NewError(seq, message, buf)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.BinaryMessage, wire)
}

func (s *Server) writeErrorReply(conn *wsConn, seq int32, err error) error {
	buf := s.framePool.Get()
	defer s.framePool.Put(buf)
	wire, ferr := s.adapter.NewError(seq, err.Error(), buf)
	if ferr != nil {
		return ferr
	}
	return conn.Send(context.Background(), wire)
}

func (s *Server) nextConnID() string {
	return "ws-" + strconv.FormatInt(s.nextID.Add(1), 10)
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
