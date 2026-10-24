package strategy

import (
	"context"
	"fmt"

	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/meta"
	"meta-targeting-copilot/backend/internal/targeting"
)

type Engine struct {
	resolver *targeting.Resolver
}

func NewEngine(r *targeting.Resolver) *Engine {
	return &Engine{resolver: r}
}

func (e *Engine) Build(ctx context.Context, in domain.BriefInput, segs []domain.AudienceSegment, ents []domain.TargetingEntity, geo domain.GeoRecommendation) []domain.AudienceStrategy {
	verified := targeting.OnlyVerified(ents)
	usable := targeting.VerifiedOrSuggested(ents)

	broad := domain.AudienceStrategy{
		Name:              "Broad / Advantage+",
		Type:              domain.StrategyBroad,
		Geo:               geo,
		Ages:              []int{geo.AgeMin, geo.AgeMax},
		Gender:            geo.Gender,
		IncludedTargeting: nil,
		ExcludedTargeting: exclusions(in.Assets, "cold"),
		Confidence:        0.74,
		ConfidenceLevel:   domain.ConfidenceMedium,
		Basis:             []string{"user brief", "marketing reasoning"},
		Explanation:       "Give Meta's delivery system room to find converters. Qualify the audience through creative and offer, not through stacked interests.",
	}

	interest := domain.AudienceStrategy{
		Name:              "Intent / Interests",
		Type:              domain.StrategyInterest,
		Geo:               geo,
		Ages:              []int{geo.AgeMin, geo.AgeMax},
		Gender:            geo.Gender,
		IncludedTargeting: uniqueByMetaID(verified),
		ExcludedTargeting: exclusions(in.Assets, "cold"),
		Confidence:        0.81,
		ConfidenceLevel:   domain.ConfidenceHigh,
		Basis:             []string{"user brief", "marketing reasoning", "Meta verified targeting"},
		Explanation:       "Only Meta-verified interests are included. Related-but-unconfirmed signals stay out of the spec.",
	}
	if len(interest.IncludedTargeting) == 0 {
		interest.Confidence = 0.4
		interest.ConfidenceLevel = domain.ConfidenceExperimental
		interest.Explanation = "No Meta-verified interests were found. Do not force targeting. Use Broad and qualify via creative."
		if len(usable) > 0 {
			interest.Explanation += " Related Meta suggestions exist but are not treated as confirmed targeting."
		}
	}

	lal := lookalikeStrategy(in, geo)
	ret := retargetingStrategy(in, geo)

	out := []domain.AudienceStrategy{broad, interest, lal}
	if ret != nil {
		out = append(out, *ret)
	}

	for i := range out {
		spec := targetingSpec(out[i])
		if e.resolver != nil {
			out[i].Reach = e.resolver.Estimate(ctx, spec)
		} else {
			out[i].Reach = domain.ReachEstimate{Available: false, Message: "Estimate unavailable"}
		}
		seg := domain.AudienceSegment{IntentLevel: domain.IntentHigh, SemanticSignals: signalNames(ents)}
		out[i].Quality = targeting.Score(seg, ents, in.Assets, out[i].Reach)
	}
	return out
}

func lookalikeStrategy(in domain.BriefInput, geo domain.GeoRecommendation) domain.AudienceStrategy {
	hasSource := in.Assets.CustomerList || in.Assets.Purchasers || in.Assets.CRM || in.Assets.Pixel
	s := domain.AudienceStrategy{
		Name:            "Lookalike",
		Type:            domain.StrategyLookalike,
		Geo:             geo,
		Ages:            []int{geo.AgeMin, geo.AgeMax},
		Confidence:      0.55,
		ConfidenceLevel: domain.ConfidenceExperimental,
		Basis:           []string{"user brief", "first-party assets"},
		Explanation:     "Recommended experiment: 1%, then 1–3%, then 3–5%. This is not a claim that a specific lookalike is best.",
	}
	if !hasSource {
		s.Explanation = "No first-party source audience was provided. Lookalike is a recommendation to collect a source (customers, purchasers, or pixel events), not a live spec."
		s.Confidence = 0.35
	}
	return s
}

func retargetingStrategy(in domain.BriefInput, geo domain.GeoRecommendation) *domain.AudienceStrategy {
	a := in.Assets
	if !a.Pixel && !a.WebsiteVisitors && !a.Instagram && !a.FacebookPage && !a.VideoViews && !a.LeadForms && !a.Purchasers {
		return &domain.AudienceStrategy{
			Name:            "Retargeting / LAL planner",
			Type:            domain.StrategyRetargeting,
			Geo:             geo,
			Confidence:      0.4,
			ConfidenceLevel: domain.ConfidenceExperimental,
			Basis:           []string{"user brief"},
			Explanation:     "Warm/hot audiences are unavailable until Pixel, page engagement, video views, or customer lists exist. Cold traffic should exclude purchasers once those lists are connected.",
		}
	}
	var included []domain.TargetingEntity
	add := func(label string) {
		included = append(included, domain.TargetingEntity{
			Query: label, Status: domain.StatusSuggested, Source: domain.SourceLLMHypothesis,
			Explanation: "First-party audience recommendation. Not a Meta interest. Create this Custom Audience in Ads Manager.",
		})
	}
	if a.Instagram {
		add("Instagram engagers — 30 days")
		add("Instagram engagers — 90 days")
	}
	if a.WebsiteVisitors || a.Pixel {
		add("Website visitors — 30 days")
	}
	if a.VideoViews {
		add("Video viewers — 50%")
	}
	if a.LeadForms {
		add("Lead form opens")
	}
	if a.Purchasers {
		add("Existing purchasers")
	}
	return &domain.AudienceStrategy{
		Name:              "Retargeting",
		Type:              domain.StrategyRetargeting,
		Geo:               geo,
		IncludedTargeting: included,
		ExcludedTargeting: exclusions(a, "warm"),
		Confidence:        0.78,
		ConfidenceLevel:   domain.ConfidenceHigh,
		Basis:             []string{"user brief", "first-party assets"},
		Explanation:       "Warm and hot layers from available first-party sources. Exclude purchasers from prospecting; exclude converters from mid-funnel retargeting.",
	}
}

func exclusions(a domain.FirstPartyAssets, layer string) []domain.TargetingEntity {
	var out []domain.TargetingEntity
	add := func(label, why string) {
		out = append(out, domain.TargetingEntity{
			Query: label, Status: domain.StatusSuggested, Source: domain.SourceLLMHypothesis, Explanation: why,
		})
	}
	if layer == "cold" && (a.Purchasers || a.CustomerList || a.CRM) {
		add("Exclude existing customers", "Cold campaigns should not pay to re-acquire current customers.")
	}
	if layer == "warm" && a.Purchasers {
		add("Exclude purchasers", "Retargeting should stop at conversion unless the offer is upsell.")
	}
	return out
}

func uniqueByMetaID(in []domain.TargetingEntity) []domain.TargetingEntity {
	seen := map[string]struct{}{}
	var out []domain.TargetingEntity
	for _, e := range in {
		if e.MetaID == "" {
			continue
		}
		if _, ok := seen[e.MetaID]; ok {
			continue
		}
		seen[e.MetaID] = struct{}{}
		out = append(out, e)
		if len(out) == 5 {
			break
		}
	}
	return out
}

func signalNames(ents []domain.TargetingEntity) []string {
	var out []string
	for _, e := range ents {
		out = append(out, e.Query)
	}
	return out
}

func targetingSpec(s domain.AudienceStrategy) meta.TargetingSpec {
	geo := map[string]any{}
	if s.Geo.CountryKey != "" {
		geo["countries"] = []string{s.Geo.CountryKey}
	} else if s.Geo.Country != "" {
		geo["countries"] = []string{guessCountryCode(s.Geo.Country)}
	}
	if s.Geo.CityKey != "" && s.Geo.LocationType == "city" {
		city := map[string]any{"key": s.Geo.CityKey}
		if s.Geo.RadiusKM > 0 {
			city["radius"] = s.Geo.RadiusKM
			city["distance_unit"] = "kilometer"
		}
		geo["cities"] = []any{city}
	}
	spec := meta.TargetingSpec{
		"geo_locations": geo,
		"age_min":       s.Geo.AgeMin,
		"age_max":       s.Geo.AgeMax,
	}
	if s.Geo.AgeMin == 0 {
		spec["age_min"] = 18
	}
	if s.Geo.AgeMax == 0 {
		spec["age_max"] = 65
	}
	var interests []map[string]string
	for _, e := range s.IncludedTargeting {
		if e.MetaID != "" && e.Status == domain.StatusVerified {
			interests = append(interests, map[string]string{"id": e.MetaID, "name": e.MetaName})
		}
	}
	if len(interests) > 0 {
		spec["flexible_spec"] = []map[string]any{{"interests": interests}}
	}
	return spec
}

func guessCountryCode(name string) string {
	switch name {
	case "Kazakhstan", "Казахстан":
		return "KZ"
	case "United States", "USA":
		return "US"
	case "Russia", "Россия":
		return "RU"
	case "United Kingdom", "UK":
		return "GB"
	default:
		if len(name) == 2 {
			return name
		}
		return name
	}
}

func Experiments(segs []domain.AudienceSegment) []domain.Experiment {
	if len(segs) == 0 {
		return nil
	}
	var variants []string
	for _, s := range segs {
		if len(s.CreativeAngles) > 0 {
			variants = append(variants, s.CreativeAngles[0])
		} else {
			variants = append(variants, s.Name)
		}
		if len(variants) == 3 {
			break
		}
	}
	return []domain.Experiment{{
		Hypothesis: fmt.Sprintf("Existing in-category users respond better to %s messaging than generic beginner messaging.", segs[0].Name),
		Audience:   segs[0].Name,
		Variable:   "creative angle",
		Variants:   variants,
	}}
}
