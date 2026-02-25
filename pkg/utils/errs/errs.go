package errs

import "net/http"

type HTTPError struct {
	Code int
	Msg  string
}

func (e HTTPError) Error() string {
	return e.Msg
}

func (e HTTPError) StatusCode() int {
	return e.Code
}

func Error(code int, msg string) HTTPError {
	if msg == "" {
		msg = "Error"
	}
	return HTTPError{Code: code, Msg: msg}
}

func BadRequest(msg string) HTTPError {
	if msg == "" {
		msg = "Bad request"
	}
	return Error(http.StatusBadRequest, msg)
}

func Unauthorized(msg string) HTTPError {
	if msg == "" {
		msg = "Unauthorized"
	}
	return Error(http.StatusUnauthorized, msg)
}

func Forbidden(msg string) HTTPError {
	if msg == "" {
		msg = "Forbidden"
	}
	return Error(http.StatusForbidden, msg)
}

func NotFound(msg string) HTTPError {
	if msg == "" {
		msg = "Resource not found"
	}
	return Error(http.StatusNotFound, msg)
}

func MethodNotAllowed(msg string) HTTPError {
	if msg == "" {
		msg = "Method not allowed"
	}
	return Error(http.StatusMethodNotAllowed, msg)
}

func Conflict(msg string) HTTPError {
	if msg == "" {
		msg = "Conflict"
	}
	return Error(http.StatusConflict, msg)
}

func UnprocessableEntity(msg string) HTTPError {
	if msg == "" {
		msg = "Unprocessable entity"
	}
	return Error(http.StatusUnprocessableEntity, msg)
}

func InternalServerError(msg string) HTTPError {
	if msg == "" {
		msg = "Internal server error"
	}
	return Error(http.StatusInternalServerError, msg)
}

func BadGateway(msg string) HTTPError {
	if msg == "" {
		msg = "Bad gateway"
	}
	return Error(http.StatusBadGateway, msg)
}
