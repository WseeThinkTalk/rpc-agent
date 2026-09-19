package main

import (
	"flag"
	"fmt"

	agentrpc "rpc-agent/agent"
	"rpc-agent/pkg/env"
	"rpc-agent/pkg/interceptors"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/agent.yaml", "the config file")

type Config = agentrpc.Config

func main() {
	flag.Parse()

	env.LoadEnv()

	var c Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		agentrpc.Register(grpcServer, c)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(interceptors.ServerErrorInterceptor())
	defer s.Stop()

	fmt.Printf("Starting unified agent rpc server at %s...\n", c.ListenOn)
	s.Start()
}
