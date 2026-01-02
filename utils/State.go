package utils

import (
	"maps"
	"sync"
)

// State is thread safe.  It uses mutexes for Getting and Setting of values and hence
// is costly to use.  It is encouraged to keep cached variables holding values retrieved
// from the State.  Obviously objects contained in the state has to facilitate
// its own mechanisms of concurrency.
type State struct {
	data  map[string]any
	mutex sync.RWMutex
}

func (s *State) Init() {
	s.data = make(map[string]any, 20)
}

func (s *State) GetOrSet(key string, value any) any {
	s.mutex.RLock()
	res := s.data[key]
	s.mutex.RUnlock()

	if res == nil {
		s.mutex.Lock()
		s.data[key] = value
		s.mutex.Unlock()
		return value
	}
	return res
}

func (s *State) Set(key string, value any) any {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value

	return value
}

func (s *State) Get(key string) any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data[key]
}

func (s *State) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)
}

func (s *State) GetKeys() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	res := make([]string, 0, len(s.data))

	for name := range maps.Keys(s.data) {
		res = append(res, name)
	}

	return res
}

var GlobalState State

func init() {
	GlobalState.Init()
}
