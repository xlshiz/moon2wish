package api

import (
	"encoding/json"
	"net/http"
)

// JSON 响应的快捷函数
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data any) {
	jsonResponse(w, http.StatusOK, data)
}
