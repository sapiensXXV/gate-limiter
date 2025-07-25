package limiterconfig

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
	Key           string          `yaml:"key"`
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

	return config, nil
}
