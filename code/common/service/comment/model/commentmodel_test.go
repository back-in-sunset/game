package model

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func Test_customCommentModel_AddComment(t *testing.T) {
	conn := sqlx.NewMysql("root:123456@tcp(127.0.0.1:3306)/game")
	c := cache.CacheConf{
		{
			Weight: 100,
			RedisConf: redis.RedisConf{
				Host: "localhost:6379",
				Type: "node",
			},
		},
	}

	type fields struct {
		CachedConn                       sqlc.CachedConn
		newCustomCommentSubjectModelFunc newCustomCommentSubjectModelFunc
		newCustomCommentIndexModelFunc   newCustomCommentIndexModelFunc
		newCustomCommentContentModelFunc newCustomCommentContentModelFunc
	}
	type args struct {
		ctx  context.Context
		data *CommentSubject
		ci   *CommentIndex
		cc   *CommentContent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommentSchema
		wantErr bool
	}{
		{
			name: "AddComment",
			fields: fields{
				CachedConn: sqlc.NewConn(conn, c),
				newCustomCommentSubjectModelFunc: func(shardingID int64) *customCommentSubjectModel {
					return newCustomCommentSubjectModel(conn, c, shardingID)
				},
				newCustomCommentIndexModelFunc: func(shardingID int64) *customCommentIndexModel {
					return newCustomCommentIndexModel(conn, c, shardingID)
				},
				newCustomCommentContentModelFunc: func(shardingID int64) *customCommentContentModel {
					return newCustomCommentContentModel(conn, c, shardingID)
				},
			},
			args: args{
				ctx: context.Background(),
				data: &CommentSubject{
					ID:        1,
					ObjId:     1,
					ObjType:   1,
					MemberId:  1,
					Count:     0,
					RootCount: 0,
					AllCount:  0,
					State:     0,
					Attrs:     0,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ci: &CommentIndex{
					ID:        1,
					ObjId:     1,
					ObjType:   1,
					MemberId:  1,
					RootId:    0,
					ReplyId:   0,
					Floor:     0,
					Count:     0,
					RootCount: 0,
					LikeCount: 0,
					HateCount: 0,
					State:     0,
					Attrs:     0,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				cc: &CommentContent{
					CommentId:   1,
					ObjId:       1,
					AtMemberIds: "",
					Ip:          "192.168.1.1",
					Platform:    1,
					Device:      "apple",
					Message:     "大胆",
					Meta:        "",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &customCommentModel{
				CachedConn:                       tt.fields.CachedConn,
				newCustomCommentSubjectModelFunc: tt.fields.newCustomCommentSubjectModelFunc,
				newCustomCommentIndexModelFunc:   tt.fields.newCustomCommentIndexModelFunc,
				newCustomCommentContentModelFunc: tt.fields.newCustomCommentContentModelFunc,
			}
			got, err := m.AddComment(tt.args.ctx, tt.args.data, tt.args.ci, tt.args.cc)
			if (err != nil) != tt.wantErr {
				t.Errorf("customCommentModel.AddComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("customCommentModel.AddComment() = %v, want %v", got, tt.want)
			}
		})
	}
}
