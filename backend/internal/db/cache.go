package db

import (
	"context"
	"encoding/json"
	"time"
)

func (s *Store) Get(ctx context.Context, query, country, kind, version string) (json.RawMessage, time.Time, bool, error) {
	return s.CacheGet(ctx, query, country, kind, version)
}

func (s *Store) Put(ctx context.Context, query, country, kind, version string, result json.RawMessage) error {
	return s.CachePut(ctx, query, country, kind, version, result)
}
