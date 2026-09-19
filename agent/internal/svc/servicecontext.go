package svc

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"

	"rpc-agent/agent/internal/config"
	"rpc-agent/client/article/article"
	"rpc-agent/client/follow/follow"
	"rpc-agent/client/like/like"
	"rpc-agent/client/reply/reply"
	"rpc-agent/client/tag/tag"
	"rpc-agent/client/user/user"
	"rpc-agent/pkg/embedding"
	"rpc-agent/pkg/interceptors"
)

type ServiceContext struct {
	Config      config.Config
	Redis       *redis.Redis
	Article     article.Article
	User        user.User
	Tag         tag.Tag
	Like        like.Like
	Follow      follow.Follow
	Reply       reply.Reply
	DB          *sql.DB        // PostgreSQL 连接，用于向量检索
	EmbedClient *embedding.Client // Doubao Embedding 客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	svc := &ServiceContext{
		Config:  c,
		Redis:   redis.MustNewRedis(c.SessionRedis),
		Article: article.NewArticle(zrpc.MustNewClient(c.ArticleRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		User:    user.NewUser(zrpc.MustNewClient(c.UserRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Tag:     tag.NewTag(zrpc.MustNewClient(c.TagRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Like:    like.NewLike(zrpc.MustNewClient(c.LikeRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Follow:  follow.NewFollow(zrpc.MustNewClient(c.FollowRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Reply:   reply.NewReply(zrpc.MustNewClient(c.ReplyRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
	}

	// 初始化 PostgreSQL 连接（用于向量检索）
	if c.DataSource != "" {
		db, err := sql.Open("postgres", c.DataSource)
		if err == nil {
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(30 * time.Minute)
			svc.DB = db
		}
	}

	// 初始化 Embedding 客户端
	if c.Embedding.APIKey != "" {
		svc.EmbedClient = embedding.NewClient(embedding.Config{
			APIKey:  c.Embedding.APIKey,
			Model:   c.Embedding.Model,
			BaseURL: c.Embedding.BaseURL,
		})
	}

	return svc
}

// ToolTimeout 返回带超时的 context。配置值单位为毫秒，默认 10 秒
func (s *ServiceContext) ToolTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	d := 10 * time.Second
	if s.Config.ToolTimeout > 0 {
		d = time.Duration(s.Config.ToolTimeout) * time.Millisecond
	}
	return context.WithTimeout(ctx, d)
}
