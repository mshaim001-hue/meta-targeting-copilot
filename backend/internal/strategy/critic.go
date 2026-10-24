package strategy

import (
	"strings"

	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/targeting"
)

func LocalCritique(res domain.PipelineResult) []domain.CriticFinding {
	var findings []domain.CriticFinding
	for _, e := range res.Entities {
		if (e.Status == domain.StatusVerified || e.Source == domain.SourceMetaVerified) && e.MetaID == "" {
			findings = append(findings, domain.CriticFinding{Severity: "error", Message: "Hallucinated targeting blocked: " + e.Query})
		}
	}
	names := map[string]int{}
	for _, s := range res.Segments {
		key := strings.ToLower(s.Name)
		names[key]++
		if names[key] > 1 {
			findings = append(findings, domain.CriticFinding{Severity: "warning", Message: "Duplicate segment name: " + s.Name})
		}
	}
	if len(targeting.OnlyVerified(res.Entities)) == 0 {
		findings = append(findings, domain.CriticFinding{
			Severity: "info",
			Message:  "No Meta-verified interests. Broad + creative qualification is the honest recommendation.",
		})
	}
	for _, st := range res.Strategies {
		if st.Type == domain.StrategyInterest && len(targeting.OnlyVerified(st.IncludedTargeting)) > 8 {
			findings = append(findings, domain.CriticFinding{Severity: "warning", Message: "Interest strategy stacks too many conditions and may over-constrain delivery."})
		}
	}
	if len(res.Offers) < len(res.Segments) {
		findings = append(findings, domain.CriticFinding{Severity: "warning", Message: "Some segments are missing a dedicated offer."})
	}
	return findings
}
