package memory

import (
	"awesome-chances/backend/internal/model"
	"sync"
)

// FeedbackStore is a bounded, process-local demo log, not a persistent user database.
type FeedbackStore struct {
	mu     sync.Mutex
	events map[string]model.Feedback
	order  []string
}

func NewFeedbackStore() *FeedbackStore { return &FeedbackStore{events: map[string]model.Feedback{}} }
func (s *FeedbackStore) Add(event model.Feedback) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, found := s.events[event.ID]; found {
		if existing != event {
			return false, model.ErrFeedbackConflict
		}
		return false, nil
	}
	if len(s.order) >= 1000 {
		delete(s.events, s.order[0])
		s.order = s.order[1:]
	}
	s.events[event.ID] = event
	s.order = append(s.order, event.ID)
	return true, nil
}
