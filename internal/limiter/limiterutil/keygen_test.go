package limiterutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIpKeyGenerator_Make(t *testing.T) {
	generator := IpKeyGenerator{}
	result := generator.Make("11.11.11.11", "comment")
	assert.Equal(t, result, "11.11.11.11:comment")
}
