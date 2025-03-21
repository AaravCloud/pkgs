package nacos

import (
	"context"
	"errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zeromicro/go-zero/core/logc"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type NacosClient struct {
	config     *ConfigNaCos
	client     config_client.IConfigClient
	configData interface{}
	mu         sync.RWMutex
}

// NewNaCosClient 创建Nacos客户端实例
func NewNaCosClient(ctx context.Context, cfg *ConfigNaCos) (*NacosClient, error) {
	if err := cfg.Validate(ctx); err != nil {
		return nil, err
	}

	// 初始化服务器配置
	serverConfigs := []constant.ServerConfig{
		{IpAddr: cfg.NaCosServer, Port: uint64(cfg.NaCosPort)},
	}

	// 客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NaCosNameSpaceId,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// 创建配置客户端
	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		logc.Error(ctx, "创建Nacos客户端失败", err)
		return nil, err
	}

	return &NacosClient{
		config: cfg,
		client: client,
	}, nil
}

// WatchAndLoad 监听并加载配置
func (nc *NacosClient) WatchAndLoad(ctx context.Context, configPtr interface{}) error {
	// 首次获取配置
	data, err := nc.client.GetConfig(vo.ConfigParam{
		DataId: nc.config.NaCosDataId,
		Group:  nc.config.NaCosGroup,
	})
	if err != nil {
		return err
	}

	// 解析配置
	if err := yaml.Unmarshal([]byte(data), configPtr); err != nil {
		logc.Error(ctx, "配置解析失败", err)
		return err
	}

	// 监听配置变更
	return nc.client.ListenConfig(vo.ConfigParam{
		DataId: nc.config.NaCosDataId,
		Group:  nc.config.NaCosGroup,
		OnChange: func(_, _, _, newData string) {
			nc.mu.Lock()
			defer nc.mu.Unlock()

			if err := yaml.Unmarshal([]byte(newData), configPtr); err != nil {
				logc.Error(ctx, "配置解析失败", err)
			}
			// 使用配置
			logc.Info(ctx, "配置解析成功：", configPtr)
		},
	})
}

// Validate 配置校验
func (c *ConfigNaCos) Validate(ctx context.Context) error {
	if c == nil {
		return errors.New("nacos config is nil")
	}
	requiredFields := []struct {
		name  string
		value interface{}
	}{
		{"NaCosServer", c.NaCosServer},
		{"NaCosPort", c.NaCosPort},
		{"NaCosDataId", c.NaCosDataId},
		{"NaCosGroup", c.NaCosGroup},
		{"NaCosNameSpaceId", c.NaCosNameSpaceId},
	}

	for _, field := range requiredFields {
		switch v := field.value.(type) {
		case string:
			if v == "" {
				logc.Error(ctx, "nacos config missing required field",
					zap.String("field", field.name))
				return errors.New("nacos config validation failed")
			}
		case int:
			if v <= 0 || v > 65535 {
				logc.Error(ctx, "invalid port number",
					zap.String("field", field.name),
					zap.Int("value", v))
				return errors.New("nacos config validation failed")
			}
		}
	}
	return nil
}
