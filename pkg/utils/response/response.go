package response

import (
	"encoding/json"
	"net/http"
)

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
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func JsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func JsonError(w http.ResponseWriter, err any) {
	var errorMsg HttpErrorMessage

	switch e := err.(type) {
	case HttpErrorMessage:
		errorMsg = e
	case *HttpErrorMessage:
		errorMsg = *e
	case error:
		errorMsg = HttpErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Error:      e.Error(),
		}
	case string:
		errorMsg = HttpErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Error:      e,
		}
	default:
		errorMsg = HttpErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Error:      "An unexpected error occurred",
		}
	}

	if errorMsg.StatusCode == 0 {
		errorMsg.StatusCode = http.StatusInternalServerError
	}

	errorMsg.Success = false

	JsonResponse(w, errorMsg.StatusCode, errorMsg)
}

func JsonSuccess(w http.ResponseWriter, statusCode int, message string, data any) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	resp := ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
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
		Data:    data,
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
