package ai

import (
	"context"
	"encoding/json"
	"strings"
	"unicode"

	"meta-targeting-copilot/backend/internal/domain"
)

// MockProvider returns deterministic structured output so the pipeline
// can run without an LLM key. It is marketing-heuristic, not Meta-aware.
type MockProvider struct{}

func NewMock() *MockProvider { return &MockProvider{} }

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) GenerateStructured(_ context.Context, prompt string, dest any) error {
	in := extractBrief(prompt)
	var payload any
	switch {
	case strings.Contains(prompt, "Task: Brief Analyzer"):
		payload = mockAnalysis(in)
	case strings.Contains(prompt, "Task: Audience Segmentation Agent"):
		payload = map[string]any{"segments": mockSegments(in)}
	case strings.Contains(prompt, "Task: Signal Generator"):
		payload = mockSignals(in)
	case strings.Contains(prompt, "Task: Geo Strategy Agent"):
		payload = mockGeo(in)
	case strings.Contains(prompt, "Task: Offer Strategist"):
		payload = map[string]any{"offers": mockOffers(in)}
	case strings.Contains(prompt, "Task: Creative Strategist"):
		payload = map[string]any{"creatives": mockCreatives(in)}
	case strings.Contains(prompt, "Task: Strategy Critic"):
		payload = map[string]any{"findings": []domain.CriticFinding{}}
	default:
		payload = map[string]any{}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

type briefBits struct {
	Product string
	Prompt  string
	City    string
	Country string
	Local   bool
	Fitness bool
	RU      bool
}

func extractBrief(prompt string) briefBits {
	b := briefBits{Product: "Business", City: "City", Country: "Country"}
	chunk := prompt
	if i := strings.LastIndex(prompt, "Input:"); i >= 0 {
		chunk = prompt[i:]
	}
	if v := jsonString(chunk, "product"); v != "" {
		b.Product = v
	}
	if v := jsonString(chunk, "audiencePrompt"); v != "" {
		b.Prompt = v
	}
	if v := jsonString(chunk, "city"); v != "" {
		b.City = v
	}
	if v := jsonString(chunk, "country"); v != "" {
		b.Country = v
	}
	blob := strings.ToLower(b.Product + " " + b.Prompt + " " + chunk)
	b.Local = strings.Contains(blob, "клуб") || strings.Contains(blob, "club") || strings.Contains(blob, "салон") || strings.Contains(blob, "local")
	b.Fitness = strings.Contains(blob, "фитнес") || strings.Contains(blob, "fitness") || strings.Contains(blob, "gym") || strings.Contains(blob, "зал")
	b.RU = containsCyrillic(b.Prompt + b.Product)
	return b
}

func jsonString(s, key string) string {
	needle := `"` + key + `": "`
	rest := s
	var last string
	for {
		v := between(rest, needle, `"`)
		if v == "" {
			break
		}
		if v != "string" && v != "..." {
			last = v
		}
		i := strings.Index(rest, needle)
		if i < 0 {
			break
		}
		rest = rest[i+len(needle):]
	}
	return last
}

func mockAnalysis(in briefBits) domain.BriefAnalysis {
	goal := "Leads"
	if in.Fitness {
		goal = "Gym memberships"
	}
	q := []string{}
	if in.RU {
		q = []string{
			"Есть ли база текущих клиентов для Lookalike?",
			"Установлен ли Meta Pixel на сайте?",
			"Бизнес локальный (один город) или онлайн?",
		}
	} else {
		q = []string{
			"Do you have an existing customer database?",
			"Do you have Meta Pixel installed?",
			"Is your business local or online?",
		}
	}
	return domain.BriefAnalysis{
		BusinessModel:    in.Product,
		ConversionGoal:   goal,
		DesiredAudience:  nonempty(in.Prompt, "people interested in "+in.Product),
		Intent:           "high-intent switchers and active category users",
		Constraints:      []string{"Must not claim Meta targeting without API verification"},
		MissingInfo:      []string{"first-party data availability"},
		LocalVsOnline:    map[bool]string{true: "local", false: "hybrid"}[in.Local],
		ClarifyingNeeded: true,
		Questions:        q,
	}
}

func mockSegments(in briefBits) []domain.AudienceSegment {
	if in.Fitness {
		if in.RU {
			return []domain.AudienceSegment{
				{
					Name: "Регулярные посетители зала", Description: "Люди с уже сформированной тренировочной привычкой, которые могут сменить клуб.",
					IntentLevel:     domain.IntentHigh,
					Pains:           []string{"переполненность", "слабые тренажёры", "неудобный график", "дорогая цена текущего клуба"},
					Motivations:     []string{"лучший зал без смены привычки", "прогресс", "комфорт"},
					SemanticSignals: []string{"bodybuilding", "strength training", "gym", "sports nutrition", "fitness equipment"},
					OfferAngles:     []string{"7 дней полного доступа без долгосрочного абонемента"},
					CreativeAngles:  []string{"performance", "equipment quality"},
				},
				{
					Name: "Функциональный тренинг", Description: "Люди, которым важны CrossFit, HIIT и функциональные форматы.",
					IntentLevel:     domain.IntentHigh,
					Pains:           []string{"нет подходящих классов", "слабое комьюнити"},
					Motivations:     []string{"интенсив", "сообщество", "результат"},
					SemanticSignals: []string{"CrossFit", "functional training", "HIIT", "endurance training"},
					OfferAngles:     []string{"пробная функциональная неделя"},
					CreativeAngles:  []string{"community", "intensity"},
				},
				{
					Name: "Бег и кардио", Description: "Бегуны и кардио-ориентированные пользователи, которым нужен зал как база.",
					IntentLevel:     domain.IntentMedium,
					Pains:           []string{"нет беговой инфраструктуры", "слабые кардио-зоны"},
					Motivations:     []string{"подготовка к стартам", "восстановление"},
					SemanticSignals: []string{"running", "marathon", "Garmin", "Strava", "running shoes", "cardio training"},
					OfferAngles:     []string{"зал как база для беговой подготовки"},
					CreativeAngles:  []string{"endurance", "gear"},
				},
			}
		}
		return []domain.AudienceSegment{
			{
				Name: "Regular Gym Users", Description: "People who already train consistently and may switch gyms.",
				IntentLevel:     domain.IntentHigh,
				Pains:           []string{"overcrowding", "weak equipment", "inconvenient hours", "price of current club"},
				Motivations:     []string{"better gym without restarting the habit", "progress"},
				SemanticSignals: []string{"bodybuilding", "strength training", "gym", "sports nutrition", "fitness equipment"},
				OfferAngles:     []string{"7-day full access without a long contract"},
				CreativeAngles:  []string{"performance", "equipment"},
			},
			{
				Name: "Functional Training", Description: "People drawn to CrossFit, HIIT and functional formats.",
				IntentLevel:     domain.IntentHigh,
				Pains:           []string{"no suitable classes", "weak community"},
				Motivations:     []string{"intensity", "community", "results"},
				SemanticSignals: []string{"CrossFit", "functional training", "HIIT", "endurance training"},
				OfferAngles:     []string{"trial functional week"},
				CreativeAngles:  []string{"community", "intensity"},
			},
			{
				Name: "Running / Cardio", Description: "Runners and cardio-oriented users who need a gym as a base.",
				IntentLevel:     domain.IntentMedium,
				Pains:           []string{"no running infrastructure", "weak cardio zone"},
				Motivations:     []string{"race prep", "recovery"},
				SemanticSignals: []string{"running", "marathon", "Garmin", "Strava", "running shoes", "cardio training"},
				OfferAngles:     []string{"gym as a running base"},
				CreativeAngles:  []string{"endurance", "gear"},
			},
		}
	}
	product := in.Product
	return []domain.AudienceSegment{
		{
			Name: "High-intent switchers", Description: "People already in-category who may change provider.",
			IntentLevel:     domain.IntentHigh,
			Pains:           []string{"current option is inconvenient", "price", "inconsistent quality"},
			Motivations:     []string{"better outcome with less friction"},
			SemanticSignals: []string{product, "reviews", "comparison", "premium"},
			OfferAngles:     []string{"low-risk trial"},
			CreativeAngles:  []string{"proof", "contrast"},
		},
		{
			Name: "Active category users", Description: "People with demonstrated interest in the category.",
			IntentLevel:     domain.IntentMedium,
			Pains:           []string{"too many options", "unclear quality"},
			Motivations:     []string{"confidence", "social proof"},
			SemanticSignals: []string{product, "how to", "best", "near me"},
			OfferAngles:     []string{"guided first visit"},
			CreativeAngles:  []string{"education", "local proof"},
		},
		{
			Name: "Adjacent interest", Description: "People in neighboring categories who can be qualified via creative.",
			IntentLevel:     domain.IntentLow,
			Pains:           []string{"not actively looking"},
			Motivations:     []string{"curiosity", "aspiration"},
			SemanticSignals: []string{product, "lifestyle", "community"},
			OfferAngles:     []string{"soft entry offer"},
			CreativeAngles:  []string{"lifestyle"},
		},
	}
}

func mockSignals(in briefBits) map[string]any {
	segs := mockSegments(in)
	out := make([]map[string]any, 0, len(segs))
	for _, s := range segs {
		out = append(out, map[string]any{"name": s.Name, "signals": s.SemanticSignals})
	}
	return map[string]any{"segments": out}
}

func mockGeo(in briefBits) domain.GeoRecommendation {
	radius := 25
	if in.Local {
		radius = 5
	}
	reason := "Geo should follow the stated market; radius is a hypothesis until Meta verifies the location."
	if in.Local && in.Fitness {
		reason = "For a local gym, conversion usually drops quickly with distance. Start tight, then test a wider radius."
	}
	return domain.GeoRecommendation{
		Country: in.Country, City: in.City, RadiusKM: radius, AltRadiusKM: radius + 3,
		AgeMin: 23, AgeMax: 42, Reason: reason,
	}
}

func mockOffers(in briefBits) []domain.Offer {
	var out []domain.Offer
	for _, s := range mockSegments(in) {
		offer := "Low-risk trial for this segment."
		hook := "Already in this category? Compare without a long commitment."
		cta := "See details"
		if in.Fitness && in.RU {
			offer = "7 дней полного доступа без долгосрочного абонемента."
			hook = "Уже тренируешься, но текущий зал перестал устраивать?"
			cta = "Попробовать клуб"
		} else if in.Fitness {
			offer = "7-day premium access without a long contract."
			hook = "Already training, but the current gym stopped working for you?"
			cta = "Try the club"
		}
		problem := ""
		if len(s.Pains) > 0 {
			problem = s.Pains[0]
		}
		out = append(out, domain.Offer{
			SegmentName: s.Name, AwarenessLevel: "solution-aware", Problem: problem,
			Motivation: first(s.Motivations), Objection: "switching cost",
			Offer: offer, Hook: hook, CTA: cta,
		})
	}
	return out
}

func mockCreatives(in briefBits) []domain.Creative {
	var out []domain.Creative
	for _, s := range mockSegments(in) {
		angle := first(s.CreativeAngles)
		hook := "A specific problem, not a generic category slogan."
		visual := "Show the promised outcome in a real environment."
		if in.Fitness && in.RU {
			hook = "Тренировка не должна состоять из ожидания тренажёра."
			visual = "Зал в час пик, свободные тренажёры, крупный план оборудования."
		} else if in.Fitness {
			hook = "Training should not be waiting for a machine."
			visual = "Peak-hour floor with free equipment, close-up of machines."
		}
		out = append(out, domain.Creative{
			SegmentName: s.Name, Angle: angle, Hook: hook, VisualIdea: visual,
			CopyAngle: s.Description, CTA: "Learn more", Format: "reel",
		})
	}
	return out
}

func between(s, a, b string) string {
	i := strings.Index(s, a)
	if i < 0 {
		return ""
	}
	s = s[i+len(a):]
	j := strings.Index(s, b)
	if j < 0 {
		return ""
	}
	return s[:j]
}

func nonempty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func containsCyrillic(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Cyrillic) {
			return true
		}
	}
	return false
}
