package utils

import "net/http"

var (
	ErrAuth           = &APIError{Status: http.StatusUnauthorized, msg: "invalid auth token"}
	ErrNotFound       = &APIError{Status: http.StatusNotFound, msg: "not found"}
	ErrDuplicate      = &APIError{Status: http.StatusBadRequest, msg: "duplicate"}
	ErrBadCredentials = &APIError{Status: http.StatusBadRequest, msg: "login and pass must have more than 5 symbols"}
	ErrNotAuthorized  = &APIError{Status: http.StatusUnauthorized, msg: "not authorized"}
	ErrAlreadyCreated = &APIError{Status: http.StatusOK, msg: "entity already created"}
	ErrWrongFormat    = &APIError{Status: http.StatusUnprocessableEntity, msg: "entity provided has unproccessable format"}
)

type APIError struct {
	Status int
	msg    string
}

func (e APIError) Error() string {
	return e.msg
}

func (e APIError) APIError() (int, string) {
	return e.Status, e.msg
}

type WrappedAPIError struct {
	error
	APIError *APIError
}

func (we WrappedAPIError) Is(err error) bool {
	return we.APIError == err
}

func WrapError(err error, APIError *APIError) error {
	return WrappedAPIError{error: err, APIError: APIError}
}
