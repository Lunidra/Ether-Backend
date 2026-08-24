package auth

type Service struct {
	challenges *ChallengeStore
	mojang     *MojangVerifier
}

func NewService() *Service {
	return &Service{
		challenges: NewChallengeStore(),
		mojang:     NewMojangVerifier(),
	}
}

func (s *Service) Challenges() *ChallengeStore {
	return s.challenges
}

func (s *Service) Mojang() *MojangVerifier {
	return s.mojang
}