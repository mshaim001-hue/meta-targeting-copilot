package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"meta-targeting-copilot/backend/internal/ai"
	"meta-targeting-copilot/backend/internal/api"
	"meta-targeting-copilot/backend/internal/config"
	"meta-targeting-copilot/backend/internal/db"
	"meta-targeting-copilot/backend/internal/jobs"
	"meta-targeting-copilot/backend/internal/meta"
	"meta-targeting-copilot/backend/internal/pipeline"
	"meta-targeting-copilot/backend/internal/targeting"
)

func main() {
	cfg := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	migDir := os.Getenv("MIGRATIONS_DIR")
	if migDir == "" {
		migDir = findMigrations()
	}
	if err := db.Migrate(ctx, pool, migDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	store := db.NewStore(pool)

	var llm ai.Provider
	if cfg.LLMAPIKey != "" {
		llm = ai.NewOpenAI(cfg.LLMAPIKey, cfg.LLMBaseURL, cfg.LLMModel)
		log.Printf("llm: openai-compatible (%s)", cfg.LLMModel)
	} else {
		llm = ai.NewMock()
		log.Printf("llm: mock (set LLM_API_KEY to enable a real provider)")
	}

	var metaProv meta.Provider
	if cfg.MetaAccessToken != "" {
		metaProv = meta.NewGraph(cfg.MetaAccessToken, cfg.MetaAdAccountID, cfg.MetaAPIVersion)
		log.Printf("meta: graph %s", cfg.MetaAPIVersion)
	} else {
		metaProv = meta.NewMock()
		log.Printf("meta: mock catalog (set META_ACCESS_TOKEN to verify against Graph API)")
	}
	metaProv = meta.WithCache(metaProv, store, cfg.CacheTTL, cfg.MetaAPIVersion, "")

	runner := pipeline.New(ai.NewEngine(llm), targeting.NewResolver(metaProv))
	go jobs.New(store, runner).Run(ctx)

	handler := api.New(cfg, store, runner, metaProv, llm.Name())
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: handler}
	go func() {
		log.Printf("api listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()
	<-ctx.Done()
	shutdown, done := context.WithTimeout(context.Background(), 8*time.Second)
	defer done()
	_ = srv.Shutdown(shutdown)
}

func findMigrations() string {
	candidates := []string{"migrations", "backend/migrations", filepath.Join("..", "..", "migrations")}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return "migrations"
}
