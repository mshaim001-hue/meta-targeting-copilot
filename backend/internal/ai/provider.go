package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"meta-targeting-copilot/backend/internal/domain"
)

type Provider interface {
	Name() string
	GenerateStructured(ctx context.Context, prompt string, dest any) error
}

type Engine struct {
	llm Provider
}

func NewEngine(llm Provider) *Engine {
	return &Engine{llm: llm}
}

func (e *Engine) Mode() string {
	if e == nil || e.llm == nil {
		return "none"
	}
	return e.llm.Name()
}

func (e *Engine) AnalyzeBrief(ctx context.Context, in domain.BriefInput) (domain.BriefAnalysis, error) {
	var out domain.BriefAnalysis
	if err := e.llm.GenerateStructured(ctx, briefAnalyzerPrompt(in), &out); err != nil {
		return out, fmt.Errorf("brief analyzer: %w", err)
	}
	if len(out.Questions) > 5 {
		out.Questions = out.Questions[:5]
	}
	return out, nil
}

func (e *Engine) SegmentAudience(ctx context.Context, in domain.BriefInput, analysis domain.BriefAnalysis) ([]domain.AudienceSegment, error) {
	var wrap struct {
		Segments []domain.AudienceSegment `json:"segments"`
	}
	if err := e.llm.GenerateStructured(ctx, segmentationPrompt(in, analysis), &wrap); err != nil {
		return nil, fmt.Errorf("segmentation: %w", err)
	}
	segs := wrap.Segments
	if len(segs) < 3 {
		return nil, fmt.Errorf("segmentation: expected at least 3 segments, got %d", len(segs))
	}
	if len(segs) > 7 {
		segs = segs[:7]
	}
	for i := range segs {
		if err := validateSegment(segs[i]); err != nil {
			return nil, fmt.Errorf("segment %d: %w", i, err)
		}
		switch segs[i].IntentLevel {
		case domain.IntentLow, domain.IntentMedium, domain.IntentHigh:
		default:
			segs[i].IntentLevel = domain.IntentMedium
		}
	}
	return segs, nil
}

func (e *Engine) GenerateSignals(ctx context.Context, segs []domain.AudienceSegment) ([]domain.AudienceSegment, error) {
	var wrap struct {
		Segments []struct {
			Name    string   `json:"name"`
			Signals []string `json:"signals"`
		} `json:"segments"`
	}
	if err := e.llm.GenerateStructured(ctx, signalPrompt(segs), &wrap); err != nil {
		return nil, fmt.Errorf("signal generator: %w", err)
	}
	byName := map[string][]string{}
	for _, s := range wrap.Segments {
		byName[s.Name] = uniqueNonEmpty(s.Signals)
	}
	out := make([]domain.AudienceSegment, len(segs))
	copy(out, segs)
	for i := range out {
		if extra, ok := byName[out[i].Name]; ok && len(extra) > 0 {
			out[i].SemanticSignals = uniqueNonEmpty(append(out[i].SemanticSignals, extra...))
		}
		if len(out[i].SemanticSignals) == 0 {
			return nil, fmt.Errorf("signal generator: segment %q has no signals", out[i].Name)
		}
	}
	return out, nil
}

func (e *Engine) RecommendGeo(ctx context.Context, in domain.BriefInput, analysis domain.BriefAnalysis) (domain.GeoRecommendation, error) {
	var out domain.GeoRecommendation
	if err := e.llm.GenerateStructured(ctx, geoPrompt(in, analysis), &out); err != nil {
		return out, fmt.Errorf("geo strategist: %w", err)
	}
	if out.Country == "" {
		out.Country = in.Country
	}
	if out.City == "" {
		out.City = in.City
	}
	if out.AgeMin <= 0 {
		out.AgeMin = 18
	}
	if out.AgeMax <= out.AgeMin {
		out.AgeMax = 45
	}
	if out.RadiusKM <= 0 {
		out.RadiusKM = 8
	}
	return out, nil
}

func (e *Engine) BuildOffers(ctx context.Context, in domain.BriefInput, segs []domain.AudienceSegment) ([]domain.Offer, error) {
	var wrap struct {
		Offers []domain.Offer `json:"offers"`
	}
	if err := e.llm.GenerateStructured(ctx, offerPrompt(in, segs), &wrap); err != nil {
		return nil, fmt.Errorf("offer strategist: %w", err)
	}
	if len(wrap.Offers) == 0 {
		return nil, fmt.Errorf("offer strategist: empty result")
	}
	return wrap.Offers, nil
}

func (e *Engine) BuildCreatives(ctx context.Context, in domain.BriefInput, segs []domain.AudienceSegment, offers []domain.Offer) ([]domain.Creative, error) {
	var wrap struct {
		Creatives []domain.Creative `json:"creatives"`
	}
	if err := e.llm.GenerateStructured(ctx, creativePrompt(in, segs, offers), &wrap); err != nil {
		return nil, fmt.Errorf("creative strategist: %w", err)
	}
	if len(wrap.Creatives) == 0 {
		return nil, fmt.Errorf("creative strategist: empty result")
	}
	return wrap.Creatives, nil
}

func (e *Engine) Critique(ctx context.Context, result domain.PipelineResult) ([]domain.CriticFinding, error) {
	var wrap struct {
		Findings []domain.CriticFinding `json:"findings"`
	}
	payload, _ := json.Marshal(map[string]any{
		"segments":   result.Segments,
		"entities":   result.Entities,
		"strategies": result.Strategies,
		"offers":     result.Offers,
	})
	if err := e.llm.GenerateStructured(ctx, criticPrompt(string(payload)), &wrap); err != nil {
		return nil, fmt.Errorf("strategy critic: %w", err)
	}
	return wrap.Findings, nil
}

func uniqueNonEmpty(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func validateSegment(s domain.AudienceSegment) error {
	if s.Name == "" || s.Description == "" {
		return fmt.Errorf("name and description required")
	}
	if len(s.SemanticSignals) == 0 {
		return fmt.Errorf("semantic signals required")
	}
	return nil
}
