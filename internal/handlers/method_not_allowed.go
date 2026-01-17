package handlers

import (
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	response.MethodNotAllowed(w, "Method Not Allowed")
}