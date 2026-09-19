package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Model        ModelConfig
	Embedding    EmbeddingConfig
	DataSource   string
	ToolTimeout  int64
	SessionTTL   int64
	ArticleRpc   zrpc.RpcClientConf
	UserRpc      zrpc.RpcClientConf
	TagRpc       zrpc.RpcClientConf
	LikeRpc      zrpc.RpcClientConf
	FollowRpc    zrpc.RpcClientConf
	ReplyRpc     zrpc.RpcClientConf
	SessionRedis redis.RedisConf
}

type ModelConfig struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string `json:",optional"` //nolint:SA5008
}

// EmbeddingConfig 用于 RAG 向量化的 Embedding 模型配置
type EmbeddingConfig struct {
	APIKey  string
	Model   string
	BaseURL string `json:",optional"` //nolint:SA5008
}
