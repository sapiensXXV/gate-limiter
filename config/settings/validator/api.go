package validator

import (
	"errors"
	"fmt"
	"net/http"
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
			return errors.New("rateLimiter.apis.identifier: apis 항목 중 identifier가 누락되었습니다. 각 API 객체에는 고유한 identifier를 반드시 지정해야 합니다. 예: identifier: comment_write 처럼 고유한 이름을 반드시 설정해 주세요.\n")
		}

		// API 경로 표현 검사
		pathExpression := api.Path.Expression
		pathValue := api.Path.Value

		if pathExpression != "regex" && pathExpression != "plain" {
			return errors.New("잘못된 rateLimiter.apis.path.expression: pathExpression 입니다. 가능한 값: plain(일반 텍스트), regex(정규식)\n")
		}
		if pathExpression == "plain" && pathValue == "" {
			return errors.New("rateLimiter.apis.path.value: pathValue가 누락되었습니다.\n")
		}
		if pathExpression == "regex" {
			_, err := regexp.Compile(pathValue)
			if err != nil {
				return fmt.Errorf("rateLimiter.apis.path.value: 잘못된 pathValue 정규표현식 입니다. error: %v\n", err)
			}
		}

		// API 메서드 검사
		switch api.Method {
		case http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodTrace:
		default:
			return fmt.Errorf("rateLimiter.apis.Method: 지원하지 않는 메서드 %q 입니다. 허용된 값: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE\n", api.Method)
		}

		// 임계치 검사
		limit := api.Limit
		if limit < 0 {
			return fmt.Errorf("rateLimiter.apis.limit: 잘못된 임계치 표현입니다. 임계치는 0보다 크거나 같은 양의 정수여야 합니다.\n")
		}

		// 윈도우 사이즈 검사
		windowSeconds := api.WindowSeconds
		if windowSeconds < 0 {
			return fmt.Errorf("rateLimiter.apis.windowSeconds: 잘못된 윈도우 사이즈 입니다. 윈도우 사이즈는 0보다 크거나 같은 정수여야 합니다.\n")
		}

		// 버킷 토큰 리필 주기 검사
		refillSeconds := api.RefillSeconds
		if refillSeconds < 0 {
			return fmt.Errorf("rateLimiter.apis.refillSeconds: 잘못된 토큰 리필 주기입니다. 토큰 리필 주기는 0보다 크거나 같은 정수여야 합니다.\n")
		}

		// 버킷 or 윈도우 만료시간 검사
		expireSeconds := api.ExpireSeconds
		if expireSeconds < 0 {
			return fmt.Errorf("rateLimiter.apis.expireSecdons: 잘못된 만료 시간(초)입니다. 버킷/윈도우의 만료 주기는 0보다 크거나 같은 정수여야 합니다.\n")
		}

	}

	return nil
}
