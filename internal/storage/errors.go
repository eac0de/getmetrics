package storage

type ErrorWithHTTPStatus struct {
	HTTPStatus int
	Err        error
}

func (ewhs *ErrorWithHTTPStatus) Error() string {
	return ewhs.Err.Error()
}

func (ewhs *ErrorWithHTTPStatus) Unwrap() error {
	return ewhs.Err
}

func NewErrorWithHTTPStatus(err error, HTTPStatus int) error {
	return &ErrorWithHTTPStatus{
		HTTPStatus: HTTPStatus,
		Err:        err,
	}
}
