package tests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	etherws "github.com/Lunidra/Ether-Backend/internal/websocket"
)

func TestServerBroadcast(t *testing.T) {
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

	err := server.Broadcast(
		"Ether",
		"Welcome to Ether!",
	)

	if err != nil {
		t.Fatalf(
			"failed to broadcast: %v",
			err,
		)
	}

	packetA := readUntilType(
		t,
		clientA,
		"broadcast",
	)

	packetB := readUntilType(
		t,
		clientB,
		"broadcast",
	)

	assertBroadcast(
		t,
		packetA,
		"Ether",
		"Welcome to Ether!",
	)

	assertBroadcast(
		t,
		packetB,
		"Ether",
		"Welcome to Ether!",
	)
}

func TestBroadcastDoesNotReachUnauthenticatedClient(
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

	client, _, err := websocket.DefaultDialer.Dial(
		wsURL,
		nil,
	)

	if err != nil {
		t.Fatalf(
			"failed to connect: %v",
			err,
		)
	}

	defer client.Close()

	err = server.Broadcast(
		"Ether",
		"You should not receive this yet.",
	)

	if err != nil {
		t.Fatalf(
			"broadcast failed: %v",
			err,
		)
	}

	_ = client.SetReadDeadline(
		time.Now().Add(250 * time.Millisecond),
	)

	_, _, err = client.ReadMessage()

	if err == nil {
		t.Fatal(
			"unauthenticated client received broadcast",
		)
	}
}

func readUntilType(
	t *testing.T,
	connection *websocket.Conn,
	packetType string,
) map[string]any {

	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {

		_ = connection.SetReadDeadline(
			deadline,
		)

		_, data, err := connection.ReadMessage()

		if err != nil {
			t.Fatalf(
				"failed waiting for %s packet: %v",
				packetType,
				err,
			)
		}

		var packet map[string]any

		if err := json.Unmarshal(
			data,
			&packet,
		); err != nil {
			continue
		}

		if packet["type"] == packetType {
			return packet
		}
	}

	t.Fatalf(
		"timed out waiting for %s packet",
		packetType,
	)

	return nil
}

func assertBroadcast(
	t *testing.T,
	packet map[string]any,
	expectedSender string,
	expectedMessage string,
) {

	t.Helper()

	if packet["type"] != "broadcast" {
		t.Fatalf(
			"expected broadcast packet, got %v",
			packet["type"],
		)
	}

	if packet["sender"] != expectedSender {
		t.Fatalf(
			"expected sender %q, got %v",
			expectedSender,
			packet["sender"],
		)
	}

	if packet["message"] != expectedMessage {
		t.Fatalf(
			"expected message %q, got %v",
			expectedMessage,
			packet["message"],
		)
	}
}
