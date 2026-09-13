package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
	Upload   UploadConfig
	Rate     RateConfig
}

type ServerConfig struct {
	Host string
	Port int
	Mode string
}

type DBConfig struct {
	Driver string
	DSN    string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type LogConfig struct {
	Level string
	File  string
}

type UploadConfig struct {
	Dir      string
	MaxSize  int64
	ImageMax int
}

type RateConfig struct {
	Enable bool
	Limit  int
	Window int
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("UPCYCLE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("db.driver", "sqlite")
	v.SetDefault("db.dsn", "upcycle.db")
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.expirehour", 24)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.file", "")
	v.SetDefault("upload.dir", "./uploads")
	v.SetDefault("upload.maxsize", 10*1024*1024)
	v.SetDefault("upload.imagemax", 1920)
	v.SetDefault("rate.enable", true)
	v.SetDefault("rate.limit", 100)
	v.SetDefault("rate.window", 60)
}
