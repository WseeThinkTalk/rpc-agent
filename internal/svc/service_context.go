package svc

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/cloudwego/eino/components/tool"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"

	"rpc-agent/client/article/article"
	"rpc-agent/client/follow/follow"
	"rpc-agent/client/like/like"
	"rpc-agent/client/reply/reply"
	"rpc-agent/client/tag/tag"
	"rpc-agent/client/user/user"
	"rpc-agent/internal/agent"
	"rpc-agent/internal/config"
	"rpc-agent/internal/model"
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
	DB          *sql.DB
	EmbedClient *embedding.Client

	AgentMgr  *agent.AgentManager
	SessMgr   *model.SessionManager
	StopChans map[string]chan struct{}
	StopMu    sync.Mutex
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds := redis.MustNewRedis(c.SessionRedis)
	svc := &ServiceContext{
		Config:    c,
		Redis:     rds,
		Article:   article.NewArticle(zrpc.MustNewClient(c.ArticleRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		User:      user.NewUser(zrpc.MustNewClient(c.UserRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Tag:       tag.NewTag(zrpc.MustNewClient(c.TagRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Like:      like.NewLike(zrpc.MustNewClient(c.LikeRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Follow:    follow.NewFollow(zrpc.MustNewClient(c.FollowRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		Reply:     reply.NewReply(zrpc.MustNewClient(c.ReplyRpc, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))),
		SessMgr:   model.NewSessionManager(rds, c.SessionTTL),
		StopChans: make(map[string]chan struct{}),
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

// InitAgent 初始化 Eino AgentManager 与工具集合
func (s *ServiceContext) InitAgent(tools []tool.BaseTool) error {
	s.AgentMgr = agent.NewAgentManager(s.Config.Model)
	return s.AgentMgr.Init(context.Background(), tools)
}

// RegisterStopChan 注册会话中断通道
func (s *ServiceContext) RegisterStopChan(sessionKey string) (chan struct{}, func()) {
	s.StopMu.Lock()
	defer s.StopMu.Unlock()
	stopCh := make(chan struct{}, 1)
	s.StopChans[sessionKey] = stopCh

	cleanup := func() {
		s.StopMu.Lock()
		delete(s.StopChans, sessionKey)
		s.StopMu.Unlock()
	}
	return stopCh, cleanup
}

// StopSession 触发会话中断
func (s *ServiceContext) StopSession(sessionKey string) {
	s.StopMu.Lock()
	defer s.StopMu.Unlock()
	if ch, ok := s.StopChans[sessionKey]; ok {
		close(ch)
	}
}

// ToolTimeout 返回带超时的 context。配置值单位为毫秒，默认 10 秒
func (s *ServiceContext) ToolTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	d := 10 * time.Second
	if s.Config.ToolTimeout > 0 {
		d = time.Duration(s.Config.ToolTimeout) * time.Millisecond
	}
	return context.WithTimeout(ctx, d)
}
