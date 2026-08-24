package auth

type MojangVerifier interface {
	HasJoined(
		username string,
		serverID string,
	) (MojangProfile, bool, error)
}

type Service struct {
	challenges *ChallengeStore
	mojang     MojangVerifier
}

func NewService() *Service {
	return NewServiceWithVerifier(
		NewMojangVerifier(),
	)
}

func NewServiceWithVerifier(
	verifier MojangVerifier,
) *Service {
	return &Service{
		challenges: NewChallengeStore(),
		mojang:     verifier,
	}
}

func (s *Service) Challenges() *ChallengeStore {
	return s.challenges
}

func (s *Service) Mojang() MojangVerifier {
	return s.mojang
}
