package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moon-readwise-server/internal/storage"
)

func NewRouter(store *storage.SQLiteStore, token string, debug bool) *chi.Mux {
	r := chi.NewRouter()

	// 基础中间件
	if debug {
		r.Use(SecurityHeadersMiddleware())
	}

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/moon", func(r chi.Router) {
		// 认证中间件
		r.Use(AuthMiddleware(token))

		// V2 API - Readwise 兼容接口
		r.Route("/v2", func(r chi.Router) {
			r.Route("/highlights", func(r chi.Router) {
				r.Post("/", CreateHighlights(store))
				r.Get("/", ListHighlights(store))

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", GetHighlight(store))
					r.Patch("/", UpdateHighlight(store))
					r.Delete("/", DeleteHighlight(store))
				})

				// 删除所有高亮（注意路径冲突，放在最后）
				r.Delete("/", DeleteAllHighlights(store))
			})

			r.Route("/export", func(r chi.Router) {
				r.Get("/", ExportJSON(store))
			})

			r.Route("/books", func(r chi.Router) {
				r.Get("/", ListBooks(store))
				r.Get("/{id}", GetBook(store))
			})
		})

		// 通用导出接口
		r.Get("/export/json", ExportJSON(store))
		r.Get("/export/csv", ExportCSV(store))
	})

	return r
}
