package game

import "sync"

type Service struct {
	mu     sync.RWMutex
	states map[int]*State
}

func New() *Service {
	return &Service{states: make(map[int]*State)}
}

func (s *Service) Init(gameID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[gameID] = NewState()
}

func (s *Service) State(gameID int) *State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[gameID]
}
