package tests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	etherws "github.com/Lunidra/Ether-Backend/internal/websocket"
)

func TestChatRelay(t *testing.T) {

	server := etherws.NewDevelopmentServer("")

	httpServer := httptest.NewServer(
		server.Handler(),
	)

	defer httpServer.Close()

	wsURL := "ws" +
		strings.TrimPrefix(
			httpServer.URL,
			"http",
		) +
		"/ws"

	clientA := connectDevClient(
		t,
		wsURL,
		"dev-a-uuid",
		"Player100",
	)

	defer clientA.Close()

	clientB := connectDevClient(
		t,
		wsURL,
		"dev-b-uuid",
		"Player200",
	)

	defer clientB.Close()

	/*
	 * Send presence first so the resulting chat message
	 * proves that the backend uses authoritative presence.
	 */
	if err := clientA.WriteJSON(
		map[string]any{
			"type":        "presence",
			"server":      "Singleplayer",
			"serverIP":    "localhost",
			"gameVersion": "1.21.11",
		},
	); err != nil {
		t.Fatalf(
			"failed to send A presence: %v",
			err,
		)
	}

	if err := clientB.WriteJSON(
		map[string]any{
			"type":        "presence",
			"server":      "Test Server",
			"serverIP":    "127.0.0.1",
			"gameVersion": "1.21.11",
		},
	); err != nil {
		t.Fatalf(
			"failed to send B presence: %v",
			err,
		)
	}

	/*
	 * Presence updates create messages in both sockets.
	 * Drain them before testing chat itself.
	 */
	drainMessages(
		clientA,
		2,
	)

	drainMessages(
		clientB,
		2,
	)

	/*
	 * This mirrors the current Ether-Client ChatPacket:
	 *
	 * {
	 *     "type": "chat",
	 *     "data": {
	 *         "message": "hello from A"
	 *     }
	 * }
	 */
	chatPacket := map[string]any{
		"type": "chat",
		"data": map[string]any{
			"message": "hello from A",
		},
	}

	if err := clientA.WriteJSON(
		chatPacket,
	); err != nil {
		t.Fatalf(
			"failed to send chat packet: %v",
			err,
		)
	}

	messageA := readMessage(
		t,
		clientA,
	)

	messageB := readMessage(
		t,
		clientB,
	)

	assertChatMessage(
		t,
		messageA,
		"dev-a-uuid",
		"Player100",
		"hello from A",
		"Singleplayer",
		"localhost",
	)

	assertChatMessage(
		t,
		messageB,
		"dev-a-uuid",
		"Player100",
		"hello from A",
		"Singleplayer",
		"localhost",
	)
}

func TestChatRejectsUnauthenticatedClient(
	t *testing.T,
) {

	server := etherws.NewDevelopmentServer("")

	httpServer := httptest.NewServer(
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
			"failed to connect: %v",
			err,
		)
	}

	defer connection.Close()

	err = connection.WriteJSON(
		map[string]any{
			"type": "chat",
			"data": map[string]any{
				"message": "unauthenticated",
			},
		},
	)

	if err != nil {
		t.Fatalf(
			"failed to send chat packet: %v",
			err,
		)
	}

	/*
	 * The backend currently logs rejected packets rather than
	 * sending protocol error packets, so we expect no response.
	 */
	_ = connection.SetReadDeadline(
		time.Now().Add(250 * time.Millisecond),
	)

	_, _, err =
		connection.ReadMessage()

	if err == nil {
		t.Fatal(
			"expected unauthenticated chat to produce no response",
		)
	}
}

func TestBroadcast(t *testing.T) {

	server := etherws.NewDevelopmentServer("")

	httpServer := httptest.NewServer(
		server.Handler(),
	)

	defer httpServer.Close()

	wsURL := "ws" +
		strings.TrimPrefix(
			httpServer.URL,
			"http",
		) +
		"/ws"

	clientA := connectDevClient(
		t,
		wsURL,
		"broadcast-a",
		"Player300",
	)

	defer clientA.Close()

	clientB := connectDevClient(
		t,
		wsURL,
		"broadcast-b",
		"Player400",
	)

	defer clientB.Close()

	if err := server.Broadcast(
		"Ether",
		"Test broadcast",
	); err != nil {
		t.Fatalf(
			"failed to send broadcast: %v",
			err,
		)
	}

	messageA := readMessage(
		t,
		clientA,
	)

	messageB := readMessage(
		t,
		clientB,
	)

	assertBroadcast(
		t,
		messageA,
		"Ether",
		"Test broadcast",
	)

	assertBroadcast(
		t,
		messageB,
		"Ether",
		"Test broadcast",
	)
}

func connectDevClient(
	t *testing.T,
	wsURL string,
	uuid string,
	username string,
) *websocket.Conn {

	t.Helper()

	connection, _, err :=
		websocket.DefaultDialer.Dial(
			wsURL,
			nil,
		)

	if err != nil {
		t.Fatalf(
			"failed to connect %s: %v",
			username,
			err,
		)
	}

	err = connection.WriteJSON(
		map[string]any{
			"type":     "auth_hello",
			"uuid":     uuid,
			"username": username,
		},
	)

	if err != nil {
		connection.Close()

		t.Fatalf(
			"failed to send auth_hello for %s: %v",
			username,
			err,
		)
	}

	challenge := readMessage(
		t,
		connection,
	)

	if challenge["type"] != "auth_challenge" {
		connection.Close()

		t.Fatalf(
			"expected auth_challenge for %s, got %v",
			username,
			challenge["type"],
		)
	}

	serverID, ok :=
		challenge["serverId"].(string)

	if !ok || serverID == "" {
		connection.Close()

		t.Fatalf(
			"auth_challenge for %s missing serverId",
			username,
		)
	}

	err = connection.WriteJSON(
		map[string]any{
			"type":     "auth_verify",
			"serverId": serverID,
		},
	)

	if err != nil {
		connection.Close()

		t.Fatalf(
			"failed to send auth_verify for %s: %v",
			username,
			err,
		)
	}

	success := readMessage(
		t,
		connection,
	)

	if success["type"] != "auth_success" {
		connection.Close()

		t.Fatalf(
			"expected auth_success for %s, got %v",
			username,
			success["type"],
		)
	}

	return connection
}

func drainMessages(
	connection *websocket.Conn,
	count int,
) {

	for i := 0; i < count; i++ {
		_ = connection.SetReadDeadline(
			time.Now().Add(2 * time.Second),
		)

		_, _, err :=
			connection.ReadMessage()

		if err != nil {
			return
		}
	}
}

func assertChatMessage(
	t *testing.T,
	message map[string]any,
	expectedUUID string,
	expectedUsername string,
	expectedText string,
	expectedServer string,
	expectedIP string,
) {

	t.Helper()

	if message["uuid"] != expectedUUID {
		t.Fatalf(
			"expected uuid %q, got %v",
			expectedUUID,
			message["uuid"],
		)
	}

	if message["username"] != expectedUsername {
		t.Fatalf(
			"expected username %q, got %v",
			expectedUsername,
			message["username"],
		)
	}

	if message["message"] != expectedText {
		t.Fatalf(
			"expected message %q, got %v",
			expectedText,
			message["message"],
		)
	}

	if message["server"] != expectedServer {
		t.Fatalf(
			"expected server %q, got %v",
			expectedServer,
			message["server"],
		)
	}

	if message["serverIP"] != expectedIP {
		t.Fatalf(
			"expected serverIP %q, got %v",
			expectedIP,
			message["serverIP"],
		)
	}

	if message["isStaff"] != false {
		t.Fatalf(
			"expected isStaff=false, got %v",
			message["isStaff"],
		)
	}

	if message["isDeveloper"] != false {
		t.Fatalf(
			"expected isDeveloper=false, got %v",
			message["isDeveloper"],
		)
	}

	if message["isBetaTester"] != false {
		t.Fatalf(
			"expected isBetaTester=false, got %v",
			message["isBetaTester"],
		)
	}

	/*
	 * A normal chat packet must NOT contain a "type".
	 *
	 * This is important because the current client falls through
	 * to Gson -> ChatMessage when processing ordinary chat.
	 */
	if _, exists := message["type"]; exists {
		t.Fatalf(
			"normal ChatMessage unexpectedly contained type: %v",
			message["type"],
		)
	}
}

/*
func assertBroadcast(
	t *testing.T,
	message map[string]any,
	expectedSender string,
	expectedMessage string,
) {

	t.Helper()

	if message["type"] != "broadcast" {
		t.Fatalf(
			"expected broadcast packet, got %v",
			message["type"],
		)
	}

	if message["sender"] != expectedSender {
		t.Fatalf(
			"expected sender %q, got %v",
			expectedSender,
			message["sender"],
		)
	}

	if message["message"] != expectedMessage {
		t.Fatalf(
			"expected message %q, got %v",
			expectedMessage,
			message["message"],
		)
	}
}
*/
// Keep auth imported here because this test file intentionally exercises
// the development authentication path through the real server.
