package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"meta-targeting-copilot/backend/internal/auth"
	"meta-targeting-copilot/backend/internal/db"
	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/meta"
	"meta-targeting-copilot/backend/internal/strategy"
)

type ctxKey int

const userKey ctxKey = 1

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := bearer(r)
		if tok == "" {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		claims, err := auth.Parse(s.cfg.JWTSecret, tok)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), userKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func currentUser(r *http.Request) *auth.Claims {
	c, _ := r.Context().Value(userKey).(*auth.Claims)
	return c
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, 400, "invalid json")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Email == "" || len(in.Password) < 8 {
		writeError(w, 400, "email and password (min 8 chars) required")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		writeError(w, 500, "hash failed")
		return
	}
	u, err := s.store.CreateUser(r.Context(), in.Email, hash)
	if err != nil {
		writeError(w, 409, "email already registered")
		return
	}
	s.issue(w, u)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, 400, "invalid json")
		return
	}
	u, err := s.store.GetUserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil || auth.CheckPassword(u.PasswordHash, in.Password) != nil {
		writeError(w, 401, "invalid credentials")
		return
	}
	s.issue(w, u)
}

func (s *Server) issue(w http.ResponseWriter, u db.User) {
	tok, err := auth.Sign(s.cfg.JWTSecret, u.ID, u.Email, 7*24*time.Hour)
	if err != nil {
		writeError(w, 500, "token failed")
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok, "user": map[string]any{"id": u.ID, "email": u.Email}})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	writeJSON(w, 200, map[string]any{"id": c.UserID, "email": c.Email})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	p, err := decodeProject(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	p.UserID = currentUser(r).UserID
	created, err := s.store.CreateProject(r.Context(), p)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, created)
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListProjects(r.Context(), currentUser(r).UserID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []domain.Project{}
	}
	writeJSON(w, 200, list)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, ok := s.ownedProject(w, r)
	if !ok {
		return
	}
	writeJSON(w, 200, p)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	existing, ok := s.ownedProject(w, r)
	if !ok {
		return
	}
	p, err := decodeProject(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	p.ID = existing.ID
	p.UserID = existing.UserID
	updated, err := s.store.UpdateProject(r.Context(), p)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, updated)
}

func (s *Server) createBrief(w http.ResponseWriter, r *http.Request) {
	p, ok := s.ownedProject(w, r)
	if !ok {
		return
	}
	var in struct {
		Prompt string `json:"prompt"`
	}
	if err := decodeJSON(r, &in); err != nil || strings.TrimSpace(in.Prompt) == "" {
		writeError(w, 400, "prompt required")
		return
	}
	b, err := s.store.CreateBrief(r.Context(), p.ID, strings.TrimSpace(in.Prompt))
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, b)
}

func (s *Server) analyze(w http.ResponseWriter, r *http.Request) {
	p, ok := s.ownedProject(w, r)
	if !ok {
		return
	}
	var in struct {
		Prompt string `json:"prompt"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	var brief db.Brief
	var err error
	if strings.TrimSpace(in.Prompt) != "" {
		brief, err = s.store.CreateBrief(r.Context(), p.ID, strings.TrimSpace(in.Prompt))
	} else {
		brief, err = s.store.LatestBrief(r.Context(), p.ID)
	}
	if err != nil {
		writeError(w, 400, "create a brief first")
		return
	}
	job, err := s.store.EnqueueJob(r.Context(), p.ID, brief.ID, "analyze")
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 202, job)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	p, ok := s.ownedProject(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "jobId"))
	if err != nil {
		writeError(w, 400, "bad job id")
		return
	}
	j, err := s.store.GetJob(r.Context(), id)
	if err != nil || j.ProjectID != p.ID {
		writeError(w, 404, "job not found")
		return
	}
	writeJSON(w, 200, j)
}

func (s *Server) latestJob(w http.ResponseWriter, r *http.Request) {
	p, ok := s.ownedProject(w, r)
	if !ok {
		return
	}
	j, err := s.store.LatestJob(r.Context(), p.ID)
	if err != nil {
		writeError(w, 404, "no job")
		return
	}
	writeJSON(w, 200, j)
}

func (s *Server) getSegments(w http.ResponseWriter, r *http.Request) {
	res, ok := s.loaded(w, r)
	if !ok {
		return
	}
	writeJSON(w, 200, res.Segments)
}

func (s *Server) getStrategy(w http.ResponseWriter, r *http.Request) {
	res, ok := s.loaded(w, r)
	if !ok {
		return
	}
	writeJSON(w, 200, res)
}

func (s *Server) exportStrategy(w http.ResponseWriter, r *http.Request) {
	res, ok := s.loaded(w, r)
	if !ok {
		return
	}
	format := r.URL.Query().Get("format")
	if format == "md" || format == "markdown" {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=strategy.md")
		_, _ = w.Write([]byte(strategy.Markdown(res)))
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=strategy.json")
	writeJSON(w, 200, res)
}

func (s *Server) searchInterests(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Query string `json:"query"`
	}
	if err := decodeJSON(r, &in); err != nil || in.Query == "" {
		writeError(w, 400, "query required")
		return
	}
	hits, err := s.meta.SearchInterests(r.Context(), in.Query)
	if err != nil {
		writeError(w, 502, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": hits, "provider": s.meta.Name(), "connected": s.meta.Connected()})
}

func (s *Server) searchGeo(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Query string `json:"query"`
	}
	if err := decodeJSON(r, &in); err != nil || in.Query == "" {
		writeError(w, 400, "query required")
		return
	}
	hits, err := s.meta.SearchLocations(r.Context(), in.Query)
	if err != nil {
		writeError(w, 502, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": hits, "provider": s.meta.Name(), "connected": s.meta.Connected()})
}

func (s *Server) estimateAudience(w http.ResponseWriter, r *http.Request) {
	var spec meta.TargetingSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeError(w, 400, "targeting spec required")
		return
	}
	est, err := s.meta.EstimateAudience(r.Context(), spec)
	if err != nil {
		writeJSON(w, 200, domain.ReachEstimate{Available: false, Message: "Estimate unavailable"})
		return
	}
	if !est.Available {
		writeJSON(w, 200, domain.ReachEstimate{Available: false, Message: "Estimate unavailable"})
		return
	}
	writeJSON(w, 200, domain.ReachEstimate{Available: true, LowerBound: est.LowerBound, UpperBound: est.UpperBound, Message: est.Message})
}

func (s *Server) ownedProject(w http.ResponseWriter, r *http.Request) (domain.Project, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 400, "bad id")
		return domain.Project{}, false
	}
	p, err := s.store.GetProject(r.Context(), id)
	if err != nil {
		writeError(w, 404, "project not found")
		return domain.Project{}, false
	}
	if p.UserID != currentUser(r).UserID {
		writeError(w, 404, "project not found")
		return domain.Project{}, false
	}
	return p, true
}

func (s *Server) loaded(w http.ResponseWriter, r *http.Request) (domain.PipelineResult, bool) {
	p, ok := s.ownedProject(w, r)
	if !ok {
		return domain.PipelineResult{}, false
	}
	brief, err := s.store.LatestBrief(r.Context(), p.ID)
	if err != nil {
		writeError(w, 404, "no brief yet")
		return domain.PipelineResult{}, false
	}
	res, err := s.store.LoadResult(r.Context(), p, brief)
	if err != nil {
		writeError(w, 500, err.Error())
		return domain.PipelineResult{}, false
	}
	if len(res.Segments) == 0 {
		writeError(w, 404, "strategy not ready")
		return domain.PipelineResult{}, false
	}
	res.LLMMode = s.llmMode
	res.MetaMode = s.meta.Name()
	return res, true
}

func decodeProject(r *http.Request) (domain.Project, error) {
	var in struct {
		BusinessName string                  `json:"businessName"`
		Website      string                  `json:"website"`
		Product      string                  `json:"product"`
		Country      string                  `json:"country"`
		City         string                  `json:"city"`
		Price        string                  `json:"price"`
		Objective    string                  `json:"objective"`
		Budget       string                  `json:"budget"`
		Assets       domain.FirstPartyAssets `json:"assets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return domain.Project{}, errors.New("invalid json")
	}
	if strings.TrimSpace(in.BusinessName) == "" || strings.TrimSpace(in.Product) == "" || strings.TrimSpace(in.Country) == "" || strings.TrimSpace(in.City) == "" {
		return domain.Project{}, errors.New("businessName, product, country, city required")
	}
	return domain.Project{
		BusinessName: in.BusinessName, Website: in.Website, Product: in.Product,
		Country: in.Country, City: in.City, Price: in.Price, Objective: in.Objective, Budget: in.Budget,
		Assets: in.Assets,
	}, nil
}
