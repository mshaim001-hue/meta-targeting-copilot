package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"meta-targeting-copilot/backend/internal/config"
	"meta-targeting-copilot/backend/internal/db"
	"meta-targeting-copilot/backend/internal/meta"
	"meta-targeting-copilot/backend/internal/pipeline"
)

type Server struct {
	cfg     config.Config
	store   *db.Store
	run     *pipeline.Runner
	meta    meta.Provider
	llmMode string
}

func New(cfg config.Config, store *db.Store, run *pipeline.Runner, metaProv meta.Provider, llmMode string) http.Handler {
	s := &Server{cfg: cfg, store: store, run: run, meta: metaProv, llmMode: llmMode}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendOrigin, "http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "llm": llmMode, "meta": metaProv.Name(), "metaConnected": metaProv.Connected(),
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", s.register)
		r.Post("/auth/login", s.login)

		r.Group(func(r chi.Router) {
			r.Use(s.auth)
			r.Get("/auth/me", s.me)
			r.Post("/projects", s.createProject)
			r.Get("/projects", s.listProjects)
			r.Get("/projects/{id}", s.getProject)
			r.Put("/projects/{id}", s.updateProject)
			r.Post("/projects/{id}/brief", s.createBrief)
			r.Post("/projects/{id}/analyze", s.analyze)
			r.Get("/projects/{id}/jobs/{jobId}", s.getJob)
			r.Get("/projects/{id}/job", s.latestJob)
			r.Get("/projects/{id}/segments", s.getSegments)
			r.Get("/projects/{id}/strategy", s.getStrategy)
			r.Post("/projects/{id}/strategy/regenerate", s.analyze)
			r.Get("/projects/{id}/export", s.exportStrategy)
			r.Post("/meta/interests/search", s.searchInterests)
			r.Post("/meta/geo/search", s.searchGeo)
			r.Post("/meta/audience/estimate", s.estimateAudience)
		})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dest any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dest)
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}
