package errors

import "net/http"

type AppError struct {
	Code       int    `json:"-"`
	Message    string `json:"message"`
	InternalErr error `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.InternalErr
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, InternalErr: err}
}

var (
	ErrNotFound       = New(http.StatusNotFound, "resource not found")
	ErrBadRequest     = New(http.StatusBadRequest, "bad request")
	ErrUnauthorized   = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden      = New(http.StatusForbidden, "forbidden")
	ErrConflict       = New(http.StatusConflict, "resource already exists")
	ErrInternal       = New(http.StatusInternalServerError, "internal server error")
	ErrUnprocessable  = New(http.StatusUnprocessableEntity, "unprocessable entity")
	ErrTooManyRequest = New(http.StatusTooManyRequests, "too many requests")
)

func IsNotFound(err error) bool {
	return isCode(err, http.StatusNotFound)
}

func IsBadRequest(err error) bool {
	return isCode(err, http.StatusBadRequest)
}

func IsUnauthorized(err error) bool {
	return isCode(err, http.StatusUnauthorized)
}

func IsConflict(err error) bool {
	return isCode(err, http.StatusConflict)
}

func isCode(err error, code int) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}
