package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const testServerURL = "ws://127.0.0.1:2832/ws"

type testMessage struct {
	Version int             `json:"version"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type identifyPayload struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type challengePayload struct {
	ServerID string `json:"serverId"`
}

type verifyPayload struct {
	ServerID string `json:"serverId"`
	Proof    string `json:"proof"`
}

func readMessage(t *testing.T, conn *websocket.Conn) testMessage {
	t.Helper()

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read WebSocket message: %v", err)
	}

	var message testMessage

	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatalf("failed to decode server message: %v\nRaw: %s", err, data)
	}

	return message
}

func sendMessage(t *testing.T, conn *websocket.Conn, message testMessage) {
	t.Helper()

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
}

func TestAuthenticationFlow(t *testing.T) {
	conn, _, err := websocket.DefaultDialer.Dial(testServerURL, nil)
	if err != nil {
		t.Fatalf(
			"failed to connect to Ether Backend at %s: %v\n"+
				"Make sure the server is running first.",
			testServerURL,
			err,
		)
	}
	defer conn.Close()

	t.Log("connected to Ether Backend")

	// ---------------------------------------------------------
	// 1. auth_identify
	// ---------------------------------------------------------

	sendMessage(t, conn, testMessage{
		Version: 1,
		Type:    "auth_identify",
		Payload: mustJSON(t, identifyPayload{
			UUID:     "test-uuid",
			Username: "AuthTestPlayer",
		}),
	})

	t.Log("sent auth_identify")

	// ---------------------------------------------------------
	// 2. auth_challenge
	// ---------------------------------------------------------

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	challengeMessage := readMessage(t, conn)

	if challengeMessage.Type != "auth_challenge" {
		t.Fatalf(
			"expected auth_challenge, got %q",
			challengeMessage.Type,
		)
	}

	var challenge challengePayload

	if err := json.Unmarshal(challengeMessage.Payload, &challenge); err != nil {
		t.Fatalf("failed to decode challenge: %v", err)
	}

	if challenge.ServerID == "" {
		t.Fatal("server returned an empty serverId")
	}

	t.Logf("received challenge: %s", challenge.ServerID)

	// ---------------------------------------------------------
	// 3. auth_verify
	// ---------------------------------------------------------

	sendMessage(t, conn, testMessage{
		Version: 1,
		Type:    "auth_verify",
		Payload: mustJSON(t, verifyPayload{
			ServerID: challenge.ServerID,
			Proof:    challenge.ServerID,
		}),
	})

	t.Log("sent auth_verify")

	// ---------------------------------------------------------
	// 4. auth_success
	// ---------------------------------------------------------

	successMessage := readMessage(t, conn)

	if successMessage.Type != "auth_success" {
		t.Fatalf(
			"expected auth_success, got %q",
			successMessage.Type,
		)
	}

	var identity identifyPayload

	if err := json.Unmarshal(successMessage.Payload, &identity); err != nil {
		t.Fatalf("failed to decode auth_success: %v", err)
	}

	if identity.UUID != "test-uuid" {
		t.Fatalf(
			"expected UUID %q, got %q",
			"test-uuid",
			identity.UUID,
		)
	}

	if identity.Username != "AuthTestPlayer" {
		t.Fatalf(
			"expected username %q, got %q",
			"AuthTestPlayer",
			identity.Username,
		)
	}

	t.Log("authentication succeeded")

	// ---------------------------------------------------------
	// 5. Replay attack test
	// ---------------------------------------------------------

	sendMessage(t, conn, testMessage{
		Version: 1,
		Type:    "auth_verify",
		Payload: mustJSON(t, verifyPayload{
			ServerID: challenge.ServerID,
			Proof:    challenge.ServerID,
		}),
	})

	t.Log("sent replayed auth_verify")

	// The server should reject the consumed challenge.
	//
	// Depending on the current router/server error handling,
	// the server may return an error packet or close the
	// connection. Either behavior proves the challenge wasn't
	// accepted a second time.

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	message, err := readMessageWithoutFatal(conn)

	if err != nil {
		t.Logf("replay rejected by connection handling: %v", err)
		return
	}

	if message.Type == "auth_success" {
		t.Fatal("SECURITY FAILURE: replayed challenge was accepted")
	}

	t.Logf("replay rejected with message type: %s", message.Type)
}

func readMessageWithoutFatal(conn *websocket.Conn) (testMessage, error) {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return testMessage{}, err
	}

	var message testMessage

	if err := json.Unmarshal(data, &message); err != nil {
		return testMessage{}, fmt.Errorf(
			"invalid server message: %w",
			err,
		)
	}

	return message, nil
}

func mustJSON(t *testing.T, value interface{}) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to encode JSON payload: %v", err)
	}

	return data
}