package main

import (
	"context"
	"flag"
	"fmt"

	"rpc-agent/agent"
	"rpc-agent/internal/config"
	agentserver "rpc-agent/internal/server/agent"
	"rpc-agent/internal/svc"
	"rpc-agent/pkg/lib/etcdx"
	"rpc-agent/pkg/lib/zapx"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func runRemoteConfig() *config.Config {
	var c config.Config
	etcdx.MustLoadRemoteConfig("/thinktalk/config/agent.rpc", &c)
	return &c
}

func main() {
	flag.Parse()

	// 从 Etcd 配置中心拉取远程配置 (Fail-Fast)
	c := runRemoteConfig()
	if c == nil {
		return
	}

	// init logger
	writer, err := zapx.NewZapWriter()
	if err == nil {
		logx.SetWriter(writer)
	}

	ctx := svc.NewServiceContext(*c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		registerServer(ctx, grpcServer)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(unaryServerInterceptor())
	defer s.Stop()

	fmt.Printf("Starting unified agent rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// registerServer 注册 RPC 服务
func registerServer(ctx *svc.ServiceContext, grpcServer grpc.ServiceRegistrar) {
	agent.RegisterAgentServer(grpcServer, agentserver.NewAgentServer(ctx))
}

// unaryServerInterceptor grpc 拦截器
func unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("[RPC Panic] info: %s, recover: %v", info.FullMethod, r)
			}
		}()

		resp, err := handler(ctx, req)
		return resp, err
	}
}

