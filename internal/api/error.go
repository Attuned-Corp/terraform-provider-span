package api

import "fmt"

type ErrorCode string

const (
	ErrorCodeUnknownError ErrorCode = "unknown_error"
)

// Error is a proxy for API errors
// @TODO: Coerce to human readable
type Error struct {
	Code    ErrorCode
	Message string
	// response *ResponseWithMeta
}

func (e Error) Error() string {
	return fmt.Sprintf("%s [%s]", e.Message, e.Code)
}

func NewUnknownError() error {
	return &Error{Code: ErrorCodeUnknownError, Message: "Unexpected API error occurred"}
}
