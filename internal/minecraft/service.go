package minecraft

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
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

func (s *Service) Link(rawUUID string, username string) error {
	parsedUUID, err := uuid.Parse(strings.TrimSpace(rawUUID))
	if err != nil {
		return fmt.Errorf("invalid Minecraft UUID")
	}

	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("missing Minecraft username")
	}

	if len(username) > 16 {
		return fmt.Errorf("invalid Minecraft username")
	}

	normalizedUUID := strings.ToLower(parsedUUID.String())

	s.mu.Lock()
	defer s.mu.Unlock()

	s.links[normalizedUUID] = Link{
		UUID:     normalizedUUID,
		Username: username,
	}

	return nil
}

func (s *Service) Unlink(rawUUID string) bool {
	parsedUUID, err := uuid.Parse(strings.TrimSpace(rawUUID))
	if err != nil {
		return false
	}

	normalizedUUID := strings.ToLower(parsedUUID.String())

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[normalizedUUID]; !exists {
		return false
	}

	delete(s.links, normalizedUUID)

	return true
}

func (s *Service) Get(rawUUID string) (Link, bool) {
	parsedUUID, err := uuid.Parse(strings.TrimSpace(rawUUID))
	if err != nil {
		return Link{}, false
	}

	normalizedUUID := strings.ToLower(parsedUUID.String())

	s.mu.RLock()
	defer s.mu.RUnlock()

	link, exists := s.links[normalizedUUID]

	return link, exists
}

func (s *Service) IsLinked(rawUUID string) bool {
	_, exists := s.Get(rawUUID)

	return exists
}
