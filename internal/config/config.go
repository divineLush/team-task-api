package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	RedisCfg  RedisConfig
	Cache     CacheConfig
	Auth      AuthConfig
	RateLimit RateLimitConfig
	LogLevel  string
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifeMin int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type CacheConfig struct {
	TTLMin           int
	CBFailures       int
	CBSuccesses      int
	CBTimeoutSec     int
}

type AuthConfig struct {
	JWTSecret      string
	JWTExpiryHours int
}

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}

	expiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))
	maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifeMin, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFE_MIN", "5"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	cacheTTLMin, _ := strconv.Atoi(getEnv("CACHE_TTL_MIN", "5"))
	cbFailures, _ := strconv.Atoi(getEnv("CB_FAILURES", "5"))
	cbSuccesses, _ := strconv.Atoi(getEnv("CB_SUCCESSES", "3"))
	cbTimeoutSec, _ := strconv.Atoi(getEnv("CB_TIMEOUT_SEC", "30"))
	rateLimitRPS, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "100"), 64)
	rateLimitBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "200"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "3306"),
			User:           getEnv("DB_USER", "teamtask"),
			Password:       getEnv("DB_PASSWORD", "teamtask"),
			Name:           getEnv("DB_NAME", "teamtask"),
			MaxOpenConns:   maxOpenConns,
			MaxIdleConns:   maxIdleConns,
			ConnMaxLifeMin: connMaxLifeMin,
		},
		RedisCfg: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Cache: CacheConfig{
			TTLMin:       cacheTTLMin,
			CBFailures:   cbFailures,
			CBSuccesses:  cbSuccesses,
			CBTimeoutSec: cbTimeoutSec,
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
			JWTExpiryHours: expiryHours,
		},
		RateLimit: RateLimitConfig{
			RPS:   rateLimitRPS,
			Burst: rateLimitBurst,
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
