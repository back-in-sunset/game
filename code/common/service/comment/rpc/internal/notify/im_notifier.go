package notify

import (
	"context"
	"fmt"

	"im/rpc/imclient"
)

type ReplyNotice struct {
	Domain      string
	TenantID    string
	ProjectID   string
	Environment string
	ObjID       int64
	ObjType     int64
	CommentID   int64
	ReplyID     int64
	SenderID    int64
	ReceiverID  int64
	Message     string
}

type CommentNotifier interface {
	NotifyReply(ctx context.Context, notice ReplyNotice) error
}

type NoopCommentNotifier struct{}

func (NoopCommentNotifier) NotifyReply(context.Context, ReplyNotice) error { return nil }

type IMCommentNotifier struct {
	facade *imclient.Facade
}

func NewIMCommentNotifier(client imclient.IM) *IMCommentNotifier {
	return &IMCommentNotifier{facade: imclient.NewFacade(client)}
}

func (n *IMCommentNotifier) NotifyReply(ctx context.Context, notice ReplyNotice) error {
	if n == nil || n.facade == nil {
		return nil
	}
	if notice.ReceiverID <= 0 || notice.ReceiverID == notice.SenderID {
		return nil
	}
	_, err := n.facade.SendSystemNotice(ctx, imclient.SystemNotice{
		Scope: imclient.ScopeRef{
			Domain:      notice.Domain,
			TenantID:    notice.TenantID,
			ProjectID:   notice.ProjectID,
			Environment: notice.Environment,
		},
		Receiver: notice.ReceiverID,
		Seq:      notice.CommentID,
		Payload: map[string]any{
			"type":       "comment_reply",
			"obj_id":     notice.ObjID,
			"obj_type":   notice.ObjType,
			"comment_id": notice.CommentID,
			"reply_id":   notice.ReplyID,
			"sender_id":  notice.SenderID,
			"message":    notice.Message,
		},
	})
	if err != nil {
		return fmt.Errorf("send reply notice: %w", err)
	}
	return nil
}
