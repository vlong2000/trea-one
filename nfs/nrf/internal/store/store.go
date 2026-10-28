package store

import (
	"strings"
	"sync"
)

type Profile map[string]any

type Store struct {
	mu       sync.RWMutex
	profiles map[string]Profile
}

func New() *Store {
	return &Store{profiles: make(map[string]Profile)}
}

func (s *Store) Put(id string, p Profile) (existed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, existed = s.profiles[id]
	copied := clone(p)
	copied["nfInstanceId"] = id
	s.profiles[id] = copied
	return existed
}

func (s *Store) Get(id string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, false
	}
	return clone(p), true
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.profiles[id]; !ok {
		return false
	}
	delete(s.profiles, id)
	return true
}

func (s *Store) List(nfType string) []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Profile, 0, len(s.profiles))
	want := strings.ToUpper(nfType)
	for _, p := range s.profiles {
		if want != "" && strings.ToUpper(asString(p["nfType"])) != want {
			continue
		}
		out = append(out, clone(p))
	}
	return out
}

func (s *Store) Discover(targetType, requesterType string) []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	target := strings.ToUpper(targetType)
	requester := strings.ToUpper(requesterType)
	out := make([]Profile, 0)
	for _, p := range s.profiles {
		if target != "" && strings.ToUpper(asString(p["nfType"])) != target {
			continue
		}
		if requester != "" && !allowed(p, requester) {
			continue
		}
		out = append(out, clone(p))
	}
	return out
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.profiles)
}

func allowed(p Profile, requester string) bool {
	raw, ok := p["allowedNfTypes"]
	if !ok || raw == nil {
		return true
	}
	list, ok := raw.([]any)
	if !ok {
		return true
	}
	if len(list) == 0 {
		return true
	}
	for _, item := range list {
		if strings.ToUpper(asString(item)) == requester {
			return true
		}
	}
	return false
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func clone(p Profile) Profile {
	out := make(Profile, len(p))
	for k, v := range p {
		out[k] = v
	}
	return out
}
