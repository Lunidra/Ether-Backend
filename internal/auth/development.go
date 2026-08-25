package auth

type DevelopmentVerifier struct{}

func NewDevelopmentVerifier() *DevelopmentVerifier {
	return &DevelopmentVerifier{}
}

func (v *DevelopmentVerifier) HasJoined(
	username string,
	serverID string,
) (MojangProfile, bool, error) {

	if username == "" {
		return MojangProfile{}, false, nil
	}

	return MojangProfile{
		Name: username,
	}, true, nil
}
