package server

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/tensorchord/openmodelz/agent/errdefs"
)

// Error defines a standard application error.
type Error struct {
	// Machine-readable error code.
	HTTPStatusCode int `json:"http_status_code,omitempty"`

	// Human-readable message.
	Message string `json:"message,omitempty"`
	Request string `json:"request,omitempty"`

	// Logical operation and nested error.
	Op  string `json:"op,omitempty"`
	Err error  `json:"error,omitempty"`
}

// Error returns the string representation of the error message.
func (e *Error) Error() string {
	var buf bytes.Buffer

	// Print the current operation in our stack, if any.
	if e.Op != "" {
		fmt.Fprintf(&buf, "%s: ", e.Op)
	}

	// If wrapping an error, print its Error() message.
	// Otherwise print the error code & message.
	if e.Err != nil {
		buf.WriteString(e.Err.Error())
	} else {
		if e.HTTPStatusCode != 0 {
			fmt.Fprintf(&buf, "<%s> ", http.StatusText(e.HTTPStatusCode))
		}
		buf.WriteString(e.Message)
	}
	return buf.String()
}

func NewError(code int, err error, op string) error {
	return &Error{
		HTTPStatusCode: code,
		Err:            err,
		Message:        err.Error(),
		Op:             op,
	}
}

func errFromErrDefs(err error, op string) error {
	if errdefs.IsCancelled(err) {
		return NewError(http.StatusRequestTimeout, err, op)
	} else if errdefs.IsConflict(err) {
		return NewError(http.StatusConflict, err, op)
	} else if errdefs.IsDataLoss(err) {
		return NewError(http.StatusInternalServerError, err, op)
	} else if errdefs.IsDeadline(err) {
		return NewError(http.StatusRequestTimeout, err, op)
	} else if errdefs.IsForbidden(err) {
		return NewError(http.StatusForbidden, err, op)
	} else if errdefs.IsInvalidParameter(err) {
		return NewError(http.StatusBadRequest, err, op)
	} else if errdefs.IsNotFound(err) {
		return NewError(http.StatusNotFound, err, op)
	} else if errdefs.IsNotImplemented(err) {
		return NewError(http.StatusNotImplemented, err, op)
	} else if errdefs.IsNotModified(err) {
		return NewError(http.StatusNotModified, err, op)
	} else if errdefs.IsSystem(err) {
		return NewError(http.StatusInternalServerError, err, op)
	} else if errdefs.IsUnauthorized(err) {
		return NewError(http.StatusUnauthorized, err, op)
	} else if errdefs.IsUnavailable(err) {
		return NewError(http.StatusServiceUnavailable, err, op)
	} else if errdefs.IsUnknown(err) {
		return NewError(http.StatusInternalServerError, err, op)
	}
	return NewError(http.StatusInternalServerError, err, op)
}
