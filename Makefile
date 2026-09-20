## rpc 生成命令
rpc:
	goctl rpc protoc agent.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=./agent --go-grpc_out=./agent --zrpc_out=./ -m --verbose --style=go_zero

run:
	go run agent.go

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/rpc-agent agent.go
