package meta

import (
	"context"
	"strings"
)

type catalogItem struct {
	Entity
	aliases []string
}

type locItem struct {
	Location
	aliases []string
}

type Mock struct {
	interests []catalogItem
	locations []locItem
}

func NewMock() *Mock {
	return &Mock{
		interests: []catalogItem{
			{Entity: Entity{ID: "6003107902433", Name: "Physical fitness", Type: "interest", AudienceSize: ptr(980000000)}, aliases: []string{"fitness", "physical fitness", "фитнес"}},
			{Entity: Entity{ID: "6003348444581", Name: "Bodybuilding", Type: "interest", AudienceSize: ptr(120000000)}, aliases: []string{"bodybuilding", "бодибилдинг", "strength training"}},
			{Entity: Entity{ID: "6003271232160", Name: "Gym", Type: "interest", AudienceSize: ptr(450000000)}, aliases: []string{"gym", "зал", "тренажёр"}},
			{Entity: Entity{ID: "6003022516946", Name: "Weight training", Type: "interest", AudienceSize: ptr(80000000)}, aliases: []string{"weight training", "strength training", "силовые"}},
			{Entity: Entity{ID: "6003270817352", Name: "Sports nutrition", Type: "interest", AudienceSize: ptr(35000000)}, aliases: []string{"sports nutrition", "спортивное питание"}},
			{Entity: Entity{ID: "6003270222142", Name: "Fitness (magazine)", Type: "interest", AudienceSize: ptr(15000000)}, aliases: []string{"fitness magazine"}},
			{Entity: Entity{ID: "6003020834693", Name: "CrossFit", Type: "interest", AudienceSize: ptr(28000000)}, aliases: []string{"crossfit", "functional training"}},
			{Entity: Entity{ID: "6003348604411", Name: "High-intensity interval training", Type: "interest", AudienceSize: ptr(18000000)}, aliases: []string{"hiit", "high-intensity"}},
			{Entity: Entity{ID: "6003348438488", Name: "Running", Type: "interest", AudienceSize: ptr(320000000)}, aliases: []string{"running", "бег", "cardio training"}},
			{Entity: Entity{ID: "6003022260983", Name: "Marathon", Type: "interest", AudienceSize: ptr(45000000)}, aliases: []string{"marathon", "марафон"}},
			{Entity: Entity{ID: "6003139266461", Name: "Garmin", Type: "interest", AudienceSize: ptr(22000000)}, aliases: []string{"garmin"}},
			{Entity: Entity{ID: "6003270982211", Name: "Strava", Type: "interest", AudienceSize: ptr(14000000)}, aliases: []string{"strava"}},
			{Entity: Entity{ID: "6003108322433", Name: "Running shoes", Type: "interest", AudienceSize: ptr(40000000)}, aliases: []string{"running shoes"}},
			{Entity: Entity{ID: "6003022516144", Name: "Exercise equipment", Type: "interest", AudienceSize: ptr(60000000)}, aliases: []string{"fitness equipment", "equipment"}},
			{Entity: Entity{ID: "6003246126468", Name: "Wellness", Type: "interest", AudienceSize: ptr(500000000)}, aliases: []string{"wellness"}},
			{Entity: Entity{ID: "6003139266999", Name: "Personal trainer", Type: "interest", AudienceSize: ptr(25000000)}, aliases: []string{"personal trainer", "тренер"}},
			{Entity: Entity{ID: "6003445123001", Name: "Real estate", Type: "interest", AudienceSize: ptr(400000000)}, aliases: []string{"real estate", "недвижимость", "квартир"}},
			{Entity: Entity{ID: "6003445123002", Name: "Home improvement", Type: "interest", AudienceSize: ptr(250000000)}, aliases: []string{"home improvement", "ремонт"}},
			{Entity: Entity{ID: "6003445123003", Name: "Education", Type: "interest", AudienceSize: ptr(600000000)}, aliases: []string{"education", "образован", "english", "school"}},
			{Entity: Entity{ID: "6003445123004", Name: "Dentistry", Type: "interest", AudienceSize: ptr(80000000)}, aliases: []string{"dentist", "стоматолог", "dentistry"}},
			{Entity: Entity{ID: "6003445123005", Name: "Restaurants", Type: "interest", AudienceSize: ptr(700000000)}, aliases: []string{"restaurant", "ресторан"}},
			{Entity: Entity{ID: "6003445123006", Name: "Beauty salon", Type: "interest", AudienceSize: ptr(120000000)}, aliases: []string{"beauty", "салон", "красот"}},
			{Entity: Entity{ID: "6003445123007", Name: "Software as a service", Type: "interest", AudienceSize: ptr(40000000)}, aliases: []string{"saas", "b2b", "software"}},
		},
		locations: []locItem{
			{Location: Location{Key: "KZ", Name: "Kazakhstan", Type: "country", CountryCode: "KZ", SupportsCity: true, SupportsRegion: true}, aliases: []string{"kazakhstan", "казахстан", "kz"}},
			{Location: Location{Key: "2430205", Name: "Astana", Type: "city", CountryCode: "KZ", Latitude: 51.1694, Longitude: 71.4491}, aliases: []string{"astana", "астана", "nur-sultan", "нур-султан"}},
			{Location: Location{Key: "2430536", Name: "Almaty", Type: "city", CountryCode: "KZ", Latitude: 43.2220, Longitude: 76.8512}, aliases: []string{"almaty", "алматы", "алма-ата"}},
			{Location: Location{Key: "US", Name: "United States", Type: "country", CountryCode: "US", SupportsCity: true, SupportsRegion: true}, aliases: []string{"united states", "usa", "us"}},
			{Location: Location{Key: "RU", Name: "Russia", Type: "country", CountryCode: "RU", SupportsCity: true, SupportsRegion: true}, aliases: []string{"russia", "россия", "ru"}},
			{Location: Location{Key: "GB", Name: "United Kingdom", Type: "country", CountryCode: "GB", SupportsCity: true}, aliases: []string{"united kingdom", "uk", "britain"}},
		},
	}
}

func (m *Mock) Name() string    { return "mock" }
func (m *Mock) Connected() bool { return false }

func (m *Mock) SearchInterests(_ context.Context, query string) ([]Entity, error) {
	return m.match(query), nil
}

func (m *Mock) SearchBehaviors(ctx context.Context, query string) ([]Entity, error) {
	return nil, nil
}

func (m *Mock) SearchDemographics(ctx context.Context, query string) ([]Entity, error) {
	return nil, nil
}

func (m *Mock) SearchLocations(_ context.Context, query string) ([]Location, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || q == "string" || q == "city" {
		return nil, nil
	}
	var out []Location
	for _, loc := range m.locations {
		if strings.EqualFold(loc.Name, query) || strings.Contains(strings.ToLower(loc.Name), q) {
			out = append(out, loc.Location)
			continue
		}
		for _, a := range loc.aliases {
			if q == a || strings.Contains(q, a) || strings.Contains(a, q) {
				out = append(out, loc.Location)
				break
			}
		}
	}
	return out, nil
}

func (m *Mock) GetSuggestions(_ context.Context, names []string) ([]Entity, error) {
	seen := map[string]struct{}{}
	var out []Entity
	for _, n := range names {
		for _, e := range m.match(n) {
			if _, ok := seen[e.ID]; ok {
				continue
			}
			seen[e.ID] = struct{}{}
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *Mock) EstimateAudience(_ context.Context, spec TargetingSpec) (Estimate, error) {
	base := int64(180000)
	if geo, ok := spec["geo_locations"].(map[string]any); ok {
		if cities, ok := geo["cities"].([]any); ok && len(cities) > 0 {
			base = 42000
		}
	}
	if ints, ok := spec["flexible_spec"]; ok && ints != nil {
		base = base / 3
		if base < 8000 {
			base = 8000
		}
	}
	hi := base * 2
	return Estimate{Available: true, LowerBound: &base, UpperBound: &hi, Message: "Mock estimate — not from Meta"}, nil
}

func (m *Mock) match(query string) []Entity {
	q := strings.ToLower(strings.TrimSpace(query))
	long := len(strings.Fields(q)) > 4
	var exact, fuzzy []Entity
	for _, item := range m.interests {
		name := strings.ToLower(item.Name)
		if name == q {
			exact = append(exact, item.Entity)
			continue
		}
		if long {
			continue
		}
		hit := strings.Contains(name, q) || strings.Contains(q, name)
		if !hit {
			for _, a := range item.aliases {
				if q == a || strings.Contains(q, a) || strings.Contains(a, q) {
					hit = true
					break
				}
			}
		}
		if hit {
			fuzzy = append(fuzzy, item.Entity)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	if len(fuzzy) > 6 {
		fuzzy = fuzzy[:6]
	}
	return fuzzy
}

func ptr(n int64) *int64 { return &n }
