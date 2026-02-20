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
			return errors.New("rateLimiter.apis.identifier: missing identifier — each API must have a unique identifier (e.g. \"comment_write\")\n")
		}

		pathExpression := api.Path.Expression
		pathValue := api.Path.Value

		if pathExpression != "regex" && pathExpression != "plain" {
			return errors.New("rateLimiter.apis.path.expression: invalid value. allowed: plain, regex\n")
		}
		if pathExpression == "plain" && pathValue == "" {
			return errors.New("rateLimiter.apis.path.value: missing path value\n")
		}
		if pathExpression == "regex" {
			_, err := regexp.Compile(pathValue)
			if err != nil {
				return fmt.Errorf("rateLimiter.apis.path.value: invalid regex pattern: %v\n", err)
			}
		}

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
			return fmt.Errorf("rateLimiter.apis.method: unsupported method %q. allowed: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE\n", api.Method)
		}

		limit := api.Limit
		if limit < 0 {
			return fmt.Errorf("rateLimiter.apis.limit: must be a non-negative integer\n")
		}

		windowSeconds := api.WindowSeconds
		if windowSeconds < 0 {
			return fmt.Errorf("rateLimiter.apis.windowSeconds: must be a non-negative integer\n")
		}

		refillSeconds := api.RefillSeconds
		if refillSeconds < 0 {
			return fmt.Errorf("rateLimiter.apis.refillSeconds: must be a non-negative integer\n")
		}

		expireSeconds := api.ExpireSeconds
		if expireSeconds < 0 {
			return fmt.Errorf("rateLimiter.apis.expireSeconds: must be a non-negative integer\n")
		}

	}

	return nil
}
