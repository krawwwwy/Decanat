package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env         string     `yaml:"env" env-default:"local"`
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	TokenTTL    string     `yaml:"token_ttl" env-default:"24h"`
	GRPC        GRPCConfig `yaml:"grpc"`
}

type GRPCConfig struct {
	Port    string `yaml:"port" env-default:"44044"`
	Timeout string `yaml:"timeout" env-default:"10h"`
}

func MustLoad() *Config {
	var cfg Config

	configFile := os.Getenv("CONFIG_PATH")
	if configFile == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist %s:", configFile)
	}
	if err := cleanenv.ReadConfig(configFile, &cfg); err != nil {
		log.Fatalf("error loading config %s:", err)
	}
	return &cfg
}
