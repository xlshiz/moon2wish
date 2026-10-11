package api

import (
	"net/http"
	"strings"
)

// SecurityHeadersMiddleware 添加安全HTTP头
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Content-Security-Policy", "default-src 'none'")
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateContentTypeMiddleware 验证Content-Type
func ValidateContentTypeMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 只对需要JSON的POST/PUT/PATCH请求进行验证
			if r.Method == http.MethodPost ||
				r.Method == http.MethodPut ||
				r.Method == http.MethodPatch {

				contentType := r.Header.Get("Content-Type")
				if contentType != "" && !strings.Contains(contentType, "application/json") {
					jsonError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
