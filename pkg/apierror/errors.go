package apierror

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

const (
	CodeValidationError   = "VALIDATION_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeProviderError     = "PROVIDER_ERROR"
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeBadRequest        = "BAD_REQUEST"
)

var (
	ErrQueryRequired = &APIError{
		Code:       CodeValidationError,
		Message:    "Query parameter 'q' is required",
		StatusCode: http.StatusBadRequest,
	}

	ErrInvalidContentType = &APIError{
		Code:       CodeValidationError,
		Message:    "Invalid content type. Must be 'video' or 'text'",
		StatusCode: http.StatusBadRequest,
	}

	ErrInvalidSortField = &APIError{
		Code:       CodeValidationError,
		Message:    "Invalid sort field. Must be 'popularity' or 'relevance'",
		StatusCode: http.StatusBadRequest,
	}

	ErrInternalServer = &APIError{
		Code:       CodeInternalError,
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrProviderUnavailable = &APIError{
		Code:       CodeProviderError,
		Message:    "Content providers are temporarily unavailable",
		StatusCode: http.StatusServiceUnavailable,
	}

	ErrRateLimitExceeded = &APIError{
		Code:       CodeRateLimitExceeded,
		Message:    "Rate limit exceeded. Please try again later",
		StatusCode: http.StatusTooManyRequests,
	}

	ErrNotFound = &APIError{
		Code:       CodeNotFound,
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}
)

func NewValidationError(message string) *APIError {
	return &APIError{
		Code:       CodeValidationError,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

func NewProviderError(providerName string, err error) *APIError {
	return &APIError{
		Code:       CodeProviderError,
		Message:    fmt.Sprintf("Provider '%s' error: %v", providerName, err),
		StatusCode: http.StatusServiceUnavailable,
	}
}

func NewInternalError(err error) *APIError {
	return &APIError{
		Code:       CodeInternalError,
		Message:    err.Error(),
		StatusCode: http.StatusInternalServerError,
	}
}
