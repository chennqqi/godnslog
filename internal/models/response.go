package models

import "net/http"

// Response represents a unified API response wrapper
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Response code constants
const (
	CodeSuccess             = 0
	CodeBadRequest          = 400
	CodeUnauthorized        = 401
	CodeForbidden           = 403
	CodeNotFound            = 404
	CodeInternalServerError = 500
	CodeServiceUnavailable  = 502
)

// Standard response messages
const (
	MessageSuccess             = "OK"
	MessageBadRequest          = "Bad Request"
	MessageUnauthorized        = "Unauthorized"
	MessageForbidden           = "Forbidden"
	MessageNotFound            = "Not Found"
	MessageInternalServerError = "Internal Server Error"
	MessageServiceUnavailable  = "Service Unavailable"
)

// SuccessResponse creates a success response with data
func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    CodeSuccess,
		Message: MessageSuccess,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// BadRequestResponse creates a 400 error response
func BadRequestResponse(message string) Response {
	if message == "" {
		message = MessageBadRequest
	}
	return Response{
		Code:    CodeBadRequest,
		Message: message,
	}
}

// UnauthorizedResponse creates a 401 error response
func UnauthorizedResponse(message string) Response {
	if message == "" {
		message = MessageUnauthorized
	}
	return Response{
		Code:    CodeUnauthorized,
		Message: message,
	}
}

// ForbiddenResponse creates a 403 error response
func ForbiddenResponse(message string) Response {
	if message == "" {
		message = MessageForbidden
	}
	return Response{
		Code:    CodeForbidden,
		Message: message,
	}
}

// NotFoundResponse creates a 404 error response
func NotFoundResponse(message string) Response {
	if message == "" {
		message = MessageNotFound
	}
	return Response{
		Code:    CodeNotFound,
		Message: message,
	}
}

// InternalServerErrorResponse creates a 500 error response
func InternalServerErrorResponse(message string) Response {
	if message == "" {
		message = MessageInternalServerError
	}
	return Response{
		Code:    CodeInternalServerError,
		Message: message,
	}
}

// ServiceUnavailableResponse creates a 502 error response
func ServiceUnavailableResponse(message string) Response {
	if message == "" {
		message = MessageServiceUnavailable
	}
	return Response{
		Code:    CodeServiceUnavailable,
		Message: message,
	}
}

// HTTPStatusCode converts response code to HTTP status code
func HTTPStatusCode(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeInternalServerError:
		return http.StatusInternalServerError
	case CodeServiceUnavailable:
		return 502 // Use 502 instead of 503 to match custom error code
	default:
		return http.StatusInternalServerError
	}
}
