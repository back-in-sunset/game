package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"

	"im/internal/auth"
	"im/internal/contracts"
	"im/internal/domain"
	"im/internal/pipeline"

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
	server *http.Server
	nextID atomic.Int64
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
	s := &Server{
		addr:      addr,
		nodeID:    nodeID,
		auth:      authProvider,
		scope:     scopeResolver,
		sessions:  sessions,
		messaging: messaging,
		presence:  presence,
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

	var login loginRequest
	if err := conn.ReadJSON(&login); err != nil {
		_ = conn.WriteJSON(map[string]any{"error": "invalid login payload"})
		_ = conn.Close()
		return
	}

	scope, err := s.scope.Resolve(domainFromString(login.Domain), domainScope(login))
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"error": err.Error()})
		_ = conn.Close()
		return
	}

	principal, err := s.auth.Authenticate(login.Token, domainFromString(login.Domain), scope)
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"error": err.Error()})
		_ = conn.Close()
		return
	}

	c := &wsConn{id: s.nextConnID(), conn: conn}
	s.sessions.Bind(principal, c)
	if s.presence != nil {
		_ = s.presence.Bind(r.Context(), principal, s.nodeID)
	}
	defer s.sessions.Unbind(principal, c.id)
	defer func() {
		if s.presence != nil {
			_ = s.presence.Unbind(context.Background(), principal, s.nodeID)
		}
	}()
	defer c.Close()

	_ = conn.WriteJSON(map[string]any{
		"type":    "login_ok",
		"user_id": principal.UserID,
		"domain":  principal.Domain,
		"scope":   principal.Scope,
	})
	if offline, err := s.messaging.DrainOffline(r.Context(), principal); err == nil {
		_ = c.Send(r.Context(), offline)
	}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		reply, err := s.messaging.HandleCommand(r.Context(), principal, data)
		if err != nil {
			_ = c.Send(r.Context(), mustJSON(map[string]any{"type": "error", "error": err.Error()}))
			continue
		}
		_ = c.Send(r.Context(), reply)
	}
}

func (s *Server) nextConnID() string {
	return "ws-" + itoa(s.nextID.Add(1))
}

type wsConn struct {
	id   string
	conn *websocket.Conn
}

func (c *wsConn) ID() string {
	return c.id
}

func (c *wsConn) Send(_ context.Context, payload []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func (c *wsConn) Close() error {
	return c.conn.Close()
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
