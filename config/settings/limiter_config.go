package settings

import (
	"fmt"
	"gate-limiter/config/settings/validator"
	"os"

	"gopkg.in/yaml.v3"
)

type RootRateLimiterConfig struct {
	RateLimiter RateLimiterConfig `yaml:"rateLimiter"`
	RedisConfig RedisClientConfig `yaml:"redis"`
	Logging     LoggingConfig     `yaml:"logging"`
}

type RedisClientConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RateLimiterConfig struct {
	Strategy  string         `yaml:"strategy"`
	Identity  ClientIdentity `yaml:"identity"`
	Client    ClientLimit    `yaml:"client"`
	Apis      []Api          `yaml:"apis"`
	Target    string         `yaml:"target"`
	Port      int            `yaml:"port"`
	AdminPort int            `yaml:"adminPort"`
}

type ClientIdentity struct {
	Key    string `yaml:"key"`
	Header string `yaml:"header"`
}

type ClientLimit struct {
	Limit         int `yaml:"limit"`
	WindowSeconds int `yaml:"windowSeconds"`
}

type Api struct {
	Identifier    string          `yaml:"identifier"`
	Path          RateLimiterPath `yaml:"path"`
	Method        string          `yaml:"method"`
	Limit         int             `yaml:"limit"`
	WindowSeconds int             `yaml:"windowSeconds"`
	RefillSeconds int             `yaml:"refillSeconds"`
	ExpireSeconds int             `yaml:"expireSeconds"`
	Target        string          `yaml:"target"`
}

type RateLimiterPath struct {
	Expression string `yaml:"expression"`
	Value      string `yaml:"value"`
}

type LoggingConfig struct {
	Level  string        `yaml:"level"`  // debug | info | warn | error (default: info)
	Format string        `yaml:"format"` // text | json (default: text)
	Output string        `yaml:"output"` // stdout | stderr | file (default: stdout)
	File   LogFileConfig `yaml:"file"`
}

type LogFileConfig struct {
	Directory  string `yaml:"directory"`  // log directory (default: ./logs)
	MaxSizeMB  int    `yaml:"maxSizeMB"`  // max file size in MB (default: 100)
	MaxAgeDays int    `yaml:"maxAgeDays"` // retention days (default: 30)
	Compress   bool   `yaml:"compress"`   // gzip after rotation (default: false)
}

// ParseAndValidateConfig parses and validates config without printing.
func ParseAndValidateConfig(path string) (*RootRateLimiterConfig, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &RootRateLimiterConfig{}
	if err := yaml.Unmarshal(buf, config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadRateLimitConfig loads config, validates, and prints banner.
func LoadRateLimitConfig(path string) (*RootRateLimiterConfig, error) {
	config, err := ParseAndValidateConfig(path)
	if err != nil {
		return nil, err
	}

	PrintBanner(config)
	PrintApiInfo(config.RateLimiter.Apis)

	return config, nil
}

func PrintBanner(config *RootRateLimiterConfig) {
	pid := os.Getpid()

	fmt.Printf("Version %s\n", "v0.2.0")
	fmt.Printf("Port: %d\n", config.RateLimiter.Port)
	fmt.Printf("PID: %d\n", pid)
	fmt.Printf("Github: https://github.com/sapiensXXV/gate-limiter\n\n")
	fmt.Printf("strategy: %-20s\n", config.RateLimiter.Strategy)
	fmt.Printf("category: %-20s\n", config.RateLimiter.Identity.Key)
}

func PrintApiInfo(apis []Api) {
	fmt.Printf("API registered\n")
	for _, api := range apis {
		fmt.Printf("  identifier       : %s\n", api.Identifier)
		fmt.Printf("  - method           : %s\n", api.Method)
		fmt.Printf("  - path expression  : %s\n", api.Path.Expression)
		fmt.Printf("  - path value       : %s\n", api.Path.Value)
		fmt.Printf("  - limit            : %d requests\n", api.Limit)
		fmt.Printf("  - window duration  : %d sec\n", api.WindowSeconds)
		fmt.Printf("  - token refill     : %d sec\n", api.RefillSeconds)
		fmt.Printf("  - expiration time  : %d ms\n", api.ExpireSeconds)
	}
}

func validateConfig(config *RootRateLimiterConfig) error {
	limiterConfig := config.RateLimiter
	_, err := validator.ValidateStrategy(limiterConfig.Strategy)
	if err != nil {
		return fmt.Errorf("validate configuration failed: %w", err)
	}

	if err := validator.ValidateIdentity(limiterConfig.Identity.Key, limiterConfig.Identity.Header); err != nil {
		return fmt.Errorf("validate configuration failed: %w", err)
	}

	apis := createValidateApis(limiterConfig.Apis)
	if err := validator.ValidateApis(apis); err != nil {
		return fmt.Errorf("validate configuration failed: %w", err)
	}

	portConfig(config)
	applyLoggingDefaults(config)
	return nil
}

func applyLoggingDefaults(config *RootRateLimiterConfig) {
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
	if config.Logging.File.Directory == "" {
		config.Logging.File.Directory = "./logs"
	}
	if config.Logging.File.MaxSizeMB <= 0 {
		config.Logging.File.MaxSizeMB = 100
	}
	if config.Logging.File.MaxAgeDays <= 0 {
		config.Logging.File.MaxAgeDays = 30
	}
}

func portConfig(config *RootRateLimiterConfig) {
	if config.RateLimiter.Port == 0 {
		config.RateLimiter.Port = 8081
	}
	if config.RateLimiter.AdminPort == 0 {
		config.RateLimiter.AdminPort = 8082
	}
}

func createValidateApis(apis []Api) []validator.ApiValidData {
	result := make([]validator.ApiValidData, 0)
	for _, api := range apis {
		newPath := validator.ApiValidPath{
			Expression: api.Path.Expression,
			Value:      api.Path.Value,
		}

		newApi := &validator.ApiValidData{
			Identifier:    api.Identifier,
			Path:          newPath,
			Method:        api.Method,
			Limit:         api.Limit,
			WindowSeconds: api.WindowSeconds,
			RefillSeconds: api.RefillSeconds,
			ExpireSeconds: api.ExpireSeconds,
		}

		result = append(result, *newApi)
	}
	return result
}
