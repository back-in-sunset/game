package logic

import (
	"comment/rpc/comment"
	"comment/rpc/model"
)

func toCommentResponse(c *model.Comment) *comment.CommentResponse {
	if c == nil {
		return &comment.CommentResponse{}
	}

	return &comment.CommentResponse{
		ID:          c.ID,
		ObjID:       c.ObjID,
		ObjType:     c.ObjType,
		MemberID:    c.MemberID,
		CommentID:   c.ID,
		AtMemberIDs: c.AtMemberIDs,
		Ip:          c.IP,
		Platform:    c.Platform,
		Device:      c.Device,
		Message:     c.Message,
		Meta:        c.Meta,
		ReplyID:     c.ReplyID,
		State:       c.State,
		RootID:      c.RootID,
		CreatedAt:   c.CreatedAt.Unix(),
		Floor:       c.Floor,
		LikeCount:   c.LikeCount,
		HateCount:   c.HateCount,
		Count:       c.Count,
	}
}
