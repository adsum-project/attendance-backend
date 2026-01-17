package handlers

import (
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	response.NotFound(w, "Resource Not Found")
}