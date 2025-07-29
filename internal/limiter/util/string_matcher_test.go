package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchPlain(t *testing.T) {
	var str1 string = "equal_string"
	var str2 string = "equal_string"
	var str3 string = "different_string"

	sameString := MatchPlain(str1, str2)
	differentString := MatchPlain(str1, str3)

	assert.True(t, sameString)
	assert.False(t, differentString)
}

func TestMatchRegex(t *testing.T) {
	var regex string = "^/api/item/\\d+/comment$"
	var target1 string = "/api/item/9/comment"

	result1 := MatchRegex(target1, regex)
	assert.True(t, result1)

	var target2 string = "/api/wrong_api"
	result2 := MatchRegex(target2, regex)
	assert.False(t, result2)
}
