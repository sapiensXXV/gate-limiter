package config_ratelimiter

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type RootRateLimiterConfig struct {
	RateLimiter RateLimiterConfig `yaml:"rateLimiter"`
}

type RateLimiterConfig struct {
	Identity struct {
		Key    string `yaml:"key"`
		Header string `yaml:"header"`
	} `yaml:"identity"`

	// 사용자 전체 요청량 제한
	Client struct {
		Limit         int `yaml:"limit"`
		WindowSeconds int `yaml:"windowSeconds"`
	} `yaml:"client"`

	// 경로/행위 기준의 제한
	Apis []struct {
		Key           string          `yaml:"key"`
		Path          RateLimiterPath `yaml:"path"`
		Method        string          `yaml:"method"`
		Limit         int             `yaml:"limit"`
		WindowSeconds int             `yaml:"windowSeconds"`
	} `yaml:"apis"`
}

type RateLimiterPath struct {
	Expression string `yaml:"expression"`
	Value      string `yaml:"value"`
}

func LoadRateLimitConfig(path string) (*RootRateLimiterConfig, error) {
	buf, err := os.ReadFile(path)
	log.Printf("설정파일=[%s]를 읽습니다.", path)
	if err != nil {
		return nil, err
	}

	config := &RootRateLimiterConfig{}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return config, nil
}
