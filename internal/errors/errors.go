package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application-level error with HTTP status code and error code.
// It implements the error interface and supports error wrapping.
type AppError struct {
	// Code is a machine-readable error code (e.g., "bad_request", "not_found").
	Code string `json:"code"`

	// Message is a human-readable error message.
	Message string `json:"message"`

	// StatusCode is the HTTP status code to return (e.g., 400, 404, 500).
	StatusCode int `json:"-"`

	// Err is the underlying error that caused this AppError.
	Err error `json:"-"`
}

// Error implements the error interface.
// It returns a formatted error message, including the underlying error if present.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain inspection.
// This allows errors.Is() and errors.As() to work correctly.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError with the specified code, message, status code, and underlying error.
func New(code, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Wrap wraps an existing error with additional context.
// This is useful for adding HTTP status codes and error codes to standard errors.
func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common error constructors

// ErrBadRequest creates a 400 Bad Request error.
func ErrBadRequest(message string, err error) *AppError {
	return New("bad_request", message, http.StatusBadRequest, err)
}

// ErrUnauthorized creates a 401 Unauthorized error.
func ErrUnauthorized(message string, err error) *AppError {
	return New("unauthorized", message, http.StatusUnauthorized, err)
}

// ErrForbidden creates a 403 Forbidden error.
func ErrForbidden(message string, err error) *AppError {
	return New("forbidden", message, http.StatusForbidden, err)
}

// ErrNotFound creates a 404 Not Found error.
func ErrNotFound(message string, err error) *AppError {
	return New("not_found", message, http.StatusNotFound, err)
}

// ErrConflict creates a 409 Conflict error.
func ErrConflict(message string, err error) *AppError {
	return New("conflict", message, http.StatusConflict, err)
}

// ErrInternal creates a 500 Internal Server Error.
func ErrInternal(message string, err error) *AppError {
	return New("internal_error", message, http.StatusInternalServerError, err)
}

// ErrServiceUnavailable creates a 503 Service Unavailable error.
func ErrServiceUnavailable(message string, err error) *AppError {
	return New("service_unavailable", message, http.StatusServiceUnavailable, err)
}

// Predefined errors for common scenarios.
// These can be used directly or wrapped with additional context.
var (
	// ErrInvalidInput indicates that the request contains invalid data.
	ErrInvalidInput = ErrBadRequest("Invalid input", nil)

	// ErrMissingAuthHeader indicates that the Authorization header is missing.
	ErrMissingAuthHeader = ErrUnauthorized("Missing authorization header", nil)

	// ErrInvalidToken indicates that the provided token is invalid or malformed.
	ErrInvalidToken = ErrUnauthorized("Invalid token", nil)

	// ErrExpiredToken indicates that the provided token has expired.
	ErrExpiredToken = ErrUnauthorized("Token expired", nil)

	// ErrResourceNotFound indicates that the requested resource doesn't exist.
	ErrResourceNotFound = ErrNotFound("Resource not found", nil)

	// ErrDatabaseError indicates a database operation failed.
	ErrDatabaseError = ErrInternal("Database error", nil)

	// ErrStorageError indicates a storage operation failed.
	ErrStorageError = ErrInternal("Storage error", nil)
)
