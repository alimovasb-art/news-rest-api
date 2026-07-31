package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   any    `json:"error"`
	Data    any    `json:"data"`
}

func SendSuccess(w http.ResponseWriter, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := APIResponse{
		Success: true,
		Code:    code,
		Message: message,
		Error:   nil,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func SendError(w http.ResponseWriter, code int, massage string, errorDetail any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := APIResponse{
		Success: false,
		Code:    code,
		Message: massage,
		Error:   errorDetail,
		Data:    nil,
	}

	json.NewEncoder(w).Encode(response)
}

func GetPaginationParams(r *http.Request) (int, int) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}
