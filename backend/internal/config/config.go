package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config 集中解析全部环境变量配置，禁止在业务代码中散落读取环境变量。
type Config struct {
	AppEnv         string `env:"APP_ENV" envDefault:"development"`
	HTTPPort       string `env:"HTTP_PORT" envDefault:"8080"`
	MongoURI       string `env:"MONGO_URI" envDefault:"mongodb://petsocial_user:petsocial_pwd@mongodb:27017/petsocial_db?authSource=admin"`
	DBName         string `env:"DB_NAME" envDefault:"petsocial_db"`
	RedisAddr      string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword  string `env:"REDIS_PASSWORD" envDefault:""`
	JWTSecret      string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpireHour  int    `env:"JWT_EXPIRE_HOUR" envDefault:"72"`
	MinIOEndpoint  string `env:"MINIO_ENDPOINT" envDefault:"minio:9000"`
	MinIOAccess    string `env:"MINIO_ACCESS_KEY" envDefault:"petsocial_minio"`
	MinIOSecret    string `env:"MINIO_SECRET_KEY" envDefault:"petsocial_minio_secret"`
	MinIOBucket    string `env:"MINIO_BUCKET" envDefault:"petsocial-media"`
	MediaPublicURL string `env:"MEDIA_PUBLIC_URL" envDefault:"http://localhost:47023"`
	CORSOrigins    string `env:"CORS_ORIGINS" envDefault:"*"`
	RateLimit      int    `env:"RATE_LIMIT_PER_MINUTE" envDefault:"120"`
	SensitiveWords string `env:"SENSITIVE_WORDS" envDefault:"违禁词,赌博,暴力,色情,诈骗"`
}

// Load 解析环境变量并返回配置。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
