package idempotency

import "sync"

type Store struct {
	processed sync.Map
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) TryMarkProcessed(eventID string) bool {
	_, alreadyExists := s.processed.LoadOrStore(eventID, struct{}{})
	return !alreadyExists
}
