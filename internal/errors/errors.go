package errors

import "fmt"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewNotFoundError(message string) *AppError {
	return &AppError{Code: 404, Message: message}
}

func NewBadRequestError(message string, err error) *AppError {
	return &AppError{Code: 400, Message: message, Err: err}
}

func NewInternalError(err error) *AppError {
	return &AppError{Code: 500, Message: "Internal server error", Err: err}
}
