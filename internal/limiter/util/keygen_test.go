package util

import (
	"gate-limiter/config/settings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIpKeyGenerator_Make(t *testing.T) {
	config := &settings.RateLimiterConfig{Strategy: "token_bucket"}
	generator := NewIpKeyGenerator(*config)
	result := generator.Make("11.11.11.11", "comment")
	assert.Equal(t, result, "token_bucket:11.11.11.11:comment")
}
