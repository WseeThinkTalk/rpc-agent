package etcdx

import (
	"os"
	"strings"
	"time"
)

type EtcdConfig struct {
	Endpoints   []string
	Key         string
	DialTimeout time.Duration
}

func newEtcdConfig(key string) *EtcdConfig {
	// 确保 10.8.0.1 绕过本地 HTTP/SOCKS 代理拦截
	ensureNoProxy("10.8.0.1")

	endpointsStr := os.Getenv("ETCD_ENDPOINTS")
	var endpoints []string
	if endpointsStr != "" {
		endpoints = strings.Split(endpointsStr, ",")
	} else {
		endpoints = []string{"10.8.0.1:2379"}
	}

	return &EtcdConfig{
		Endpoints:   endpoints,
		Key:         key,
		DialTimeout: 5 * time.Second,
	}
}

func ensureNoProxy(target string) {
	current := os.Getenv("NO_PROXY")
	if current == "" {
		_ = os.Setenv("NO_PROXY", "localhost,127.0.0.1,"+target)
		return
	}
	if !strings.Contains(current, target) {
		_ = os.Setenv("NO_PROXY", current+","+target)
	}
}
