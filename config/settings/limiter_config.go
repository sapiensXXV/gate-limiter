package settings

import (
	"fmt"
	"gate-limiter/config/settings/validator"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type RootRateLimiterConfig struct {
	RateLimiter RateLimiterConfig `yaml:"rateLimiter"`
	RedisConfig RedisClientConfig `yaml:"redis"`
}

type RedisClientConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RateLimiterConfig struct {
	Strategy string         `yaml:"strategy"`
	Identity ClientIdentity `yaml:"identity"`
	Client   ClientLimit    `yaml:"client"`
	Apis     []Api          `yaml:"apis"`
	Target   string         `yaml:"target"`
	Port     int            `yaml:"port"`
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

func LoadRateLimitConfig(path string) (*RootRateLimiterConfig, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &RootRateLimiterConfig{}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	validateConfig(config)

	printBanner(config)
	fmt.Printf("strategy: %-20s\n", config.RateLimiter.Strategy)
	fmt.Printf("category: %-20s\n", config.RateLimiter.Identity.Key)

	printApiInfo(config.RateLimiter.Apis)

	return config, nil
}

func printApiInfo(apis []Api) {
	fmt.Printf("📘API registered\n")
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

// validateConfig 설정정보가 올바른지 검사하는 메서드
func validateConfig(config *RootRateLimiterConfig) {
	limiterConfig := config.RateLimiter
	_, err := validator.ValidateStrategy(limiterConfig.Strategy)
	if err != nil {
		log.Fatal("validate configuration failed: ", err)
	}

	if err := validator.ValidateIdentity(limiterConfig.Identity.Key, limiterConfig.Identity.Header); err != nil {
		log.Fatal("validate configuration failed: ", err)
	}

	apis := createValidateApis(limiterConfig.Apis)
	if err := validator.ValidateApis(apis); err != nil {
		log.Fatal("validate configuration failed: ", err)
	}

	portConfig(config)
}

func portConfig(config *RootRateLimiterConfig) {
	if config.RateLimiter.Port == 0 {
		config.RateLimiter.Port = 8081 // 포트 기본값 8081
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

func printBanner(config *RootRateLimiterConfig) {
	pid := os.Getpid()

	fmt.Printf("Version %s\n", "v0.1.0")
	fmt.Printf("Port: %d\n", config.RateLimiter.Port)
	fmt.Printf("PID: %d\n", pid)
	fmt.Printf("Github: https://github.com/sapiensXXV/gate-limiter\n\n")
}
