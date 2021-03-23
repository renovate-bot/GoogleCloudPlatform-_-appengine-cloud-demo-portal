package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type httpError struct {
	err    error
	status int    // http status code
	msg    string // http response message
}

// Error returns internal message of e.
func (e *httpError) Error() string {
	return fmt.Sprintf("%s (%d): %s", e.msg, e.status, e.err)
}

// Errorf constructs a new error for handler.
func Errorf(ctx context.Context, status int, msg string, format string, a ...interface{}) error {
	he := &httpError{
		err:    fmt.Errorf(format, a...),
		status: status,
		msg:    msg,
	}

	first := false

	if len(a) > 0 {
		pe, ok := a[len(a)-1].(*httpError)
		if ok {
			he.status = pe.status
			he.msg = pe.msg
		} else {
			first = true
		}
	} else {
		first = true
	}

	if first {
		logger := log.Ctx(ctx).With().CallerWithSkipFrameCount(3).Logger()
		logger.Error().Err(he).Msg(he.Error())
	}

	return he
}

// Wrapf wraps an error.
func Wrapf(format string, a ...interface{}) error {
	he := &httpError{
		err:    fmt.Errorf(format, a...),
		status: http.StatusInternalServerError,
		msg:    http.StatusText(http.StatusInternalServerError),
	}

	pe, ok := a[len(a)-1].(*httpError)
	if ok {
		he.status = pe.status
		he.msg = pe.msg
	}

	return he
}

// RespondError responds an error.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	e, ok := err.(*httpError)
	if ok {
		http.Error(w, e.msg, e.status)
	} else {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

type errjson struct {
	Status  int    `json:"status"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// RespondErrorJSON responds error as JSON and logs error.
func RespondErrorJSON(w http.ResponseWriter, r *http.Request, err error) {
	ej := &errjson{}

	e, ok := err.(*httpError)
	if ok {
		ej.Status = e.status
		ej.Message = e.msg
	} else {
		ej.Status = http.StatusInternalServerError
		ej.Message = http.StatusText(ej.Status)
	}
	ej.Kind = http.StatusText(ej.Status)

	RespondJSON(w, r, ej.Status, ej)
}