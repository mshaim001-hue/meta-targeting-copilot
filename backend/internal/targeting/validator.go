package targeting

import (
	"fmt"
	"strings"

	"meta-targeting-copilot/backend/internal/domain"
)

// StripHallucinations removes any targeting that looks like Meta targeting
// but was not produced by the Meta resolver.
func StripHallucinations(entities []domain.TargetingEntity) []domain.TargetingEntity {
	var out []domain.TargetingEntity
	for _, e := range entities {
		if e.MetaID != "" && e.Source != domain.SourceMetaVerified && e.Source != domain.SourceMetaSuggestion {
			e.MetaID = ""
			e.MetaName = ""
			e.Status = domain.StatusUnavailable
			e.Source = domain.SourceNotAvailable
			e.Explanation = "Rejected: Meta IDs cannot originate from the LLM."
		}
		if (e.Status == domain.StatusVerified || e.Source == domain.SourceMetaVerified) && e.MetaID == "" {
			e.Status = domain.StatusUnavailable
			e.Source = domain.SourceNotAvailable
			e.Explanation = "Rejected: verified status requires a Meta ID from the API."
		}
		out = append(out, e)
	}
	return out
}

func DedupSegments(segs []domain.AudienceSegment) []domain.AudienceSegment {
	seen := map[string]struct{}{}
	var out []domain.AudienceSegment
	for _, s := range segs {
		key := strings.ToLower(strings.TrimSpace(s.Name))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
}

func Score(seg domain.AudienceSegment, ents []domain.TargetingEntity, assets domain.FirstPartyAssets, reach domain.ReachEstimate) domain.QualityScore {
	verified, suggested := 0, 0
	for _, e := range ents {
		if seg.Name != "" && e.GeneratedFrom != "" && e.GeneratedFrom != seg.Name {
			continue
		}
		switch e.Status {
		case domain.StatusVerified:
			verified++
		case domain.StatusSuggested:
			suggested++
		}
	}
	signals := len(seg.SemanticSignals)
	if signals == 0 {
		signals = 1
	}
	avail := int(float64(verified+suggested) / float64(signals) * 100)
	if avail > 100 {
		avail = 100
	}
	intent := 55
	switch seg.IntentLevel {
	case domain.IntentHigh:
		intent = 88
	case domain.IntentMedium:
		intent = 70
	}
	fp := 0
	if assets.Pixel || assets.WebsiteVisitors {
		fp += 30
	}
	if assets.CustomerList || assets.Purchasers || assets.CRM {
		fp += 40
	}
	if assets.Instagram || assets.FacebookPage || assets.VideoViews || assets.LeadForms {
		fp += 20
	}
	if fp > 100 {
		fp = 100
	}
	size := 50
	if reach.Available && reach.LowerBound != nil {
		switch {
		case *reach.LowerBound >= 100000:
			size = 90
		case *reach.LowerBound >= 20000:
			size = 76
		case *reach.LowerBound >= 5000:
			size = 60
		default:
			size = 35
		}
	}
	rel := 70
	if verified > 0 {
		rel = 90
	}
	total := (rel + avail + size + intent + fp) / 5
	return domain.QualityScore{
		SemanticRelevance: rel,
		MetaAvailability:  avail,
		AudienceSize:      size,
		IntentStrength:    intent,
		FirstPartySignal:  fp,
		Total:             total,
	}
}

func ValidateSpecEntities(ents []domain.TargetingEntity) error {
	for _, e := range ents {
		if e.Status == domain.StatusVerified && e.MetaID == "" {
			return fmt.Errorf("verified entity %q is missing meta id", e.Query)
		}
	}
	return nil
}
