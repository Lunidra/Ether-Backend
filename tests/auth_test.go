package tests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Lunidra/Ether-Backend/internal/auth"
	etherws "github.com/Lunidra/Ether-Backend/internal/websocket"
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

func TestAuthenticationFlow(
	t *testing.T,
) {

	verifier := &mockMojangVerifier{
		profile: auth.MojangProfile{
			ID:   "test-profile-id",
			Name: "TestPlayer",
		},
		verified: true,
	}

	authService :=
		auth.NewServiceWithVerifier(
			verifier,
		)

	server :=
		etherws.NewServerWithService(
			"",
			authService,
		)

	httpServer :=
		httptest.NewServer(
			server.Handler(),
		)

	defer httpServer.Close()

	wsURL := "ws" +
		strings.TrimPrefix(
			httpServer.URL,
			"http",
		) +
		"/ws"

	connection, _, err :=
		websocket.DefaultDialer.Dial(
			wsURL,
			nil,
		)

	if err != nil {
		t.Fatalf(
			"failed to connect to Ether Backend: %v",
			err,
		)
	}

	defer connection.Close()

	t.Log(
		"connected to Ether Backend",
	)

	identify := map[string]any{
		"type":     "auth_hello",
		"uuid":     "test-profile-id",
		"username": "TestPlayer",
	}

	if err := connection.WriteJSON(
		identify,
	); err != nil {

		t.Fatalf(
			"failed to send auth_hello: %v",
			err,
		)
	}

	t.Log(
		"sent auth_hello",
	)

	challengeMessage :=
		readMessage(
			t,
			connection,
		)

	if challengeMessage["type"] !=
		"auth_challenge" {

		t.Fatalf(
			"expected auth_challenge, got %v",
			challengeMessage["type"],
		)
	}

	serverID, ok :=
		challengeMessage["serverId"].(string)

	if !ok || serverID == "" {
		t.Fatal(
			"auth_challenge did not contain serverId",
		)
	}

	t.Logf(
		"received challenge: %s",
		serverID,
	)

	verify := map[string]any{
		"type":     "auth_verify",
		"serverId": serverID,
	}

	if err := connection.WriteJSON(
		verify,
	); err != nil {

		t.Fatalf(
			"failed to send auth_verify: %v",
			err,
		)
	}

	t.Log(
		"sent auth_verify",
	)

	successMessage :=
		readMessage(
			t,
			connection,
		)

	if successMessage["type"] !=
		"auth_success" {

		t.Fatalf(
			"expected auth_success, got %v",
			successMessage["type"],
		)
	}

	token, ok :=
		successMessage["token"].(string)

	if !ok || token == "" {
		t.Fatal(
			"auth_success did not contain token",
		)
	}

	t.Logf(
		"received authentication token: %s",
		token,
	)

	if verifier.lastUsername !=
		"TestPlayer" {

		t.Fatalf(
			"Mojang verifier received wrong username: %q",
			verifier.lastUsername,
		)
	}

	if verifier.lastServerID != serverID {

		t.Fatalf(
			"Mojang verifier received wrong serverId: %q",
			verifier.lastServerID,
		)
	}

	// The challenge must be consumed after successful
	// authentication. Reusing it must therefore fail.
	if err := connection.WriteJSON(
		verify,
	); err != nil {

		t.Fatalf(
			"failed to send replay auth_verify: %v",
			err,
		)
	}

	// The server currently logs rejected packets rather than
	// returning protocol error packets, so we simply make sure
	// no second auth_success is generated.
	_ = connection.SetReadDeadline(
		time.Now().Add(250 * time.Millisecond),
	)

	_, _, err =
		connection.ReadMessage()

	if err == nil {
		t.Fatal(
			"expected replayed challenge to produce no response",
		)
	}

	t.Log(
		"challenge replay correctly rejected",
	)
}

func readMessage(
	t *testing.T,
	connection *websocket.Conn,
) map[string]any {

	t.Helper()

	_, data, err :=
		connection.ReadMessage()

	if err != nil {
		t.Fatalf(
			"failed to read server message: %v",
			err,
		)
	}

	var message map[string]any

	if err := json.Unmarshal(
		data,
		&message,
	); err != nil {

		t.Fatalf(
			"invalid JSON from server: %v",
			err,
		)
	}

	return message
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestDevelopmentVerifier(t *testing.T) {
	verifier := auth.NewDevelopmentVerifier()

	profile, verified, err := verifier.HasJoined(
		"Player676",
		"development-test",
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !verified {
		t.Fatal(
			"expected development account to be verified",
		)
	}

	if profile.Name != "Player676" {
		t.Fatalf(
			"expected Player676, got %q",
			profile.Name,
		)
	}
}

// Keep url imported while this test evolves toward testing
// external HTTP endpoint construction.
