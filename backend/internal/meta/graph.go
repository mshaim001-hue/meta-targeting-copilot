package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GraphClient struct {
	token      string
	adAccount  string
	version    string
	httpClient *http.Client
}

func NewGraph(token, adAccount, version string) *GraphClient {
	if version == "" {
		version = "v26.0"
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return &GraphClient{
		token:      token,
		adAccount:  strings.TrimPrefix(adAccount, "act_"),
		version:    version,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *GraphClient) Name() string    { return "meta-graph:" + c.version }
func (c *GraphClient) Connected() bool { return c.token != "" }

func (c *GraphClient) SearchInterests(ctx context.Context, query string) ([]Entity, error) {
	return c.searchEntities(ctx, url.Values{
		"type": {"adinterest"},
		"q":    {query},
	})
}

func (c *GraphClient) SearchBehaviors(ctx context.Context, query string) ([]Entity, error) {
	return c.searchEntities(ctx, url.Values{
		"type":  {"adTargetingCategory"},
		"class": {"behaviors"},
		"q":     {query},
	})
}

func (c *GraphClient) SearchDemographics(ctx context.Context, query string) ([]Entity, error) {
	return c.searchEntities(ctx, url.Values{
		"type":  {"adTargetingCategory"},
		"class": {"demographics"},
		"q":     {query},
	})
}

func (c *GraphClient) SearchLocations(ctx context.Context, query string) ([]Location, error) {
	q := url.Values{
		"type":           {"adgeolocation"},
		"q":              {query},
		"location_types": {`["country","region","city"]`},
		"access_token":   {c.token},
	}
	var wrap struct {
		Data []struct {
			Key            string  `json:"key"`
			Name           string  `json:"name"`
			Type           string  `json:"type"`
			CountryCode    string  `json:"country_code"`
			RegionID       any     `json:"region_id"`
			SupportsCity   bool    `json:"supports_city"`
			SupportsRegion bool    `json:"supports_region"`
			Latitude       float64 `json:"latitude"`
			Longitude      float64 `json:"longitude"`
		} `json:"data"`
		Error *graphError `json:"error"`
	}
	if err := c.get(ctx, "/search", q, &wrap); err != nil {
		return nil, err
	}
	if wrap.Error != nil {
		return nil, wrap.Error
	}
	out := make([]Location, 0, len(wrap.Data))
	for _, d := range wrap.Data {
		out = append(out, Location{
			Key: d.Key, Name: d.Name, Type: d.Type, CountryCode: d.CountryCode,
			RegionID: fmt.Sprint(d.RegionID), SupportsCity: d.SupportsCity,
			SupportsRegion: d.SupportsRegion, Latitude: d.Latitude, Longitude: d.Longitude,
		})
	}
	return out, nil
}

func (c *GraphClient) GetSuggestions(ctx context.Context, names []string) ([]Entity, error) {
	list, _ := json.Marshal(names)
	return c.searchEntities(ctx, url.Values{
		"type":          {"adinterestsuggestion"},
		"interest_list": {string(list)},
	})
}

func (c *GraphClient) EstimateAudience(ctx context.Context, spec TargetingSpec) (Estimate, error) {
	if c.adAccount == "" {
		return Estimate{Available: false, Message: "META_AD_ACCOUNT_ID is not set"}, nil
	}
	rawSpec, err := json.Marshal(spec)
	if err != nil {
		return Estimate{}, err
	}
	q := url.Values{
		"targeting_spec": {string(rawSpec)},
		"access_token":   {c.token},
	}
	var wrap struct {
		Data struct {
			UsersLower *int64 `json:"users_lower_bound"`
			UsersUpper *int64 `json:"users_upper_bound"`
			Estimate   *int64 `json:"estimate_mau"`
		} `json:"data"`
		UsersLower *int64      `json:"users_lower_bound"`
		UsersUpper *int64      `json:"users_upper_bound"`
		Error      *graphError `json:"error"`
	}
	if err := c.get(ctx, "/act_"+c.adAccount+"/reachestimate", q, &wrap); err != nil {
		return Estimate{Available: false, Message: err.Error()}, nil
	}
	if wrap.Error != nil {
		return Estimate{Available: false, Message: wrap.Error.Error()}, nil
	}
	lo, hi := wrap.UsersLower, wrap.UsersUpper
	if lo == nil {
		lo = wrap.Data.UsersLower
	}
	if hi == nil {
		hi = wrap.Data.UsersUpper
	}
	if lo == nil && hi == nil && wrap.Data.Estimate != nil {
		lo, hi = wrap.Data.Estimate, wrap.Data.Estimate
	}
	if lo != nil && *lo < 0 {
		return Estimate{Available: false, Message: "Estimate unavailable"}, nil
	}
	if lo == nil && hi == nil {
		return Estimate{Available: false, Message: "Estimate unavailable"}, nil
	}
	return Estimate{Available: true, LowerBound: lo, UpperBound: hi}, nil
}

func (c *GraphClient) searchEntities(ctx context.Context, q url.Values) ([]Entity, error) {
	q.Set("access_token", c.token)
	q.Set("limit", "8")
	var wrap struct {
		Data []struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			Type         string   `json:"type"`
			Path         []string `json:"path"`
			AudienceSize *int64   `json:"audience_size"`
			Description  string   `json:"description"`
		} `json:"data"`
		Error *graphError `json:"error"`
	}
	if err := c.get(ctx, "/search", q, &wrap); err != nil {
		return nil, err
	}
	if wrap.Error != nil {
		return nil, wrap.Error
	}
	out := make([]Entity, 0, len(wrap.Data))
	typ := q.Get("type")
	for _, d := range wrap.Data {
		metaType := d.Type
		if metaType == "" {
			switch typ {
			case "adinterest", "adinterestsuggestion":
				metaType = "interest"
			case "adTargetingCategory":
				if q.Get("class") == "behaviors" {
					metaType = "behavior"
				} else {
					metaType = "demographic"
				}
			default:
				metaType = "interest"
			}
		}
		out = append(out, Entity{
			ID: d.ID, Name: d.Name, Type: metaType, Path: d.Path,
			AudienceSize: d.AudienceSize, Description: d.Description,
		})
	}
	return out, nil
}

func (c *GraphClient) get(ctx context.Context, path string, q url.Values, dest any) error {
	u := fmt.Sprintf("https://graph.facebook.com/%s%s?%s", c.version, path, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("meta json: %w (%s)", err, truncate(string(b), 240))
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("meta http %d: %s", resp.StatusCode, truncate(string(b), 240))
	}
	return nil
}

type graphError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}

func (e *graphError) Error() string {
	return fmt.Sprintf("meta api: %s (code %d)", e.Message, e.Code)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
