package svc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"comment/rpc/internal/config"
	"comment/rpc/internal/eventbus"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/singleflight"
)

type ServiceContext struct {
	Config            config.Config
	CommentModel      model.CommentModel
	SignleFlightGroup singleflight.Group
	BizRedis          *redis.Redis
	LikeEventBus      eventbus.LikeEventBus

	bgStartOnce sync.Once
	bgStopFn    context.CancelFunc
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	rds, err := redis.NewRedis(redis.RedisConf{
		Host: c.BizRedis.Host,
		Pass: c.BizRedis.Pass,
		Type: c.BizRedis.Type,
	})
	if err != nil {
		panic(err)
	}

	likeEventBus, err := eventbus.NewRedisLikeEventBus(rds)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:       c,
		CommentModel: model.NewCommentModel(conn, c.CacheRedis),
		BizRedis:     rds,
		LikeEventBus: likeEventBus,
	}
}

func (s *ServiceContext) Start() {
	s.bgStartOnce.Do(func() {
		bgCtx, cancel := context.WithCancel(context.Background())
		s.bgStopFn = cancel

		consumerID := fmt.Sprintf("comment-rpc-%d", time.Now().UnixNano())
		go s.LikeEventBus.ConsumeLikeEvents(bgCtx, consumerID, func(ctx context.Context, event eventbus.LikeEvent) error {
			_, err := s.CommentModel.AdjustCommentLikeCount(ctx, event.ObjID, event.CommentID, event.Delta)
			if err != nil && err != model.ErrNotFound {
				return err
			}
			return nil
		})
	})
}

func (s *ServiceContext) Stop() {
	if s.bgStopFn != nil {
		s.bgStopFn()
	}
	if s.LikeEventBus != nil {
		if err := s.LikeEventBus.Close(); err != nil {
			logx.Errorf("close like event bus error: %v", err)
		}
	}
}
