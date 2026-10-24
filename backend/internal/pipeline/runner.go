package pipeline

import (
	"context"
	"fmt"

	"meta-targeting-copilot/backend/internal/ai"
	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/strategy"
	"meta-targeting-copilot/backend/internal/targeting"
)

type StageFunc func(domain.JobStage)

type Runner struct {
	ai       *ai.Engine
	resolver *targeting.Resolver
	strategy *strategy.Engine
}

func New(aiEngine *ai.Engine, resolver *targeting.Resolver) *Runner {
	return &Runner{
		ai:       aiEngine,
		resolver: resolver,
		strategy: strategy.NewEngine(resolver),
	}
}

func (r *Runner) Run(ctx context.Context, in domain.BriefInput, onStage StageFunc) (domain.PipelineResult, error) {
	report := func(s domain.JobStage) {
		if onStage != nil {
			onStage(s)
		}
	}

	report(domain.StageAnalyzingBrief)
	analysis, err := r.ai.AnalyzeBrief(ctx, in)
	if err != nil {
		return domain.PipelineResult{}, err
	}

	report(domain.StageCreatingSegments)
	segs, err := r.ai.SegmentAudience(ctx, in, analysis)
	if err != nil {
		return domain.PipelineResult{}, err
	}
	segs = targeting.DedupSegments(segs)

	report(domain.StageGeneratingSignals)
	segs, err = r.ai.GenerateSignals(ctx, segs)
	if err != nil {
		return domain.PipelineResult{}, err
	}

	report(domain.StageVerifyingMetaTargeting)
	geo, err := r.ai.RecommendGeo(ctx, in, analysis)
	if err != nil {
		return domain.PipelineResult{}, err
	}
	geo = r.resolver.ResolveGeo(ctx, geo)
	ents := targeting.StripHallucinations(r.resolver.ResolveSignals(ctx, segs))

	report(domain.StageBuildingStrategies)
	strats := r.strategy.Build(ctx, in, segs, ents, geo)

	report(domain.StageBuildingCreatives)
	offers, err := r.ai.BuildOffers(ctx, in, segs)
	if err != nil {
		return domain.PipelineResult{}, err
	}
	creatives, err := r.ai.BuildCreatives(ctx, in, segs, offers)
	if err != nil {
		return domain.PipelineResult{}, err
	}
	alignOffers(segs, &offers)
	alignCreatives(segs, &creatives)

	res := domain.PipelineResult{
		Brief:       in,
		Analysis:    analysis,
		Segments:    segs,
		Signals:     flattenSignals(segs),
		Entities:    ents,
		Geo:         geo,
		Strategies:  strats,
		Offers:      offers,
		Creatives:   creatives,
		Experiments: strategy.Experiments(segs),
		MetaMode:    r.resolver.Mode(),
		LLMMode:     r.ai.Mode(),
	}
	criticLLM, err := r.ai.Critique(ctx, res)
	if err != nil {
		criticLLM = nil
	}
	res.Critic = append(strategy.LocalCritique(res), criticLLM...)

	if err := targeting.ValidateSpecEntities(ents); err != nil {
		return res, fmt.Errorf("grounding: %w", err)
	}
	report(domain.StageComplete)
	return res, nil
}

func flattenSignals(segs []domain.AudienceSegment) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range segs {
		for _, sig := range s.SemanticSignals {
			if _, ok := seen[sig]; ok {
				continue
			}
			seen[sig] = struct{}{}
			out = append(out, sig)
		}
	}
	return out
}

func alignOffers(segs []domain.AudienceSegment, offers *[]domain.Offer) {
	if len(*offers) >= len(segs) {
		return
	}
	by := map[string]domain.Offer{}
	for _, o := range *offers {
		by[o.SegmentName] = o
	}
	var out []domain.Offer
	for _, s := range segs {
		if o, ok := by[s.Name]; ok {
			out = append(out, o)
			continue
		}
		out = append(out, domain.Offer{SegmentName: s.Name, Offer: first(s.OfferAngles), Hook: s.Description, CTA: "Learn more"})
	}
	*offers = out
}

func alignCreatives(segs []domain.AudienceSegment, creatives *[]domain.Creative) {
	if len(*creatives) >= len(segs) {
		return
	}
	by := map[string]domain.Creative{}
	for _, c := range *creatives {
		by[c.SegmentName] = c
	}
	var out []domain.Creative
	for _, s := range segs {
		if c, ok := by[s.Name]; ok {
			out = append(out, c)
			continue
		}
		out = append(out, domain.Creative{SegmentName: s.Name, Angle: first(s.CreativeAngles), Hook: s.Description, Format: "reel"})
	}
	*creatives = out
}

func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
