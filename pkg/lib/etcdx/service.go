package etcdx

import (
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var EtcdService *Service

type Service struct {
	Config     *EtcdConfig
	EtcdClient *clientv3.Client
}

func InitEtcd(key string) *Service {
	cf := newEtcdConfig(key)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cf.Endpoints,
		DialTimeout: cf.DialTimeout,
	})
	if err != nil {
		panic(fmt.Errorf("[etcdx] failed to connect etcd cluster %v: %w", cf.Endpoints, err))
	}

	svc := &Service{
		Config:     cf,
		EtcdClient: cli,
	}

	EtcdService = svc
	return svc
}
