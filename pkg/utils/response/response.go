package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/pagination"
	"github.com/adsum-project/attendance-backend/pkg/utils/validation"
)

func emptySliceIfNil(v any) any {
	if v == nil {
		return []any{}
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		return reflect.MakeSlice(reflect.SliceOf(rv.Type().Elem()), 0, 0).Interface()
	}
	return v
}

type HttpErrorMessage struct {
	StatusCode int    `json:"-"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
	Details    any    `json:"details,omitempty"`
}

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

func JsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func JsonError(w http.ResponseWriter, err any) {
	var msg HttpErrorMessage
	switch e := err.(type) {
	case HttpErrorMessage:
		msg = e
	case *HttpErrorMessage:
		msg = *e
	case error:
		var vErr validation.ValidationError
		if errors.As(e, &vErr) {
			msg = HttpErrorMessage{
				StatusCode: vErr.StatusCode(),
				Success:    false,
				Error:      "There are one or more validation errors",
				Details:    vErr.Details(),
			}
		} else {
			msg = HttpErrorMessage{
				StatusCode: http.StatusInternalServerError,
				Success:    false,
				Error:      e.Error(),
			}
			var hErr errs.HTTPError
			if errors.As(e, &hErr) {
				msg.StatusCode = hErr.Code
			}
		}
	case string:
		msg = HttpErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Error:      e,
		}
	default:
		msg = HttpErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Error:      "An unexpected error occurred",
		}
	}
	if msg.StatusCode == 0 {
		msg.StatusCode = http.StatusInternalServerError
	}
	msg.Success = false
	JsonResponse(w, msg.StatusCode, msg)
}

func JsonSuccess(w http.ResponseWriter, statusCode int, message string, data any) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	resp := ApiResponse{
		Success: true,
		Message: message,
		Data:    emptySliceIfNil(data),
	}

	JsonResponse(w, statusCode, resp)
}

func JsonSuccessWithMeta(w http.ResponseWriter, statusCode int, message string, data any, meta any) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	resp := ApiResponse{
		Success: true,
		Message: message,
		Data:    emptySliceIfNil(data),
		Meta:    meta,
	}

	JsonResponse(w, statusCode, resp)
}

func OK(w http.ResponseWriter, message string, data any) {
	JsonSuccess(w, http.StatusOK, message, data)
}

func Created(w http.ResponseWriter, message string, data any) {
	JsonSuccess(w, http.StatusCreated, message, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, message string, details any) {
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusBadRequest,
		Error:      message,
		Details:    details,
	})
}

func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusUnauthorized,
		Error:      message,
	})
}

func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusForbidden,
		Error:      message,
	})
}

func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusNotFound,
		Error:      message,
	})
}

func MethodNotAllowed(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Method not allowed"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusMethodNotAllowed,
		Error:      message,
	})
}

func Conflict(w http.ResponseWriter, message string, details any) {
	if message == "" {
		message = "Conflict"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusConflict,
		Error:      message,
		Details:    details,
	})
}

func UnprocessableEntity(w http.ResponseWriter, message string, details any) {
	if message == "" {
		message = "Unprocessable entity"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusUnprocessableEntity,
		Error:      message,
		Details:    details,
	})
}

func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusInternalServerError,
		Error:      message,
	})
}

func BadGateway(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Bad gateway"
	}
	JsonError(w, HttpErrorMessage{
		StatusCode: http.StatusBadGateway,
		Error:      message,
	})
}

func PaginatedResponse(w http.ResponseWriter, message string, data any, page, perPage, total int) {
	totalPages := (total + perPage - 1) / perPage
	meta := PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
	JsonSuccessWithMeta(w, http.StatusOK, message, data, meta)
}

// PaginatedResponseFromResult writes a paginated JSON response from pagination.Result.
func PaginatedResponseFromResult[T any](w http.ResponseWriter, message string, result *pagination.Result[T]) {
	PaginatedResponse(w, message, result.Data, result.Page, result.PerPage, result.Total)
}
