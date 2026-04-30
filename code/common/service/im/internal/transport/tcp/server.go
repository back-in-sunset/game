package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"sync"
	"sync/atomic"

	"im/internal/auth"
	"im/internal/contracts"
	"im/internal/domain"
	"im/internal/pipeline"
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
	listener net.Listener
	closed   chan struct{}
	nextID   atomic.Int64
	wg       sync.WaitGroup
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
) *Server {
	return &Server{
		addr:      addr,
		nodeID:    nodeID,
		auth:      authProvider,
		scope:     scopeResolver,
		sessions:  sessions,
		messaging: messaging,
		presence:  presence,
		closed:    make(chan struct{}),
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

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	var login loginRequest
	if err := json.Unmarshal(line, &login); err != nil {
		_, _ = conn.Write([]byte("{\"error\":\"invalid login payload\"}\n"))
		return
	}

	scope, err := s.scope.Resolve(domain.IMDomain(login.Domain), domainScope(login))
	if err != nil {
		_, _ = conn.Write([]byte("{\"error\":\"" + err.Error() + "\"}\n"))
		return
	}

	principal, err := s.auth.Authenticate(login.Token, domain.IMDomain(login.Domain), scope)
	if err != nil {
		_, _ = conn.Write([]byte("{\"error\":\"" + err.Error() + "\"}\n"))
		return
	}

	c := &tcpConn{id: s.nextConnID(), conn: conn}
	s.sessions.Bind(principal, c)
	if s.presence != nil {
		_ = s.presence.Bind(context.Background(), principal, s.nodeID)
	}
	defer s.sessions.Unbind(principal, c.id)
	defer func() {
		if s.presence != nil {
			_ = s.presence.Unbind(context.Background(), principal, s.nodeID)
		}
	}()

	_, _ = conn.Write([]byte("{\"type\":\"login_ok\"}\n"))
	if offline, err := s.messaging.DrainOffline(context.Background(), principal); err == nil {
		_ = c.Send(context.Background(), append(offline, '\n'))
	}

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		reply, err := s.messaging.HandleCommand(context.Background(), principal, data)
		if err != nil {
			_, _ = conn.Write([]byte("{\"type\":\"error\",\"error\":\"" + err.Error() + "\"}\n"))
			continue
		}
		_ = c.Send(context.Background(), append(reply, '\n'))
	}
}

func (s *Server) nextConnID() string {
	return "tcp-" + itoa(s.nextID.Add(1))
}

type tcpConn struct {
	id   string
	conn net.Conn
}

func (c *tcpConn) ID() string {
	return c.id
}

func (c *tcpConn) Send(_ context.Context, payload []byte) error {
	_, err := c.conn.Write(payload)
	return err
}

func (c *tcpConn) Close() error {
	return c.conn.Close()
}

func domainScope(login loginRequest) domain.Scope {
	return domain.Scope{
		TenantID:    login.Scope.TenantID,
		ProjectID:   login.Scope.ProjectID,
		Environment: login.Scope.Environment,
	}
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
