// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client // import "github.com/docker/docker/client"

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"

	"github.com/tensorchord/openmodelz/agent/errdefs"
)

// errConnectionFailed implements an error returned when connection failed.
type errConnectionFailed struct {
	host string
}

// Error returns a string representation of an errConnectionFailed
func (err errConnectionFailed) Error() string {
	if err.host == "" {
		return "Cannot connect to the backend"
	}
	return fmt.Sprintf("Cannot connect at %s", err.host)
}

// IsErrConnectionFailed returns true if the error is caused by connection failed.
func IsErrConnectionFailed(err error) bool {
	return errors.As(err, &errConnectionFailed{})
}

// ErrorConnectionFailed returns an error with host in the error message when connection to docker daemon failed.
func ErrorConnectionFailed(host string) error {
	return errConnectionFailed{host: host}
}

// Deprecated: use the errdefs.NotFound() interface instead. Kept for backward compatibility
type notFound interface {
	error
	NotFound() bool
}

// IsErrNotFound returns true if the error is a NotFound error, which is returned
// by the API when some object is not found.
func IsErrNotFound(err error) bool {
	if errdefs.IsNotFound(err) {
		return true
	}
	var e notFound
	return errors.As(err, &e)
}

type objectNotFoundError struct {
	object string
	id     string
}

func (e objectNotFoundError) NotFound() {}

func (e objectNotFoundError) Error() string {
	return fmt.Sprintf("Error: No such %s: %s", e.object, e.id)
}

// IsErrUnauthorized returns true if the error is caused
// when a remote registry authentication fails
//
// Deprecated: use errdefs.IsUnauthorized
func IsErrUnauthorized(err error) bool {
	return errdefs.IsUnauthorized(err)
}

type pluginPermissionDenied struct {
	name string
}

func (e pluginPermissionDenied) Error() string {
	return "Permission denied while installing plugin " + e.name
}

// IsErrNotImplemented returns true if the error is a NotImplemented error.
// This is returned by the API when a requested feature has not been
// implemented.
//
// Deprecated: use errdefs.IsNotImplemented
func IsErrNotImplemented(err error) bool {
	return errdefs.IsNotImplemented(err)
}

func wrapResponseError(err error, resp serverResponse, object, id string) error {
	switch {
	case err == nil:
		return nil
	case resp.statusCode == http.StatusNotFound:
		return objectNotFoundError{object: object, id: id}
	case resp.statusCode == http.StatusNotImplemented:
		return errdefs.NotImplemented(err)
	default:
		return err
	}
}
