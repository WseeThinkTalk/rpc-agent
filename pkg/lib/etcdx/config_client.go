package etcdx

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const configPath = "./etc/etcd.yaml"

// saveConfigToFile 将拉取到的远程配置落盘持久化到本地 etc/etcd.yaml
func saveConfigToFile(content []byte) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(content); err != nil {
		return err
	}
	return writer.Flush()
}

// GetRemoteConfig 从 Etcd 拉取指定 Key 的完整 YAML 配置并解析到结构体 v
func (s *Service) GetRemoteConfig(v any) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.DialTimeout)
	defer cancel()

	resp, err := s.EtcdClient.Get(ctx, s.Config.Key)
	if err != nil {
		return fmt.Errorf("failed to get config key %s: %w", s.Config.Key, err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("config key %s not found in etcd cluster %v", s.Config.Key, s.Config.Endpoints)
	}

	content := resp.Kvs[0].Value
	if err = conf.LoadFromYamlBytes(content, v); err != nil {
		return fmt.Errorf("failed to unmarshal yaml config for key %s: %w", s.Config.Key, err)
	}

	// 持久化保存到 ./etc/etcd.yaml
	if err := saveConfigToFile(content); err != nil {
		fmt.Printf("[etcdx] Warning: failed to save %s: %v\n", configPath, err)
	} else {
		fmt.Printf("[etcdx] Successfully created/updated %s\n", configPath)
	}

	fmt.Printf("[etcdx] Successfully loaded remote config from %s (length: %d bytes)\n", s.Config.Key, len(content))

	// 开启异步热监听
	go s.watchConfig(v)

	return nil
}

// watchConfig 持续监听 Etcd 配置变更事件
func (s *Service) watchConfig(v any) {
	rch := s.EtcdClient.Watch(context.Background(), s.Config.Key)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == clientv3.EventTypePut {
				logx.Infof("[etcdx] Config changed for key: %s, updating in-memory config", s.Config.Key)
				if err := conf.LoadFromYamlBytes(ev.Kv.Value, v); err != nil {
					logx.Errorf("[etcdx] Failed to reload changed config: %v", err)
				} else {
					_ = saveConfigToFile(ev.Kv.Value)
					logx.Infof("[etcdx] Config successfully reloaded and saved to %s", configPath)
				}
			}
		}
	}
}

// MustLoadRemoteConfig 对标 CYZX 规范的顶层快捷加载函数，遇错直接 panic (Fail-Fast)
func MustLoadRemoteConfig(key string, v any) {
	svc := InitEtcd(key)
	if err := svc.GetRemoteConfig(v); err != nil {
		panic(fmt.Errorf("[etcdx] Fatal error loading remote config: %w", err))
	}
}
