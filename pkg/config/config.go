package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig        `yaml:"app"`
	Server    ServerConfig     `yaml:"server"`
	Database  DatabaseConfig   `yaml:"database"`
	Redis     RedisConfig      `yaml:"redis"`
	Provider  ProviderConfig   `yaml:"provider"`
	Providers []ProviderSource `yaml:"providers"`
}

type ProviderSource struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Format    string `yaml:"format"`
	RateLimit int    `yaml:"rate_limit"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type ProviderConfig struct {
	Timeout                 time.Duration `yaml:"timeout"`
	CircuitBreakerThreshold int           `yaml:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `yaml:"circuit_breaker_timeout"`
	RateLimitMax            int           `yaml:"rate_limit_max"`
	RateLimitWindow         time.Duration `yaml:"rate_limit_window"`
}

func Load(path string) (*Config, error) {
	godotenv.Load()

	cfg := &Config{}

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	cfg.applyEnvOverrides()

	cfg.setDefaults()

	return cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("APP_ENV"); v != "" {
		c.App.Env = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		c.Server.Port = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		c.Database.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		c.Database.Name = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		c.Redis.Port = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}

	if v := os.Getenv("PROVIDER_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil {
			c.Provider.Timeout = time.Duration(timeout) * time.Second
		}
	}
	if v := os.Getenv("CIRCUIT_BREAKER_THRESHOLD"); v != "" {
		if threshold, err := strconv.Atoi(v); err == nil {
			c.Provider.CircuitBreakerThreshold = threshold
		}
	}
	if v := os.Getenv("CIRCUIT_BREAKER_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil {
			c.Provider.CircuitBreakerTimeout = time.Duration(timeout) * time.Second
		}
	}
	if v := os.Getenv("RATE_LIMIT_MAX"); v != "" {
		if max, err := strconv.Atoi(v); err == nil {
			c.Provider.RateLimitMax = max
		}
	}
	if v := os.Getenv("RATE_LIMIT_WINDOW"); v != "" {
		if window, err := strconv.Atoi(v); err == nil {
			c.Provider.RateLimitWindow = time.Duration(window) * time.Second
		}
	}
}

func (c *Config) setDefaults() {
	if c.App.Name == "" {
		c.App.Name = "search-engine"
	}
	if c.App.Env == "" {
		c.App.Env = "development"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Database.Host == "" {
		c.Database.Host = "localhost"
	}
	if c.Database.Port == "" {
		c.Database.Port = "5432"
	}
	if c.Database.User == "" {
		c.Database.User = "postgres"
	}
	if c.Database.Password == "" {
		c.Database.Password = "postgres"
	}
	if c.Database.Name == "" {
		c.Database.Name = "search_engine"
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Redis.Host == "" {
		c.Redis.Host = "localhost"
	}
	if c.Redis.Port == "" {
		c.Redis.Port = "6379"
	}
	if c.Provider.Timeout == 0 {
		c.Provider.Timeout = 5 * time.Second
	}
	if c.Provider.CircuitBreakerThreshold == 0 {
		c.Provider.CircuitBreakerThreshold = 5
	}
	if c.Provider.CircuitBreakerTimeout == 0 {
		c.Provider.CircuitBreakerTimeout = 30 * time.Second
	}
	if c.Provider.RateLimitMax == 0 {
		c.Provider.RateLimitMax = 100
	}
	if c.Provider.RateLimitWindow == 0 {
		c.Provider.RateLimitWindow = 60 * time.Second
	}
}

func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Name +
		" sslmode=" + c.SSLMode
}

func (c *DatabaseConfig) ConnectionString() string {
	return "postgres://" + c.User + ":" + c.Password +
		"@" + c.Host + ":" + c.Port +
		"/" + c.Name + "?sslmode=" + c.SSLMode
}

func (c *RedisConfig) Addr() string {
	return c.Host + ":" + c.Port
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
