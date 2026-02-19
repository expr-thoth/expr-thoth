package ds

import (
	"fmt"
	"sync"
)

type Map struct{}

type SyncMap struct {
	mu   sync.RWMutex
	data map[string]any
}

func (s Map) Set(m any, key string, value any) bool {
	switch m := m.(type) {
	case map[string]any:
		m[key] = value
	case map[string]string:
		m[key] = fmt.Sprintf("%v", value)
	}

	return true
}

func (s Map) CreateSyncMap(m map[string]any) *SyncMap {
	return &SyncMap{
		data: m,
	}
}

func (s *SyncMap) Set(key string, value any) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return true
}

func (s *SyncMap) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	return value, ok
}
