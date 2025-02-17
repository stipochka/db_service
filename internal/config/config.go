package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spf13/pflag"
)

type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	DbUrl    string        `yaml:"db_url" env-required:"true"`
	GRPC     GRPCConfig    `yaml:"grpc"`
	Producer KafkaProducer `yaml:"kafka"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type KafkaProducer struct {
	Address string `yaml:"address" env-required:"true"`
	Topic   string `yaml:"topic" env-required:"true"`
	GroupID string `yaml:"group_id" env-required:"true"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if _, err := os.Stat(configPath); err != nil {
		panic("Config file doesn't exists")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("Failed to read config")
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	pflag.StringVar(&res, "config", "", "path to config file")
	pflag.Parse()

	if res == "" {
		panic("Not given path to config")
	}
	return res

}
