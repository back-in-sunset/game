package model

import (
	"comment/rpc/pb/comment"
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	CommentModel interface {
		AddComment(ctx context.Context, data *CommentSubject, ci *CommentIndex, cc *CommentContent) (*comment.CommentResponse, error)
		CommentByObjId(ctx context.Context, objId uint64, objType uint64, sortLikeCount uint64, sortPublishTime string, sortField string, limit uint64) (*comment.CommentListResponse, error)
		FindOneByObjId(ctx context.Context, objId uint64, id uint64) (*Comment, error)
	}

	customCommentModel struct {
		*customCommentContentModel
		*customCommentIndexModel
		*customCommentSubjectModel
		tableFn func(uint64) string
	}
)

type Comment struct {
	Id          int64     `db:"id"`
	ObjId       uint64    `db:"obj_id"`        // 评论对象ID使用唯一id的话不用type联合主键
	ObjType     uint64    `db:"obj_type"`      // 评论对象类型
	MemberId    uint64    `db:"member_id"`     // 作者用户ID
	RootId      uint64    `db:"root_id"`       // 根评论ID 不为0表示是回复评论
	ReplyId     uint64    `db:"reply_id"`      // 被回复的评论ID
	Floor       uint64    `db:"floor"`         // 评论楼层
	Count       int64     `db:"count"`         // 挂载子评论总数 可见
	RootCount   int64     `db:"root_count"`    // 挂载子评论总数 所以
	LikeCount   int64     `db:"like_count"`    // 点赞数
	HateCount   int64     `db:"hate_count"`    // 点踩数
	State       uint64    `db:"state"`         // 0-正常, 1-隐藏
	Attrs       int64     `db:"attrs"`         // 属性(bit 0-运营置顶, 1-owner置顶 2-大数据)
	AtMemberIds string    `db:"at_member_ids"` // at用户ID列表
	Ip          string    `db:"ip"`            // 评论IP
	Platform    uint64    `db:"platform"`      // 评论平台
	Device      string    `db:"device"`        // 评论设备
	Message     string    `db:"message"`       // 评论内容
	Meta        string    `db:"meta"`          // 评论元数据 背景 字体
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewCommentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentModel {
	return &customCommentModel{
		customCommentContentModel: newCustomCommentContentModel(conn, c, opts...),
		customCommentIndexModel:   newCustomCommentIndexModel(conn, c, opts...),
		customCommentSubjectModel: newCustomCommentSubjectModel(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
		},
	}
}

func (m *customCommentModel) AddComment(ctx context.Context, data *CommentSubject, ci *CommentIndex, cc *CommentContent) (*comment.CommentResponse, error) {
	ctx = context.WithValue(ctx, "objId", data.ObjId)
	m.defaultCommentSubjectModel.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		cj, err := m.customCommentSubjectModel.FindOneByStateAttrsObjIdObjType(ctx, 0, data.ObjId, data.ObjType)
		if err != nil && err != ErrNotFound {
			return err
		}
		if err == ErrNotFound {
			_, err := m.customCommentSubjectModel.Insert(ctx, &CommentSubject{
				ObjId:     data.ObjId,
				ObjType:   data.ObjType,
				MemberId:  data.MemberId,
				Count:     1,
				RootCount: 1,
				AllCount:  1,
				State:     data.State,
				Attrs:     data.Attrs,
			})
			if err != nil {
				return err
			}
		} else {
			err = m.customCommentSubjectModel.Update(ctx, &CommentSubject{
				Id:        cj.Id,
				ObjId:     cj.ObjId,
				ObjType:   cj.ObjType,
				Count:     cj.Count + 1,
				RootCount: cj.RootCount + 1,
				AllCount:  cj.AllCount + 1,
			})
			if err != nil {
				return err
			}
		}

		if ci.RootId > 0 {
			oci, err := m.customCommentIndexModel.FindOne(ctx, ci.RootId)
			if err != nil {
				return err
			}

			err = m.customCommentIndexModel.Update(ctx, &CommentIndex{
				Id:        ci.RootId,
				ObjId:     data.ObjId,
				ObjType:   data.ObjType,
				MemberId:  data.MemberId,
				RootId:    ci.RootId,
				ReplyId:   ci.ReplyId,
				Floor:     oci.Floor + 1,
				Count:     oci.Count + 1,
				RootCount: oci.RootCount + 1,
				LikeCount: oci.LikeCount,
				HateCount: oci.HateCount,
			})
			if err != nil {
				return err
			}
		}

		cires, err := m.customCommentIndexModel.Insert(ctx, &CommentIndex{
			ObjId:     data.ObjId,
			ObjType:   data.ObjType,
			MemberId:  data.MemberId,
			RootId:    ci.RootId,
			ReplyId:   ci.ReplyId,
			Floor:     1,
			Count:     1,
			RootCount: 1,
			LikeCount: 0,
			HateCount: 0,
			State:     ci.State,
			Attrs:     ci.Attrs,
		})
		if err != nil {
			return err
		}
		commentId, _ := cires.LastInsertId()
		ci.Id = uint64(commentId)
		return nil
	})

	cc.CommentId = uint64(ci.Id)
	_, err := m.customCommentContentModel.Insert(ctx, cc)
	if err != nil {
		return nil, err
	}
	return &comment.CommentResponse{
		CommentId: cc.CommentId,
	}, nil
}

func (m *customCommentModel) FindOneByObjId(ctx context.Context, objId uint64, id uint64) (*Comment, error) {
	ctx = context.WithValue(ctx, "objId", objId)
	oci, err := m.customCommentIndexModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	occ, err := m.customCommentContentModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Comment{
		Id:          int64(oci.Id),
		ObjId:       oci.ObjId,
		ObjType:     oci.ObjType,
		MemberId:    oci.MemberId,
		RootId:      oci.RootId,
		ReplyId:     oci.ReplyId,
		Floor:       oci.Floor,
		Count:       oci.Count,
		RootCount:   oci.RootCount,
		LikeCount:   oci.LikeCount,
		HateCount:   oci.HateCount,
		State:       oci.State,
		Attrs:       oci.Attrs,
		AtMemberIds: occ.AtMemberIds,
		Ip:          occ.Ip,
		Platform:    occ.Platform,
		Device:      occ.Device,
		Message:     occ.Message,
		Meta:        occ.Meta,
		CreatedAt:   oci.CreatedAt,
	}, nil

}

func (m *customCommentModel) CommentByObjId(ctx context.Context, objId uint64, objType uint64, sortLikeCount uint64, sortPublishTime string, sortField string, limit uint64) (*comment.CommentListResponse, error) {
	return nil, nil
}
