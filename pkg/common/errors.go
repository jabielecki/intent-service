package common

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// MultiError implements errors with multiple causes
type MultiError []error

// Error implements default errors interface for MultiError.
func (m MultiError) Error() string {
	var msgs []string
	for _, e := range m {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

//ErrorNotFound for not found error.
var ErrorNotFound = grpc.Errorf(codes.NotFound, "not found")

//ErrorUnauthenticated for unauthenticated error.
var ErrorUnauthenticated = grpc.Errorf(codes.Unauthenticated, "Unauthenticated")

//ErrorPermissionDenied for permission denied errror.
var ErrorPermissionDenied = grpc.Errorf(codes.PermissionDenied, "Permission Denied")

//ErrorInternal for Internal Server Error.
var ErrorInternal = grpc.Errorf(codes.Internal, "Internal Server Error")

//ErrorConflict is for resource conflict error.
var ErrorConflict = grpc.Errorf(codes.AlreadyExists, "Resource conflict")

// IsNotFound returns true if error is of NotFound type.
func IsNotFound(err error) bool {
	return grpc.Code(errors.Cause(err)) == codes.NotFound
}

// IsConflict returns true if error is of Conflict type.
func IsConflict(err error) bool {
	return grpc.Code(errors.Cause(err)) == codes.AlreadyExists
}

//ErrorForbiddenf makes forbidden error with format.
func ErrorForbiddenf(format string, a ...interface{}) error {
	return grpc.Errorf(codes.PermissionDenied, format, a)
}

//ErrorForbidden makes forbidden error.
func ErrorForbidden(message string) error {
	if message == "" {
		message = "forbidden error"
	}
	return ErrorForbiddenf(message)
}

//ErrorBadRequestf makes bad request error with format.
func ErrorBadRequestf(format string, a ...interface{}) error {
	return grpc.Errorf(codes.InvalidArgument, format, a)
}

//ErrorBadRequest makes bad request error.
func ErrorBadRequest(message string) error {
	if message == "" {
		message = "bad request error"
	}
	return ErrorBadRequestf(message)
}

//ErrorNotFoundf makes not found error.
func ErrorNotFoundf(message string, a ...interface{}) error {
	if message == "" {
		message = "not found"
	}
	return grpc.Errorf(codes.NotFound, message, a...)
}

// ErrorConflictf makes not found error.
func ErrorConflictf(format string, a ...interface{}) error {
	if format == "" {
		return ErrorConflict
	}
	return grpc.Errorf(codes.AlreadyExists, format, a...)
}

// HTTPStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// nolint: gocyclo
func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusForbidden
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

// ToHTTPError translates grpc error to error.
func ToHTTPError(err error) error {
	cause := errors.Cause(err)
	code := HTTPStatusFromCode(grpc.Code(cause))
	return echo.NewHTTPError(code, grpc.ErrorDesc(cause))
}
