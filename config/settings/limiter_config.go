package settings

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type RootRateLimiterConfig struct {
	RateLimiter RateLimiterConfig `yaml:"rateLimiter"`
}

type RateLimiterConfig struct {
	Strategy string         `yaml:"strategy"`
	Identity ClientIdentity `yaml:"identity"`
	Client   ClientLimit    `yaml:"client"`
	Apis     []Api          `yaml:"apis"`
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
	BucketSize    int             `yaml:"bucketSize"`
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

	log.Printf("[사용전략] %20s\n", config.RateLimiter.Strategy)
	log.Printf("[유저구분] %20s\n", config.RateLimiter.Identity.Key)
	var apis []Api
	apis = config.RateLimiter.Apis
	log.Printf("[API 구분]\n")
	for _, api := range apis {
		log.Printf("  [이름] %s\n", api.Identifier)
		log.Printf("  [경로]")
		log.Printf("    -표현법: %s\n", api.Path.Expression)
		log.Printf("    -값: %s\n", api.Path.Value)
		log.Printf("  [메서드]: %s\n", api.Method)
		log.Printf("  [제한 요청 수]: %d\n", api.Limit)
		log.Printf("  [윈도우 초기화 시간(초)]: %d\n", api.WindowSeconds)
		log.Printf("  [토큰 버킷 리필 주기(초)]: %d\n", api.RefillSeconds)
		log.Printf("  [버킷 만료 시간(초)]: %d\n", api.ExpireSeconds)
		log.Printf("  [(누출)버킷 사이즈]: %d\n", api.BucketSize)

	}

	return config, nil
}
