package jobs

import (
	"context"
	"log"
	"time"

	"meta-targeting-copilot/backend/internal/db"
	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/pipeline"
)

type Worker struct {
	store    *db.Store
	runner   *pipeline.Runner
	interval time.Duration
}

func New(store *db.Store, runner *pipeline.Runner) *Worker {
	return &Worker{store: store, runner: runner, interval: 1500 * time.Millisecond}
}

func (w *Worker) Run(ctx context.Context) {
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	jobs, err := w.store.ClaimJobs(ctx, 2)
	if err != nil {
		log.Printf("job claim: %v", err)
		return
	}
	for _, j := range jobs {
		w.process(ctx, j)
	}
}

func (w *Worker) process(ctx context.Context, j db.Job) {
	if j.BriefID == nil {
		_ = w.store.FailJob(ctx, j.ID, "brief id missing")
		return
	}
	project, err := w.store.GetProject(ctx, j.ProjectID)
	if err != nil {
		_ = w.store.FailJob(ctx, j.ID, err.Error())
		return
	}
	brief, err := w.store.GetBrief(ctx, *j.BriefID)
	if err != nil {
		_ = w.store.FailJob(ctx, j.ID, err.Error())
		return
	}
	in := domain.BriefInput{
		Product: project.Product, AudiencePrompt: brief.Prompt,
		Location: project.City + ", " + project.Country,
		Country:  project.Country, City: project.City, Website: project.Website,
		Price: project.Price, Objective: project.Objective, Budget: project.Budget,
		Assets: project.Assets,
	}
	res, err := w.runner.Run(ctx, in, func(stage domain.JobStage) {
		_ = w.store.UpdateJobStage(ctx, j.ID, string(stage), "running", nil)
	})
	if err != nil {
		msg := err.Error()
		_ = w.store.FailJob(ctx, j.ID, msg)
		return
	}
	if err := w.store.PersistResult(ctx, project, brief.ID, res); err != nil {
		_ = w.store.FailJob(ctx, j.ID, err.Error())
		return
	}
	_ = w.store.CompleteJob(ctx, j.ID, map[string]any{"ok": true})
}
