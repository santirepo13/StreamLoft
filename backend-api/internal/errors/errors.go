package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ErrorCode string

const (
	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden   ErrorCode = "FORBIDDEN"
	ErrCodeInternal    ErrorCode = "INTERNAL"
	ErrCodeConflict    ErrorCode = "CONFLICT"
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string `json:"message"`
	Status int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func NewAPIError(code ErrorCode, message string, status int) *APIError {
	return &APIError{Code: code, Message: message, Status: status}
}

func BadRequest(message string) *APIError {
	return NewAPIError(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

func NotFound(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, http.StatusNotFound)
}

func Unauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, http.StatusForbidden)
}

func Internal(message string) *APIError {
	return NewAPIError(ErrCodeInternal, message, http.StatusInternalServerError)
}

func Conflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, http.StatusConflict)
}

func RespondWithError(c *gin.Context, err error) {
	apiErr, ok := err.(*APIError)
	if !ok {
		log.Error().Err(err).Msg("unexpected error type")
		apiErr = Internal("An unexpected error occurred")
	}

	c.JSON(apiErr.Status, ErrorResponse{Error: *apiErr})
}

func BindJSONError(c *gin.Context, err error) {
	RespondWithError(c, BadRequest("Invalid JSON: "+err.Error()))
}

func ValidationErrors(c *gin.Context, field, message string) {
	RespondWithError(c, BadRequest(message))
}