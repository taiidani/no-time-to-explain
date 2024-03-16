package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type memoryStore struct {
	data map[string][]byte
}

func (s *memoryStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (err error) {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	s.data[dbPrefix+key] = data
	return nil
}

func (s *memoryStore) Get(ctx context.Context, key string, value interface{}) error {
	data, ok := s.data[dbPrefix+key]
	if !ok {
		return fmt.Errorf("key not found")
	}

	return json.Unmarshal(data, value)
}
