package meta

import (
	"context"
	"encoding/json"
	"time"
)

type Entity struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Path         []string `json:"path,omitempty"`
	AudienceSize *int64   `json:"audienceSize,omitempty"`
	Description  string   `json:"description,omitempty"`
	Restricted   bool     `json:"restricted,omitempty"`
}

type Location struct {
	Key            string  `json:"key"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	CountryCode    string  `json:"countryCode,omitempty"`
	RegionID       string  `json:"regionId,omitempty"`
	SupportsCity   bool    `json:"supportsCity,omitempty"`
	SupportsRegion bool    `json:"supportsRegion,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
}

type Estimate struct {
	Available  bool            `json:"available"`
	LowerBound *int64          `json:"lowerBound,omitempty"`
	UpperBound *int64          `json:"upperBound,omitempty"`
	Message    string          `json:"message,omitempty"`
	Raw        json.RawMessage `json:"raw,omitempty"`
}

type TargetingSpec map[string]any

type Provider interface {
	Name() string
	Connected() bool
	SearchInterests(ctx context.Context, query string) ([]Entity, error)
	SearchBehaviors(ctx context.Context, query string) ([]Entity, error)
	SearchDemographics(ctx context.Context, query string) ([]Entity, error)
	SearchLocations(ctx context.Context, query string) ([]Location, error)
	GetSuggestions(ctx context.Context, names []string) ([]Entity, error)
	EstimateAudience(ctx context.Context, spec TargetingSpec) (Estimate, error)
}

type CacheStore interface {
	Get(ctx context.Context, query, country, kind, version string) (json.RawMessage, time.Time, bool, error)
	Put(ctx context.Context, query, country, kind, version string, result json.RawMessage) error
}

type CachedProvider struct {
	inner   Provider
	cache   CacheStore
	ttl     time.Duration
	version string
	country string
}

func WithCache(inner Provider, cache CacheStore, ttl time.Duration, version, country string) Provider {
	if cache == nil {
		return inner
	}
	return &CachedProvider{inner: inner, cache: cache, ttl: ttl, version: version, country: country}
}

func (c *CachedProvider) Name() string    { return c.inner.Name() }
func (c *CachedProvider) Connected() bool { return c.inner.Connected() }

func (c *CachedProvider) SearchInterests(ctx context.Context, query string) ([]Entity, error) {
	return cachedSlice[Entity](ctx, c, "interest", query, func() ([]Entity, error) {
		return c.inner.SearchInterests(ctx, query)
	})
}

func (c *CachedProvider) SearchBehaviors(ctx context.Context, query string) ([]Entity, error) {
	return cachedSlice[Entity](ctx, c, "behavior", query, func() ([]Entity, error) {
		return c.inner.SearchBehaviors(ctx, query)
	})
}

func (c *CachedProvider) SearchDemographics(ctx context.Context, query string) ([]Entity, error) {
	return cachedSlice[Entity](ctx, c, "demographic", query, func() ([]Entity, error) {
		return c.inner.SearchDemographics(ctx, query)
	})
}

func (c *CachedProvider) SearchLocations(ctx context.Context, query string) ([]Location, error) {
	return cachedSlice[Location](ctx, c, "geo", query, func() ([]Location, error) {
		return c.inner.SearchLocations(ctx, query)
	})
}

func (c *CachedProvider) GetSuggestions(ctx context.Context, names []string) ([]Entity, error) {
	key, _ := json.Marshal(names)
	return cachedSlice[Entity](ctx, c, "suggest", string(key), func() ([]Entity, error) {
		return c.inner.GetSuggestions(ctx, names)
	})
}

func (c *CachedProvider) EstimateAudience(ctx context.Context, spec TargetingSpec) (Estimate, error) {
	key, _ := json.Marshal(spec)
	raw, ts, ok, err := c.cache.Get(ctx, string(key), c.country, "estimate", c.version)
	if err == nil && ok && time.Since(ts) < c.ttl {
		var est Estimate
		if json.Unmarshal(raw, &est) == nil {
			return est, nil
		}
	}
	est, err := c.inner.EstimateAudience(ctx, spec)
	if err != nil {
		return est, err
	}
	b, _ := json.Marshal(est)
	_ = c.cache.Put(ctx, string(key), c.country, "estimate", c.version, b)
	return est, nil
}

func cachedSlice[T any](ctx context.Context, c *CachedProvider, kind, query string, fetch func() ([]T, error)) ([]T, error) {
	raw, ts, ok, err := c.cache.Get(ctx, query, c.country, kind, c.version)
	if err == nil && ok && time.Since(ts) < c.ttl {
		var out []T
		if json.Unmarshal(raw, &out) == nil {
			return out, nil
		}
	}
	out, err := fetch()
	if err != nil {
		return out, err
	}
	b, _ := json.Marshal(out)
	_ = c.cache.Put(ctx, query, c.country, kind, c.version, b)
	return out, nil
}
