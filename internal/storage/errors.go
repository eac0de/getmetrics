package storage

type ErrorWithHTTPStatus struct {
	HTTPStatus uint16
	Err        error
}

func (ewhs *ErrorWithHTTPStatus) Error() string {
	return ewhs.Error()
}

func (ewhs *ErrorWithHTTPStatus) Unwrap() error {
	return ewhs.Err
}

func NewErrorWithHTTPStatus(err error, HTTPStatus uint16) error {
	return &ErrorWithHTTPStatus{
		HTTPStatus: HTTPStatus,
		Err:        err,
	}
}
