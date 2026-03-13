package customerrors

import (
	"errors"
	"net/http"
)

type HTTPError struct {
	Status int
	Msg    string
}

func (e HTTPError) Error() string {
	return e.Msg
}

func HandleError(rw http.ResponseWriter, err error) {
	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		http.Error(rw, httpErr.Error(), httpErr.Status)
	} else {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
