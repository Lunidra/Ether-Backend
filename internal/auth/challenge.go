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
	ExpiresAt time.Time
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
		ExpiresAt: time.Now().Add(challengeLifetime),
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

	if time.Now().After(challenge.ExpiresAt) {

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
	clientID string,
) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, exists :=
		s.challenges[serverID]

	if !exists {
		return false
	}

	if challenge.ClientID != clientID {
		return false
	}

	delete(
		s.challenges,
		serverID,
	)

	return !time.Now().After(challenge.ExpiresAt)
}

func (s *ChallengeStore) RemoveForClient(
	clientID string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for serverID, challenge := range s.challenges {
		if challenge.ClientID == clientID {
			delete(s.challenges, serverID)
		}
	}
}
