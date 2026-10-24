package targeting

import (
	"testing"

	"meta-targeting-copilot/backend/internal/domain"
)

func TestDedupSegments(t *testing.T) {
	out := DedupSegments([]domain.AudienceSegment{
		{Name: "Regular Gym Users"},
		{Name: "regular gym users"},
		{Name: "Runners"},
	})
	if len(out) != 2 {
		t.Fatalf("got %d", len(out))
	}
}

func TestOnlyVerifiedRequiresMetaID(t *testing.T) {
	got := OnlyVerified([]domain.TargetingEntity{
		{Query: "x", Status: domain.StatusVerified, Source: domain.SourceMetaVerified},
		{Query: "y", MetaID: "1", Status: domain.StatusVerified, Source: domain.SourceMetaVerified},
		{Query: "z", MetaID: "2", Status: domain.StatusSuggested, Source: domain.SourceMetaSuggestion},
	})
	if len(got) != 1 || got[0].Query != "y" {
		t.Fatalf("%+v", got)
	}
}
