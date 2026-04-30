package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"im/internal/auth"
	"im/internal/config"
	"im/internal/domain"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type mysqlArchive struct {
	conn sqlx.SqlConn
}

type redisMySQLStateStore struct {
	conn      sqlx.SqlConn
	redis     *redis.Client
	keyPrefix string
}

type mysqlConversationRow struct {
	PrincipalKey    string    `db:"principal_key"`
	ConversationKey string    `db:"conversation_key"`
	Domain          string    `db:"domain"`
	TenantID        string    `db:"tenant_id"`
	ProjectID       string    `db:"project_id"`
	Environment     string    `db:"environment"`
	OwnerUserID     int64     `db:"owner_user_id"`
	PeerUserID      int64     `db:"peer_user_id"`
	LastSeq         int64     `db:"last_seq"`
	LastMessageJSON string    `db:"last_message_json"`
	UnreadCount     int64     `db:"unread_count"`
	ReadSeq         int64     `db:"read_seq"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type mysqlMessageRow struct {
	ID              int64     `db:"id"`
	Domain          string    `db:"domain"`
	TenantID        string    `db:"tenant_id"`
	ProjectID       string    `db:"project_id"`
	Environment     string    `db:"environment"`
	ConversationKey string    `db:"conversation_key"`
	Sender          int64     `db:"sender"`
	Receiver        int64     `db:"receiver"`
	MsgType         string    `db:"msg_type"`
	Seq             int64     `db:"seq"`
	PayloadJSON     string    `db:"payload_json"`
	SentAt          time.Time `db:"sent_at"`
}

func NewMySQLRedisStore(cfg config.Config) (*CompositeStore, error) {
	conn := sqlx.NewMysql(cfg.Mysql.DataSource)
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	return NewCompositeStore(
		&mysqlArchive{conn: conn},
		&redisMySQLStateStore{conn: conn, redis: rdb, keyPrefix: cfg.Redis.KeyPrefix},
	), nil
}

func (m *mysqlArchive) AppendMessage(ctx context.Context, envelope domain.Envelope) (StoredMessage, error) {
	key, err := domain.ConversationKey(envelope.Domain, envelope.Scope, envelope.Sender, envelope.Receiver)
	if err != nil {
		return StoredMessage{}, err
	}
	payload, err := json.Marshal(envelope.Payload)
	if err != nil {
		return StoredMessage{}, err
	}
	scope := envelope.Scope.Normalize()
	res, err := m.conn.ExecCtx(
		ctx,
		"insert into im_message (domain,tenant_id,project_id,environment,conversation_key,sender,receiver,msg_type,seq,payload_json,sent_at) values (?,?,?,?,?,?,?,?,?,?,?)",
		string(envelope.Domain), scope.TenantID, scope.ProjectID, scope.Environment, key,
		envelope.Sender, envelope.Receiver, envelope.MsgType, envelope.Seq, string(payload), envelope.SentAt,
	)
	if err != nil {
		return StoredMessage{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return StoredMessage{}, err
	}
	return StoredMessage{ID: id, ConversationKey: key, Envelope: envelope}, nil
}

func (m *mysqlArchive) ListMessages(ctx context.Context, principal auth.Principal, peerUserID int64, limit int) ([]domain.Envelope, error) {
	key, err := domain.ConversationKey(principal.Domain, principal.Scope, principal.UserID, peerUserID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []mysqlMessageRow
	if err := m.conn.QueryRowsCtx(ctx, &rows,
		"select id,domain,tenant_id,project_id,environment,conversation_key,sender,receiver,msg_type,seq,payload_json,sent_at from im_message where conversation_key=? order by sent_at desc,id desc limit ?",
		key, limit,
	); err != nil {
		return nil, err
	}
	out := make([]domain.Envelope, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		payload := make(map[string]any)
		if err := json.Unmarshal([]byte(rows[i].PayloadJSON), &payload); err != nil {
			return nil, err
		}
		out = append(out, domain.Envelope{
			Domain:   domain.IMDomain(rows[i].Domain),
			Scope:    domain.Scope{TenantID: rows[i].TenantID, ProjectID: rows[i].ProjectID, Environment: rows[i].Environment},
			Sender:   rows[i].Sender,
			Receiver: rows[i].Receiver,
			MsgType:  rows[i].MsgType,
			Seq:      rows[i].Seq,
			Payload:  payload,
			SentAt:   rows[i].SentAt,
		})
	}
	return out, nil
}

func (s *redisMySQLStateStore) UpsertConversationPair(ctx context.Context, stored StoredMessage) error {
	return s.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := s.upsertConversation(ctx, session, stored, stored.Envelope.Sender, stored.Envelope.Receiver, 0); err != nil {
			return err
		}
		return s.upsertConversation(ctx, session, stored, stored.Envelope.Receiver, stored.Envelope.Sender, 1)
	})
}

func (s *redisMySQLStateStore) SaveOffline(ctx context.Context, principal auth.Principal, envelope domain.Envelope) error {
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return s.redis.RPush(ctx, s.offlineKey(principal), data).Err()
}

func (s *redisMySQLStateStore) DrainOffline(ctx context.Context, principal auth.Principal) ([]domain.Envelope, error) {
	key := s.offlineKey(principal)
	items, err := s.redis.LRange(ctx, key, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return nil, err
	}
	out := make([]domain.Envelope, 0, len(items))
	for _, item := range items {
		var envelope domain.Envelope
		if err := json.Unmarshal([]byte(item), &envelope); err != nil {
			return nil, err
		}
		out = append(out, envelope)
	}
	return out, nil
}

func (s *redisMySQLStateStore) ListConversations(ctx context.Context, principal auth.Principal) ([]domain.Conversation, error) {
	rows := make([]mysqlConversationRow, 0)
	if err := s.conn.QueryRowsCtx(ctx, &rows,
		"select principal_key,conversation_key,domain,tenant_id,project_id,environment,owner_user_id,peer_user_id,last_seq,last_message_json,unread_count,read_seq,updated_at from im_conversation where principal_key=? order by updated_at desc,id desc",
		principalKey(principal),
	); err != nil {
		return nil, err
	}
	out := make([]domain.Conversation, 0, len(rows))
	for _, row := range rows {
		lastMessage, err := decodeEnvelopeJSON(row.LastMessageJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.Conversation{
			Key:         row.ConversationKey,
			Domain:      domain.IMDomain(row.Domain),
			Scope:       domain.Scope{TenantID: row.TenantID, ProjectID: row.ProjectID, Environment: row.Environment},
			PeerUserID:  row.PeerUserID,
			LastMessage: lastMessage,
			UnreadCount: row.UnreadCount,
			UpdatedAt:   row.UpdatedAt,
			ReadSeq:     row.ReadSeq,
		})
	}
	return out, nil
}

func (s *redisMySQLStateStore) MarkRead(ctx context.Context, principal auth.Principal, peerUserID int64, seq int64) error {
	key, err := domain.ConversationKey(principal.Domain, principal.Scope, principal.UserID, peerUserID)
	if err != nil {
		return err
	}
	_, err = s.conn.ExecCtx(ctx,
		"update im_conversation set read_seq=greatest(read_seq, ?), unread_count=case when last_seq <= ? then 0 else greatest(last_seq - ?, 0) end where principal_key=? and conversation_key=?",
		seq, seq, seq, principalKey(principal), key,
	)
	return err
}

func (s *redisMySQLStateStore) upsertConversation(ctx context.Context, session sqlx.Session, stored StoredMessage, ownerUserID, peerUserID int64, unreadDelta int64) error {
	payload, err := json.Marshal(stored.Envelope)
	if err != nil {
		return err
	}
	scope := stored.Envelope.Scope.Normalize()
	_, err = session.ExecCtx(ctx,
		"insert into im_conversation (principal_key,conversation_key,domain,tenant_id,project_id,environment,owner_user_id,peer_user_id,last_message_id,last_seq,last_message_json,unread_count,read_seq,updated_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?) "+
			"on duplicate key update last_message_id=values(last_message_id),last_seq=values(last_seq),last_message_json=values(last_message_json),updated_at=values(updated_at),unread_count=case when values(unread_count)=0 then unread_count else unread_count + values(unread_count) end",
		principalKey(auth.Principal{UserID: ownerUserID, Domain: stored.Envelope.Domain, Scope: scope}),
		stored.ConversationKey,
		string(stored.Envelope.Domain),
		scope.TenantID,
		scope.ProjectID,
		scope.Environment,
		ownerUserID,
		peerUserID,
		stored.ID,
		stored.Envelope.Seq,
		string(payload),
		unreadDelta,
		0,
		stored.Envelope.SentAt,
	)
	return err
}

func (s *redisMySQLStateStore) offlineKey(principal auth.Principal) string {
	return fmt.Sprintf("%s:offline:%s", s.keyPrefix, principalKey(principal))
}

func (s *redisMySQLStateStore) Close() error {
	return s.redis.Close()
}

func decodeEnvelopeJSON(raw string) (domain.Envelope, error) {
	var envelope domain.Envelope
	err := json.Unmarshal([]byte(raw), &envelope)
	return envelope, err
}
