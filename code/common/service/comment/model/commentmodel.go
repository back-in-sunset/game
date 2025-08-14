package model

import (
	"comment/rpc/comment"
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	// CommentModel 评论模型
	CommentModel interface {
		AddComment(ctx context.Context, data *CommentSubject, ci *CommentIndex, cc *CommentContent) (*comment.CommentResponse, error)
		CommentByObjID(ctx context.Context, objID uint64, objType uint64, sortField string, limit uint64) ([]*Comment, error)
		FindOneByObjID(ctx context.Context, objID uint64, id uint64) (*Comment, error)
	}

	customCommentModel struct {
		*customCommentContentModel
		*customCommentIndexModel
		*customCommentSubjectModel
		tableFn func(uint64) string
	}
)

// Comment 评论
type Comment struct {
	ID          int64     `db:"id"`
	ObjID       uint64    `db:"obj_id"`        // 评论对象ID使用唯一id的话不用type联合主键
	ObjType     uint64    `db:"obj_type"`      // 评论对象类型
	MemberID    uint64    `db:"member_id"`     // 作者用户ID
	RootID      uint64    `db:"root_id"`       // 根评论ID 不为0表示是回复评论
	ReplyID     uint64    `db:"reply_id"`      // 被回复的评论ID
	Floor       uint64    `db:"floor"`         // 评论楼层
	Count       int64     `db:"count"`         // 挂载子评论总数 可见
	RootCount   int64     `db:"root_count"`    // 挂载子评论总数 所以
	LikeCount   int64     `db:"like_count"`    // 点赞数
	HateCount   int64     `db:"hate_count"`    // 点踩数
	State       uint64    `db:"state"`         // 0-正常, 1-隐藏
	Attrs       int64     `db:"attrs"`         // 属性(bit 0-运营置顶, 1-owner置顶 2-大数据)
	AtMemberIDs string    `db:"at_member_ids"` // at用户ID列表
	IP          string    `db:"ip"`            // 评论IP
	Platform    uint64    `db:"platform"`      // 评论平台
	Device      string    `db:"device"`        // 评论设备
	Message     string    `db:"message"`       // 评论内容
	Meta        string    `db:"meta"`          // 评论元数据 背景 字体
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// NewCommentModel 创建评论模型
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

// AddComment 添加评论
func (m *customCommentModel) AddComment(ctx context.Context, data *CommentSubject, ci *CommentIndex, cc *CommentContent) (*comment.CommentResponse, error) {
	m.defaultCommentSubjectModel.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		cj, err := m.customCommentSubjectModel.FindOneByStateAttrsObjIDObjType(ctx, 0, data.ObjId, data.ObjType)
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
			// 缓存更新
			err = m.customCommentSubjectModel.Update(ctx, &CommentSubject{
				ID:        cj.ID,
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
			oci, err := m.customCommentIndexModel.FindOne(ctx, ci.ObjId, ci.RootId)
			if err != nil {
				return err
			}
			// 缓存更新
			err = m.customCommentIndexModel.Update(ctx, &CommentIndex{
				ID:        ci.RootId,
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
		commentID, _ := cires.LastInsertId()
		ci.ID = uint64(commentID)
		return nil
	})

	cc.CommentId = uint64(ci.ID)
	_, err := m.customCommentContentModel.Insert(ctx, cc)
	if err != nil {
		return nil, err
	}
	return &comment.CommentResponse{
		CommentID: cc.CommentId,
	}, nil
}

// FindOneByObjID 查询评论
func (m *customCommentModel) FindOneByObjID(ctx context.Context, objID uint64, id uint64) (*Comment, error) {
	oci, err := m.customCommentIndexModel.FindOne(ctx, objID, id)
	if err != nil {
		return nil, err
	}
	occ, err := m.customCommentContentModel.FindOne(ctx, objID, id)
	if err != nil {
		return nil, err
	}
	return &Comment{
		ID:          int64(oci.ID),
		ObjID:       oci.ObjId,
		ObjType:     oci.ObjType,
		MemberID:    oci.MemberId,
		RootID:      oci.RootId,
		ReplyID:     oci.ReplyId,
		Floor:       oci.Floor,
		Count:       oci.Count,
		RootCount:   oci.RootCount,
		LikeCount:   oci.LikeCount,
		HateCount:   oci.HateCount,
		State:       oci.State,
		Attrs:       oci.Attrs,
		AtMemberIDs: occ.AtMemberIds,
		IP:          occ.Ip,
		Platform:    occ.Platform,
		Device:      occ.Device,
		Message:     occ.Message,
		Meta:        occ.Meta,
		CreatedAt:   oci.CreatedAt,
	}, nil

}

// CommentByObjID 查询评论
func (m *customCommentModel) CommentByObjID(ctx context.Context, objID uint64, objType uint64, sortField string, limit uint64) ([]*Comment, error) {
	var (
		err error
		sql string
		// anyField any
		commentIndexs []*CommentIndex
	)

	if sortField == "like_count" {
		// anyField := sortLikeCount
		sql = fmt.Sprintf("select " + commentIndexRows + " from " + m.customCommentIndexModel.tableFn(objID) +
			" where obj_id=? and obj_type=? order by like_count desc limit ?")
	} else {
		// anyField := sortPublishTime
		sql = fmt.Sprintf("select " + commentIndexRows + " from " + m.customCommentIndexModel.tableFn(objID) +
			" where obj_id=? and obj_type=? order by publish_time desc limit ?")
	}

	// sql = fmt.Sprintf("select "+commentIndexRows+" from "+m.customCommentIndexModel.tableFn(objID)+
	// 	" where obj_id=? and obj_type=? order by %s desc limit ?", sortField)

	err = m.defaultCommentIndexModel.QueryRowsNoCacheCtx(ctx, &commentIndexs, sql, objID, objType, limit)
	if err != nil {
		logx.Error("QueryRowNoCacheCtx commentIndexs err:", err)
		return nil, err
	}

	var comments = make([]*Comment, 0, len(commentIndexs))
	for _, commentIndex := range commentIndexs {
		comment, err := m.FindOneByObjID(ctx, objID, commentIndex.ID)
		if err != nil {
			logx.Error("FindOneByObjId err:", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
