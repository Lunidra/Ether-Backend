package minecraft

import (
	"fmt"
	"sync"
)

type Service struct {
	mu    sync.RWMutex
	links map[string]Link
}

func NewService() *Service {
	return &Service{
		links: make(map[string]Link),
	}
}

func (s *Service) Link(uuid string, username string) error {
	uuid = normalizeUUID(uuid)

	if uuid == "" {
		return fmt.Errorf("missing Minecraft UUID")
	}

	if username == "" {
		return fmt.Errorf("missing Minecraft username")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.links[uuid] = Link{
		UUID:     uuid,
		Username: username,
	}

	return nil
}

func (s *Service) Unlink(uuid string) bool {
	uuid = normalizeUUID(uuid)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[uuid]; !exists {
		return false
	}

	delete(s.links, uuid)

	return true
}

func (s *Service) Get(uuid string) (Link, bool) {
	uuid = normalizeUUID(uuid)

	s.mu.RLock()
	defer s.mu.RUnlock()

	link, exists := s.links[uuid]

	return link, exists
}

func (s *Service) IsLinked(uuid string) bool {
	_, exists := s.Get(uuid)

	return exists
}
