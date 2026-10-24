package pipeline

import (
	"context"
	"testing"

	"meta-targeting-copilot/backend/internal/ai"
	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/meta"
	"meta-targeting-copilot/backend/internal/targeting"
)

func TestFitnessBriefPipeline(t *testing.T) {
	in := domain.BriefInput{
		Product:        "Fitness club",
		AudiencePrompt: "У меня фитнес-клуб в Астане. Хочу привлечь людей 25–40 лет, которые уже занимаются спортом и думают поменять зал.",
		Location:       "Astana, Kazakhstan",
		Country:        "Kazakhstan",
		City:           "Astana",
		Objective:      "Membership",
		Assets:         domain.FirstPartyAssets{Pixel: true, Instagram: true},
	}
	res, err := New(ai.NewEngine(ai.NewMock()), targeting.NewResolver(meta.NewMock())).Run(context.Background(), in, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Segments) < 3 {
		t.Fatalf("need >=3 segments, got %d", len(res.Segments))
	}
	types := map[domain.StrategyType]bool{}
	for _, s := range res.Strategies {
		types[s.Type] = true
		for _, e := range s.IncludedTargeting {
			if e.Status == domain.StatusVerified && e.MetaID == "" {
				t.Fatalf("strategy %s leaked unverified entity %q", s.Name, e.Query)
			}
		}
	}
	if !types[domain.StrategyBroad] || !types[domain.StrategyInterest] {
		t.Fatalf("missing required strategies: %+v", types)
	}
	if !res.Geo.Verified {
		t.Fatalf("geo not verified: %+v", res.Geo)
	}
	if len(res.Offers) < 3 || len(res.Creatives) < 3 {
		t.Fatalf("offers=%d creatives=%d", len(res.Offers), len(res.Creatives))
	}
	for _, e := range res.Entities {
		if e.Source == domain.SourceMetaVerified && e.MetaID == "" {
			t.Fatalf("hallucinated verified entity: %+v", e)
		}
	}
}

func TestUnknownSignalStaysUnavailable(t *testing.T) {
	r := targeting.NewResolver(meta.NewMock())
	ents := r.ResolveSignals(context.Background(), []domain.AudienceSegment{{
		Name: "X", SemanticSignals: []string{"people who recently got a mortgage and now hate contractors"},
	}})
	if ents[0].Status == domain.StatusVerified {
		t.Fatal("unsearchable characteristic must not become a Meta interest")
	}
}
