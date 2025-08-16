package config

import (
	"flag"
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	ListenIP         string
	ListenPort       int
	TokenAuthFile    string
	TokenAuthEnabled bool
}

// Load 加载配置
func Load() *Config {
	cfg := &Config{}

	// 设置命令行参数 - 支持短参数和长参数
	flag.StringVar(&cfg.ListenIP, "ip", "127.0.0.1", "监听IP地址")
	flag.StringVar(&cfg.ListenIP, "i", "127.0.0.1", "监听IP地址 (短参数)")

	flag.IntVar(&cfg.ListenPort, "port", 15000, "监听端口号")
	flag.IntVar(&cfg.ListenPort, "p", 15000, "监听端口号 (短参数)")

	flag.StringVar(&cfg.TokenAuthFile, "token-file", "tokens.txt", "Token白名单文件路径")
	flag.StringVar(&cfg.TokenAuthFile, "t", "tokens.txt", "Token白名单文件路径 (短参数)")

	flag.Parse()

	// 验证参数
	if cfg.ListenPort < 1 || cfg.ListenPort > 65535 {
		fmt.Printf("无效的端口号: %d, 端口范围应为 1-65535\n", cfg.ListenPort)
		os.Exit(1)
	}

	return cfg
}

// GetAddress 获取监听地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.ListenIP, c.ListenPort)
}
