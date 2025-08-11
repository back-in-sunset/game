package model

import (
	"context"
	"database/sql"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TagModel = (*customTagModel)(nil)

type (
	// TagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTagModel.
	TagModel interface {
		tagModel
	}

	customTagModel struct {
		*defaultTagModel
		*dynamodb.Client
	}
)

// NewTagModel returns a model for the database table.
func NewTagModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TagModel {
	return &customTagModel{
		defaultTagModel: newTagModel(conn, c, opts...),
	}
}

func (m *customTagModel) Insert(ctx context.Context, data *Tag) (sql.Result, error) {
	panic("not implemented") // TODO: Implement
}

func (m *customTagModel) FindOne(ctx context.Context, id uint64) (*Tag, error) {
	panic("not implemented") // TODO: Implement
}

func (m *customTagModel) Update(ctx context.Context, data *Tag) error {
	panic("not implemented") // TODO: Implement
}

func (m *customTagModel) Delete(ctx context.Context, id uint64) error {
	panic("not implemented") // TODO: Implement
}
