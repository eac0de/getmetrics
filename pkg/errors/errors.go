package errors

import (
	"errors"
	"net/http"
)

type ErrorWithHTTPStatus struct {
	statusCode int
	msg        string
	err        error
}

func (ewhs *ErrorWithHTTPStatus) Error() string {
	return ewhs.msg
}

func (ewhs *ErrorWithHTTPStatus) Unwrap() error {
	return ewhs.err
}

func NewErrorWithHTTPStatus(err error, msg string, statucCode int) error {
	return &ErrorWithHTTPStatus{
		statusCode: statucCode,
		err:        err,
		msg:        msg,
	}
}

func GetMessageAndStatusCode(err error) (string, int) {
	var ewhs *ErrorWithHTTPStatus
	if errors.As(err, &ewhs) {
		return ewhs.msg, ewhs.statusCode
	} else {
		return err.Error(), http.StatusInternalServerError
	}
}
