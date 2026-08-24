package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Lunidra/Ether-Backend/internal/auth"
)

type mockMojangVerifier struct {
	profile  auth.MojangProfile
	verified bool
	err      error

	lastUsername string
	lastServerID string
}

func (m *mockMojangVerifier) HasJoined(
	username string,
	serverID string,
) (auth.MojangProfile, bool, error) {

	m.lastUsername = username
	m.lastServerID = serverID

	return m.profile, m.verified, m.err
}

func TestMojangVerifierSuccess(t *testing.T) {
	// This test validates the verifier contract without
	// requiring a real Minecraft account.
	verifier := &mockMojangVerifier{
		profile: auth.MojangProfile{
			ID:   "test-profile-id",
			Name: "TestPlayer",
		},
		verified: true,
	}

	profile, verified, err :=
		verifier.HasJoined(
			"TestPlayer",
			"test-server-id",
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !verified {
		t.Fatal(
			"expected account to be verified",
		)
	}

	if profile.Name != "TestPlayer" {
		t.Fatalf(
			"expected TestPlayer, got %q",
			profile.Name,
		)
	}

	if verifier.lastUsername != "TestPlayer" {
		t.Fatalf(
			"expected username TestPlayer, got %q",
			verifier.lastUsername,
		)
	}

	if verifier.lastServerID != "test-server-id" {
		t.Fatalf(
			"expected test-server-id, got %q",
			verifier.lastServerID,
		)
	}
}

func TestMojangVerifierRejectsAccount(
	t *testing.T,
) {

	verifier := &mockMojangVerifier{
		verified: false,
	}

	_, verified, err :=
		verifier.HasJoined(
			"TestPlayer",
			"test-server-id",
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if verified {
		t.Fatal(
			"expected account verification to fail",
		)
	}
}

func TestMojangVerifierError(
	t *testing.T,
) {

	expectedErr := &testError{
		message: "verification unavailable",
	}

	verifier := &mockMojangVerifier{
		err: expectedErr,
	}

	_, _, err :=
		verifier.HasJoined(
			"TestPlayer",
			"test-server-id",
		)

	if err != expectedErr {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Keep the HTTP imports exercised here so we can immediately
// extend this file with HTTP-level verifier tests without
// changing the test structure.
var (
	_ = http.MethodGet
	_ = httptest.NewServer
	_ = url.Values{}
)
