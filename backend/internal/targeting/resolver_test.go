package targeting

import (
	"context"
	"testing"

	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/meta"
)

func TestStripHallucinations(t *testing.T) {
	in := []domain.TargetingEntity{
		{Query: "Bodybuilding", MetaID: "6001", Status: domain.StatusVerified, Source: domain.SourceLLMHypothesis},
		{Query: "gym", MetaID: "6002", Status: domain.StatusVerified, Source: domain.SourceMetaVerified},
		{Query: "serious gym users", Status: domain.StatusVerified, Source: domain.SourceMetaVerified},
	}
	out := StripHallucinations(in)
	if out[0].MetaID != "" || out[0].Status != domain.StatusUnavailable {
		t.Fatalf("llm-originated meta id must be stripped: %+v", out[0])
	}
	if out[1].MetaID != "6002" {
		t.Fatalf("real meta entity dropped")
	}
	if out[2].Status != domain.StatusUnavailable {
		t.Fatalf("verified without id must be rejected")
	}
}

func TestResolverDoesNotInventIDs(t *testing.T) {
	r := NewResolver(meta.NewMock())
	segs := []domain.AudienceSegment{{
		Name:            "Regular Gym Users",
		SemanticSignals: []string{"bodybuilding", "people who secretly hate their current gym"},
	}}
	ents := r.ResolveSignals(context.Background(), segs)
	if len(ents) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(ents))
	}
	if ents[0].Status != domain.StatusVerified || ents[0].MetaID == "" {
		t.Fatalf("bodybuilding should resolve: %+v", ents[0])
	}
	if ents[1].Status == domain.StatusVerified || ents[1].MetaID != "" {
		t.Fatalf("unknown signal must not become verified: %+v", ents[1])
	}
	for _, e := range OnlyVerified(ents) {
		if e.Source != domain.SourceMetaVerified {
			t.Fatalf("verified set leaked non-meta source: %+v", e)
		}
	}
}

func TestGeoAstana(t *testing.T) {
	r := NewResolver(meta.NewMock())
	geo := r.ResolveGeo(context.Background(), domain.GeoRecommendation{Country: "Kazakhstan", City: "Astana", RadiusKM: 5})
	if !geo.Verified || geo.CityKey == "" {
		t.Fatalf("astana should verify: %+v", geo)
	}
}
