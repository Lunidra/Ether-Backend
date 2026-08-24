package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const challengeLength = 32

const challengeLifetime = 2 * time.Minute

type Challenge struct {
	ServerID  string
	ClientID  string
	CreatedAt time.Time
}

type ChallengeStore struct {
	mu         sync.Mutex
	challenges map[string]Challenge
}

func NewChallengeStore() *ChallengeStore {
	return &ChallengeStore{
		challenges: make(
			map[string]Challenge,
		),
	}
}

func (s *ChallengeStore) Create(
	clientID string,
) (Challenge, error) {

	bytes := make(
		[]byte,
		challengeLength,
	)

	if _, err := rand.Read(bytes); err != nil {
		return Challenge{}, err
	}

	challenge := Challenge{
		ServerID:  hex.EncodeToString(bytes),
		ClientID:  clientID,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.challenges[challenge.ServerID] = challenge
	s.mu.Unlock()

	return challenge, nil
}

func (s *ChallengeStore) Get(
	serverID string,
) (Challenge, bool) {

	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, exists :=
		s.challenges[serverID]

	if !exists {
		return Challenge{}, false
	}

	if time.Since(
		challenge.CreatedAt,
	) > challengeLifetime {

		delete(
			s.challenges,
			serverID,
		)

		return Challenge{}, false
	}

	return challenge, true
}

func (s *ChallengeStore) Consume(
	serverID string,
) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, exists :=
		s.challenges[serverID]

	if !exists {
		return false
	}

	delete(
		s.challenges,
		serverID,
	)

	return time.Since(
		challenge.CreatedAt,
	) <= challengeLifetime
}