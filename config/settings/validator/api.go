package validator

import (
	"errors"
	"fmt"
	"regexp"
)

type ApiValidData struct {
	Identifier    string
	Path          ApiValidPath
	Method        string
	Limit         int
	WindowSeconds int
	RefillSeconds int
	ExpireSeconds int
	Target        string
}

type ApiValidPath struct {
	Expression string
	Value      string
}

func ValidateApis(apis []ApiValidData) error {
	for _, api := range apis {
		identifier := api.Identifier
		if identifier == "" {
			return errors.New("apis 항목 중 identifier가 누락되었습니다. 각 API 객체에는 고유한 identifier를 반드시 지정해야 합니다. 예: identifier: comment_write 처럼 고유한 이름을 반드시 설정해 주세요.\n")
		}

		pathExpression := api.Path.Expression
		pathValue := api.Path.Value

		if pathExpression != "regex" && pathExpression != "plain" {
			return errors.New("잘못된 pathExpression 입니다. 가능한 값: plain(일반 텍스트), regex(정규식)\n")
		}
		if pathExpression == "plain" && pathValue == "" {
			return errors.New("pathValue가 누락되었습니다.\n")
		}
		if pathExpression == "regex" {
			_, err := regexp.Compile(pathValue)
			if err != nil {
				return fmt.Errorf("잘못된 pathValue 정규표현식 입니다. error: %v\n", err)
			}
		}
	}
}
