package agentrpc

import (
	"rpc-agent/agent/internal/config"
	"rpc-agent/agent/internal/server"
	"rpc-agent/agent/internal/svc"
	"rpc-agent/agent/pb"

	"google.golang.org/grpc"
)

type Config = config.Config
type ModelConfig = config.ModelConfig
type EmbeddingConfig = config.EmbeddingConfig

func Register(grpcServer *grpc.Server, c Config) {
	svcCtx := svc.NewServiceContext(c)
	agentServer := server.NewAgentServer(svcCtx)
	pb.RegisterAgentServer(grpcServer, agentServer)
}
