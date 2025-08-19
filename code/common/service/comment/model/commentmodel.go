package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	// CommentModel 评论模型
	CommentModel interface {
		AddComment(ctx context.Context, data *CommentSubject, ci *CommentIndex, cc *CommentContent) (*CommentSchema, error)
		CommentListByObjID(ctx context.Context, objID int64, objType int64, sortField string, limit int64) ([]*Comment, error)
		FindOneByObjID(ctx context.Context, objID int64, id int64) (*Comment, error)
		CacheCommentsByIDs(ctx context.Context, objID int64, ids []int64) ([]*Comment, error)
	}

	customCommentModel struct {
		sqlc.CachedConn
		// *customCommentContentModel
		// *customCommentIndexModel
		// *customCommentSubjectModel
		newCustomCommentSubjectModelFunc
		newCustomCommentIndexModelFunc
		newCustomCommentContentModelFunc
	}
)

// Comment 评论
type Comment struct {
	ID          int64     `db:"id"`
	ObjID       int64     `db:"obj_id"`        // 评论对象ID使用唯一id的话不用type联合主键
	ObjType     int64     `db:"obj_type"`      // 评论对象类型
	MemberID    int64     `db:"member_id"`     // 作者用户ID
	RootID      int64     `db:"root_id"`       // 根评论ID 不为0表示是回复评论
	ReplyID     int64     `db:"reply_id"`      // 被回复的评论ID
	Floor       int64     `db:"floor"`         // 评论楼层
	Count       int64     `db:"count"`         // 挂载子评论总数 可见
	RootCount   int64     `db:"root_count"`    // 挂载子评论总数 所以
	LikeCount   int64     `db:"like_count"`    // 点赞数
	HateCount   int64     `db:"hate_count"`    // 点踩数
	State       int64     `db:"state"`         // 0-正常, 1-隐藏
	Attrs       int64     `db:"attrs"`         // 属性(bit 0-运营置顶, 1-owner置顶 2-大数据)
	AtMemberIDs string    `db:"at_member_ids"` // at用户ID列表
	IP          string    `db:"ip"`            // 评论IP
	Platform    int64     `db:"platform"`      // 评论平台
	Device      string    `db:"device"`        // 评论设备
	Message     string    `db:"message"`       // 评论内容
	Meta        string    `db:"meta"`          // 评论元数据 背景 字体
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// NewCommentModel 创建评论模型
func NewCommentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentModel {
	return &customCommentModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		// customCommentContentModel: newCustomCommentContentModel(conn, c, opts...),
		// customCommentIndexModel:   newCustomCommentIndexModel(conn, c, opts...),
		// // customCommentSubjectModel: newCustomCommentSubjectModel(conn, c, opts...),
		// tableFn: func(shardingId int64) string {
		// 	// Use the last 8 bits of the shardingId for determining the table suffix.
		// 	//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
		// 	return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
		// },
		newCustomCommentSubjectModelFunc: func(shardingID int64) *customCommentSubjectModel {
			return newCustomCommentSubjectModel(conn, c, shardingID, opts...)
		},
		newCustomCommentIndexModelFunc: func(shardingID int64) *customCommentIndexModel {
			return newCustomCommentIndexModel(conn, c, shardingID, opts...)
		},
		newCustomCommentContentModelFunc: func(shardingID int64) *customCommentContentModel {
			return newCustomCommentContentModel(conn, c, shardingID, opts...)
		},
	}
}

// AddComment 添加评论
func (m *customCommentModel) AddComment(ctx context.Context, data *CommentSubject, ci *CommentIndex, cc *CommentContent) (*CommentSchema, error) {
	m.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		cj, err := m.newCustomCommentSubjectModelFunc(data.ObjId).FindOneByStateObjIdObjType(ctx, 0, data.ObjId, data.ObjType)
		if err != nil && err != ErrNotFound {
			return err
		}
		if err == ErrNotFound {
			_, err := m.newCustomCommentSubjectModelFunc(data.ObjId).Insert(ctx, &CommentSubject{
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
			err = m.newCustomCommentSubjectModelFunc(data.ObjId).Update(ctx, &CommentSubject{
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
			oci, err := m.newCustomCommentIndexModelFunc(data.ObjId).FindOne(ctx, ci.RootId)
			if err != nil {
				return err
			}
			// 缓存更新
			err = m.newCustomCommentIndexModelFunc(data.ObjId).Update(ctx, &CommentIndex{
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

		cires, err := m.newCustomCommentIndexModelFunc(data.ObjId).Insert(ctx, &CommentIndex{
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
		ci.ID = commentID
		return nil
	})

	cc.CommentId = ci.ID
	_, err := m.newCustomCommentContentModelFunc(data.ObjId).Insert(ctx, cc)
	if err != nil {
		return nil, err
	}
	return &CommentSchema{
		CommentID: cc.CommentId,
	}, nil
}

// FindOneByObjID 查询评论
func (m *customCommentModel) FindOneByObjID(ctx context.Context, objID int64, id int64) (*Comment, error) {
	if id < 0 {
		return &Comment{
			ID:    id,
			ObjID: objID,
		}, nil

	}

	var ci *CommentIndex
	var cc *CommentContent

	err := mr.Finish(
		func() error {
			var err error
			ci, err = m.newCustomCommentIndexModelFunc(objID).FindOne(ctx, id)
			if err != nil && err != ErrNotFound {
				return err
			}
			return nil
		},
		func() error {
			var err error
			cc, err = m.newCustomCommentContentModelFunc(objID).FindOne(ctx, id)
			if err != nil && err != ErrNotFound {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &Comment{
		ID:          ci.ID,
		ObjID:       ci.ObjId,
		ObjType:     ci.ObjType,
		MemberID:    ci.MemberId,
		RootID:      ci.RootId,
		ReplyID:     ci.ReplyId,
		Floor:       ci.Floor,
		Count:       ci.Count,
		RootCount:   ci.RootCount,
		LikeCount:   ci.LikeCount,
		HateCount:   ci.HateCount,
		State:       ci.State,
		Attrs:       ci.Attrs,
		AtMemberIDs: cc.AtMemberIds,
		IP:          cc.Ip,
		Platform:    cc.Platform,
		Device:      cc.Device,
		Message:     cc.Message,
		Meta:        cc.Meta,
		CreatedAt:   ci.CreatedAt,
	}, nil
}

// CommentIndexID 评论索引ID
type CommentIndexID struct {
	ID int64 `db:"id"`
}

// CommentListByObjID 查询评论
func (m *customCommentModel) CommentListByObjID(ctx context.Context, objID int64, objType int64, sortField string, limit int64) ([]*Comment, error) {
	var (
		err error
		sql string
		// anyField any
		commentIndexIDs []*CommentIndexID
	)

	if sortField == "like_count" {
		// anyField := sortLikeCount
		sql = fmt.Sprintf("select " + " id " + " from " + m.newCustomCommentIndexModelFunc(objID).table +
			" where obj_id=? and obj_type=? order by like_count desc limit ?")
	} else {
		// anyField := sortPublishTime
		sql = fmt.Sprintf("select " + " id " + " from " + m.newCustomCommentIndexModelFunc(objID).table +
			" where obj_id=? and obj_type=? order by created_at desc limit ?")
	}

	err = m.newCustomCommentIndexModelFunc(objID).QueryRowsNoCacheCtx(ctx, &commentIndexIDs, sql, objID, objType, limit)
	if err != nil {
		logx.Error("QueryRowNoCacheCtx commentIndexs err:", err)
		return nil, err
	}

	var comments = make([]*Comment, 0, len(commentIndexIDs))
	for _, commentIndex := range commentIndexIDs {
		comment, err := m.FindOneByObjID(ctx, objID, commentIndex.ID)
		if err != nil {
			logx.Error("FindOneByObjId err:", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (m *customCommentModel) CacheCommentsByIDs(ctx context.Context, objID int64, ids []int64) ([]*Comment, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	type CommentID struct {
		ID    int64
		ObjID int64
	}

	comments, err := mr.MapReduce(
		func(source chan<- CommentID) {
			for _, id := range ids {
				source <- CommentID{
					ID:    id,
					ObjID: objID,
				}
			}
		},
		func(cid CommentID, writer mr.Writer[*Comment], cancel func(error)) {
			comment, err := m.FindOneByObjID(ctx, cid.ObjID, cid.ID)
			if err != nil {
				cancel(err)
				return
			}
			writer.Write(comment)
		},
		func(pipe <-chan *Comment, writer mr.Writer[[]*Comment], cancel func(error)) {
			var comments []*Comment
			for comment := range pipe {
				comments = append(comments, comment)
			}
			writer.Write(comments)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by IDs: %w", err)
	}
	return comments, nil
}

func (m *customCommentModel) commentsKey(objID, sortType int64) string {
	return fmt.Sprintf(PrefixCommentObjSortType, objID, sortType)
}

func (m *customCommentModel) AddCacheComments(ctx context.Context, objID int64, objType int64, sortType int64, comments []*Comment) error {
	// 缓存评论ID
	// if len(comments) == 0 {
	// 	return nil
	// }
	// // 缓存评论ID
	// commentIDs := make([]int64, 0, len(comments))
	// for _, comment := range comments {
	// 	commentIDs = append(commentIDs, comment.ID)
	// }

	panic("unimplemented")
}
