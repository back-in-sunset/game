package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
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
	listener    net.Listener
	closed      chan struct{}
	nextID      atomic.Int64
	wg          sync.WaitGroup
	adapter     *goimx.Adapter
	framePool   *mempool.BufferPool
	ringSize    int
	readerSize  int
	writerSize  int
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
	return &Server{
		addr:        addr,
		nodeID:      nodeID,
		auth:        authProvider,
		scope:       scopeResolver,
		sessions:    sessions,
		messaging:   messaging,
		presence:    presence,
		closed:      make(chan struct{}),
		adapter:     goimx.NewAdapter(),
		framePool:   mempool.NewBufferPool(cfg.FrameBufferSize),
		ringSize:    cfg.RingSize,
		readerSize:  cfg.ReaderBufferSize,
		writerSize:  cfg.WriterBufferSize,
		heartbeat:   time.Duration(cfg.HeartbeatInterval) * time.Second,
		missLimit:   cfg.HeartbeatMisses,
		flushPeriod: time.Duration(cfg.WriteFlushInterval) * time.Millisecond,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-s.closed:
					return
				default:
					return
				}
			}
			s.wg.Add(1)
			go func(c net.Conn) {
				defer s.wg.Done()
				s.handleConn(c)
			}(conn)
		}
	}()
	return nil
}

func (s *Server) Stop(_ context.Context) error {
	close(s.closed)
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.wg.Wait()
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReaderSize(conn, s.readerSize)
	principal, loginSeq, err := s.authenticateConn(reader)
	if err != nil {
		_ = s.writeFrame(conn, 0, err.Error())
		return
	}

	c := newTCPConn(s.nextConnID(), conn, s.ringSize, s.writerSize, s.flushPeriod)
	ctx, cancel := context.WithCancel(context.Background())
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

	_ = conn.SetReadDeadline(time.Now().Add(s.heartbeat * time.Duration(s.missLimit)))
	for {
		frame, err := s.adapter.Decode(reader)
		if err != nil {
			return
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

func (s *Server) authenticateConn(reader *bufio.Reader) (auth.Principal, int32, error) {
	frame, err := s.adapter.Decode(reader)
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
	scope, err := s.scope.Resolve(domain.IMDomain(login.Domain), domainScope(login))
	if err != nil {
		return auth.Principal{}, frame.Seq, err
	}
	principal, err := s.auth.Authenticate(login.Token, domain.IMDomain(login.Domain), scope)
	return principal, frame.Seq, err
}

func (s *Server) writeFrame(conn net.Conn, seq int32, message string) error {
	wire, err := s.adapter.NewError(seq, message, s.framePool.Get())
	if err != nil {
		return err
	}
	_, err = conn.Write(wire)
	return err
}

func (s *Server) writeErrorReply(conn *tcpConn, seq int32, err error) error {
	wire, ferr := s.adapter.NewError(seq, err.Error(), s.framePool.Get())
	if ferr != nil {
		return ferr
	}
	return conn.Send(context.Background(), wire)
}

func (s *Server) nextConnID() string {
	return "tcp-" + itoa(s.nextID.Add(1))
}

type tcpConn struct {
	id          string
	conn        net.Conn
	writer      *bufio.Writer
	ring        *session.Ring
	mu          sync.Mutex
	signal      chan struct{}
	closed      atomic.Bool
	flushPeriod time.Duration
}

func newTCPConn(id string, conn net.Conn, ringSize, writerSize int, flushPeriod time.Duration) *tcpConn {
	return &tcpConn{
		id:          id,
		conn:        conn,
		writer:      bufio.NewWriterSize(conn, writerSize),
		ring:        session.NewRing(ringSize),
		signal:      make(chan struct{}, 1),
		flushPeriod: flushPeriod,
	}
}

func (c *tcpConn) ID() string {
	return c.id
}

func (c *tcpConn) Send(_ context.Context, payload []byte) error {
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

func (c *tcpConn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	select {
	case c.signal <- struct{}{}:
	default:
	}
	return c.conn.Close()
}

func (c *tcpConn) writeLoop(ctx context.Context) {
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
		wrote := false
		for {
			c.mu.Lock()
			payload, err := c.ring.Pop()
			c.mu.Unlock()
			if err != nil {
				break
			}
			if _, err := c.writer.Write(payload); err != nil {
				_ = c.Close()
				return
			}
			wrote = true
		}
		if wrote {
			if err := c.writer.Flush(); err != nil {
				_ = c.Close()
				return
			}
		}
	}
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
