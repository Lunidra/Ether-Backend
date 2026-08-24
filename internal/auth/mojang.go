package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type HTTPMojangVerifier struct {
	client *http.Client
}

func NewMojangVerifier() *HTTPMojangVerifier {
	return &HTTPMojangVerifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type MojangProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (v *HTTPMojangVerifier) HasJoined(
	username string,
	serverID string,
) (MojangProfile, bool, error) {

	query := url.Values{}

	query.Set(
		"username",
		username,
	)

	query.Set(
		"serverId",
		serverID,
	)

	endpoint :=
		"https://sessionserver.mojang.com/session/minecraft/hasJoined?" +
			query.Encode()

	request, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)

	if err != nil {
		return MojangProfile{}, false, err
	}

	response, err := v.client.Do(request)

	if err != nil {
		return MojangProfile{}, false, fmt.Errorf(
			"Mojang verification request failed: %w",
			err,
		)
	}

	defer response.Body.Close()

	switch response.StatusCode {

	case http.StatusOK:

		var profile MojangProfile

		if err := json.NewDecoder(
			response.Body,
		).Decode(&profile); err != nil {
			return MojangProfile{}, false, fmt.Errorf(
				"invalid Mojang response: %w",
				err,
			)
		}

		if profile.ID == "" || profile.Name == "" {
			return MojangProfile{}, false, fmt.Errorf(
				"invalid Mojang profile response",
			)
		}

		return profile, true, nil

	case http.StatusNoContent:
		return MojangProfile{}, false, nil

	default:
		return MojangProfile{}, false, fmt.Errorf(
			"Mojang returned HTTP %d",
			response.StatusCode,
		)
	}
}
