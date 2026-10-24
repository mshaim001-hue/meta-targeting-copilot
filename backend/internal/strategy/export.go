package strategy

import (
	"bytes"
	"fmt"
	"strings"

	"meta-targeting-copilot/backend/internal/domain"
)

func Markdown(res domain.PipelineResult) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# Targeting strategy\n\n")
	fmt.Fprintf(&b, "**Product:** %s  \n**Location:** %s, %s  \n**Audience:** %s\n\n",
		res.Brief.Product, res.Brief.City, res.Brief.Country, res.Brief.AudiencePrompt)
	fmt.Fprintf(&b, "LLM: `%s` · Meta: `%s`\n\n", res.LLMMode, res.MetaMode)

	fmt.Fprintf(&b, "## Segments\n\n")
	for _, s := range res.Segments {
		fmt.Fprintf(&b, "### %s (%s intent)\n\n%s\n\n", s.Name, s.IntentLevel, s.Description)
		fmt.Fprintf(&b, "- Pains: %s\n- Signals: %s\n\n", strings.Join(s.Pains, "; "), strings.Join(s.SemanticSignals, ", "))
	}

	fmt.Fprintf(&b, "## Meta targeting\n\n")
	for _, e := range res.Entities {
		label := e.Query
		if e.MetaName != "" {
			label = fmt.Sprintf("%s → %s (`%s`)", e.Query, e.MetaName, e.MetaID)
		}
		fmt.Fprintf(&b, "- **%s** · %s · %s — %s\n", e.Source, e.Status, label, e.Explanation)
	}

	fmt.Fprintf(&b, "\n## Geo\n\n%s / %s, radius %d km. Verified: %v\n\n%s\n\n",
		res.Geo.City, res.Geo.Country, res.Geo.RadiusKM, res.Geo.Verified, res.Geo.Reason)

	fmt.Fprintf(&b, "## Strategies\n\n")
	for _, st := range res.Strategies {
		fmt.Fprintf(&b, "### %s (`%s`)\n\n%s\n\nConfidence: %s (%.2f)\n\n", st.Name, st.Type, st.Explanation, st.ConfidenceLevel, st.Confidence)
		if st.Reach.Available {
			fmt.Fprintf(&b, "Reach: %v – %v\n\n", deref(st.Reach.LowerBound), deref(st.Reach.UpperBound))
		} else {
			fmt.Fprintf(&b, "Reach: Estimate unavailable\n\n")
		}
	}

	fmt.Fprintf(&b, "## Offers\n\n")
	for _, o := range res.Offers {
		fmt.Fprintf(&b, "### %s\n\n- Problem: %s\n- Offer: %s\n- Hook: %s\n- CTA: %s\n\n", o.SegmentName, o.Problem, o.Offer, o.Hook, o.CTA)
	}

	fmt.Fprintf(&b, "## Creatives\n\n")
	for _, c := range res.Creatives {
		fmt.Fprintf(&b, "### %s — %s\n\n- Hook: %s\n- Visual: %s\n- Format: %s\n\n", c.SegmentName, c.Angle, c.Hook, c.VisualIdea, c.Format)
	}
	return b.String()
}

func deref(n *int64) string {
	if n == nil {
		return "—"
	}
	return fmt.Sprintf("%d", *n)
}
