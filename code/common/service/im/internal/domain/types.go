package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type IMDomain string

const (
	DomainPlatform IMDomain = "platform"
	DomainTenant   IMDomain = "tenant"
)

var validMessageTypes = []string{
	"direct_message",
	"system_notice",
	"biz_push",
}

type Scope struct {
	TenantID    string `json:"tenant_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Environment string `json:"environment,omitempty"`
}

func (s Scope) Normalize() Scope {
	s.TenantID = strings.TrimSpace(s.TenantID)
	s.ProjectID = strings.TrimSpace(s.ProjectID)
	s.Environment = strings.TrimSpace(s.Environment)
	return s
}

func (s Scope) Validate(domain IMDomain) error {
	s = s.Normalize()
	switch domain {
	case DomainPlatform:
		return nil
	case DomainTenant:
		if s.TenantID == "" || s.ProjectID == "" || s.Environment == "" {
			return fmt.Errorf("tenant scope requires tenant_id, project_id and environment")
		}
		return nil
	default:
		return fmt.Errorf("unknown domain %q", domain)
	}
}

type Envelope struct {
	Domain   IMDomain       `json:"domain"`
	Scope    Scope          `json:"scope"`
	Sender   int64          `json:"sender"`
	Receiver int64          `json:"receiver"`
	MsgType  string         `json:"msg_type"`
	Seq      int64          `json:"seq"`
	Payload  map[string]any `json:"payload"`
	SentAt   time.Time      `json:"sent_at"`
}

type Conversation struct {
	Key         string    `json:"key"`
	Domain      IMDomain  `json:"domain"`
	Scope       Scope     `json:"scope"`
	PeerUserID  int64     `json:"peer_user_id"`
	LastMessage Envelope  `json:"last_message"`
	UnreadCount int64     `json:"unread_count"`
	UpdatedAt   time.Time `json:"updated_at"`
	ReadSeq     int64     `json:"read_seq"`
}

func (e Envelope) Validate() error {
	if err := e.Scope.Validate(e.Domain); err != nil {
		return err
	}
	if e.Sender <= 0 && e.MsgType == "direct_message" {
		return fmt.Errorf("sender is required")
	}
	if e.Receiver <= 0 {
		return fmt.Errorf("receiver is required")
	}
	if !slices.Contains(validMessageTypes, e.MsgType) {
		return fmt.Errorf("unsupported msg_type %q", e.MsgType)
	}
	return nil
}

func ConversationKey(domain IMDomain, scope Scope, uidA, uidB int64) (string, error) {
	if err := scope.Validate(domain); err != nil {
		return "", err
	}
	left, right := uidA, uidB
	if left > right {
		left, right = right, left
	}
	switch domain {
	case DomainPlatform:
		return fmt.Sprintf("platform:%d:%d", left, right), nil
	case DomainTenant:
		scope = scope.Normalize()
		return fmt.Sprintf("tenant:%s:%s:%s:%d:%d", scope.TenantID, scope.ProjectID, scope.Environment, left, right), nil
	default:
		return "", fmt.Errorf("unknown domain %q", domain)
	}
}
