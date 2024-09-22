package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Options  *Options  `yaml:"options"`
	Pipeline []Pipline `yaml:"pipeline"`
}

type Options struct {
	GrpcServer string `yaml:"grpc_server"`
	RetryDelay []int  `yaml:"retry_delay"`
}

type Pipline struct {
	Name    string   `yaml:"name"`
	Reward  Reward   `yaml:"reward"`
	Actions []Action `yaml:"actions"`
}

type Reward struct {
	Wallets []Wallet `yaml:"wallets"`
}

type Wallet struct {
	Path     string `yaml:"path"`
	Password string `yaml:"password"`
}

type Action struct {
	Type    string   `yaml:"type"`
	Pct     float64  `yaml:"pct"`
	Time    []string `yaml:"time"`
	Targets []string `yaml:"targets"`
}

func LoadFromFile(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// 创建一个Config结构体实例
	var config Config

	// 将YAML解析为结构体
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &config, nil
}
