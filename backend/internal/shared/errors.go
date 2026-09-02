package shared

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorDetail represents detailed information about an error.
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse matches the OpenAPI specification for error payloads.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// AppError represents a structured application error with HTTP status code.
type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]interface{}
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeInternal           = "INTERNAL_SERVER_ERROR"
)

// NewAppError creates a new AppError.
func NewAppError(statusCode int, code, message string, details map[string]interface{}, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
		Err:        err,
	}
}

// Predefined error builders
func ErrBadRequest(code, message string, details map[string]interface{}) *AppError {
	return NewAppError(http.StatusBadRequest, code, message, details, nil)
}

func ErrUnauthorized(code, message string) *AppError {
	return NewAppError(http.StatusUnauthorized, code, message, nil, nil)
}

func ErrForbidden(code, message string) *AppError {
	return NewAppError(http.StatusForbidden, code, message, nil, nil)
}

func ErrNotFound(code, message string) *AppError {
	return NewAppError(http.StatusNotFound, code, message, nil, nil)
}

func ErrConflict(code, message string) *AppError {
	return NewAppError(http.StatusConflict, code, message, nil, nil)
}

func ErrInternal(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, ErrCodeInternal, "Error intern del servidor.", nil, err)
}

func ErrPayloadTooLarge(code, message string, details map[string]interface{}) *AppError {
	return NewAppError(http.StatusRequestEntityTooLarge, code, message, details, nil)
}

// RespondWithError writes a standardized JSON error response.
func RespondWithError(c *gin.Context, appErr *AppError) {
	c.JSON(appErr.StatusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	})
}
