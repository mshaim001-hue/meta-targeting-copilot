package targeting

import (
	"context"
	"strings"

	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/meta"
)

type Resolver struct {
	meta meta.Provider
}

func NewResolver(p meta.Provider) *Resolver {
	return &Resolver{meta: p}
}

func (r *Resolver) Mode() string {
	if r == nil || r.meta == nil {
		return "none"
	}
	return r.meta.Name()
}

func (r *Resolver) Connected() bool {
	return r.meta != nil && r.meta.Connected()
}

// ResolveSignals never trusts LLM names as Meta entities.
// A signal stays llm_hypothesis / not_available until Meta returns a match.
func (r *Resolver) ResolveSignals(ctx context.Context, segs []domain.AudienceSegment) []domain.TargetingEntity {
	var out []domain.TargetingEntity
	for _, seg := range segs {
		for _, signal := range seg.SemanticSignals {
			ent := r.resolveOne(ctx, signal, seg.Name)
			out = append(out, ent)
		}
	}
	return out
}

func (r *Resolver) resolveOne(ctx context.Context, signal, segment string) domain.TargetingEntity {
	base := domain.TargetingEntity{
		Query:         signal,
		GeneratedFrom: segment,
		Status:        domain.StatusUnavailable,
		Source:        domain.SourceLLMHypothesis,
		Explanation:   "Marketing signal. Not yet confirmed by Meta API.",
	}
	if r.meta == nil {
		base.Source = domain.SourceNotAvailable
		base.Explanation = "Meta provider is not configured."
		return base
	}
	hits, err := r.meta.SearchInterests(ctx, signal)
	if err != nil {
		base.Source = domain.SourceNotAvailable
		base.Explanation = "Meta verification temporarily unavailable: " + err.Error()
		return base
	}
	if len(hits) == 0 {
		base.Source = domain.SourceNotAvailable
		base.Explanation = "No direct targeting entity found. Prefer Broad + creative qualification."
		return base
	}
	best := pickBest(signal, hits)
	if best.Restricted {
		return domain.TargetingEntity{
			Query: signal, MetaID: best.ID, MetaName: best.Name, MetaType: best.Type,
			Status: domain.StatusRestricted, Source: domain.SourceMetaVerified,
			GeneratedFrom: segment, AudienceSize: best.AudienceSize,
			Explanation: "Meta returned this entity but it is restricted/unsupported for ads.",
		}
	}
	status := domain.StatusVerified
	source := domain.SourceMetaVerified
	expl := "Confirmed by Meta Targeting Search."
	if !closeMatch(signal, best.Name) {
		status = domain.StatusSuggested
		source = domain.SourceMetaSuggestion
		expl = "Meta returned a related entity, not an exact match for the marketing signal."
	}
	return domain.TargetingEntity{
		Query: signal, MetaID: best.ID, MetaName: best.Name, MetaType: best.Type,
		Status: status, Source: source, GeneratedFrom: segment,
		AudienceSize: best.AudienceSize, Explanation: expl,
	}
}

func (r *Resolver) ResolveGeo(ctx context.Context, rec domain.GeoRecommendation) domain.GeoRecommendation {
	if r.meta == nil {
		return rec
	}
	q := strings.TrimSpace(rec.City)
	if q == "" {
		q = rec.Country
	}
	locs, err := r.meta.SearchLocations(ctx, q)
	if err != nil || len(locs) == 0 {
		if rec.Country != "" {
			locs, err = r.meta.SearchLocations(ctx, rec.Country)
		}
	}
	if err != nil || len(locs) == 0 {
		rec.Verified = false
		return rec
	}
	city := firstOfType(locs, "city")
	country := firstOfType(locs, "country")
	if city != nil {
		rec.City = city.Name
		rec.CityKey = city.Key
		rec.LocationType = "city"
		rec.Latitude = city.Latitude
		rec.Longitude = city.Longitude
		if city.CountryCode != "" {
			rec.CountryKey = city.CountryCode
		}
		rec.Verified = true
	}
	if country != nil {
		rec.Country = country.Name
		rec.CountryKey = country.Key
		if !rec.Verified {
			rec.LocationType = "country"
			rec.Verified = true
		}
	}
	if rec.CountryKey == "" && len(locs) > 0 && locs[0].CountryCode != "" {
		rec.CountryKey = locs[0].CountryCode
	}
	return rec
}

func (r *Resolver) Estimate(ctx context.Context, spec meta.TargetingSpec) domain.ReachEstimate {
	if r.meta == nil {
		return domain.ReachEstimate{Available: false, Message: "Estimate unavailable"}
	}
	est, err := r.meta.EstimateAudience(ctx, spec)
	if err != nil {
		return domain.ReachEstimate{Available: false, Message: "Estimate unavailable"}
	}
	if !est.Available {
		msg := est.Message
		if msg == "" {
			msg = "Estimate unavailable"
		}
		return domain.ReachEstimate{Available: false, Message: msg}
	}
	return domain.ReachEstimate{Available: true, LowerBound: est.LowerBound, UpperBound: est.UpperBound, Message: est.Message}
}

func pickBest(query string, hits []meta.Entity) meta.Entity {
	best := hits[0]
	bestScore := scoreName(query, best.Name)
	for _, h := range hits[1:] {
		s := scoreName(query, h.Name)
		if s > bestScore {
			best, bestScore = h, s
		}
	}
	return best
}

func scoreName(query, name string) int {
	q := strings.ToLower(strings.TrimSpace(query))
	n := strings.ToLower(strings.TrimSpace(name))
	if q == n {
		return 100
	}
	if strings.Contains(n, q) || strings.Contains(q, n) {
		return 80
	}
	return 10
}

func closeMatch(query, name string) bool {
	return scoreName(query, name) >= 80
}

func firstOfType(locs []meta.Location, typ string) *meta.Location {
	for i := range locs {
		if locs[i].Type == typ {
			return &locs[i]
		}
	}
	return nil
}

// OnlyVerified returns entities that may appear as available Meta targeting.
func OnlyVerified(in []domain.TargetingEntity) []domain.TargetingEntity {
	var out []domain.TargetingEntity
	for _, e := range in {
		if e.Status == domain.StatusVerified && e.MetaID != "" && e.Source == domain.SourceMetaVerified {
			out = append(out, e)
		}
	}
	return out
}

func VerifiedOrSuggested(in []domain.TargetingEntity) []domain.TargetingEntity {
	var out []domain.TargetingEntity
	for _, e := range in {
		if e.MetaID == "" {
			continue
		}
		if e.Status == domain.StatusVerified || e.Status == domain.StatusSuggested {
			out = append(out, e)
		}
	}
	return out
}
