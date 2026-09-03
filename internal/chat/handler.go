package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Lunidra/Ether-Backend/internal/broadcast"
	"github.com/Lunidra/Ether-Backend/internal/protocol"
	"github.com/Lunidra/Ether-Backend/internal/session"
)

const MaxMessageLength = 2048

type Handler struct {
	broadcast *broadcast.Service
}

func NewHandler(
	broadcastService *broadcast.Service,
) *Handler {
	return &Handler{
		broadcast: broadcastService,
	}
}

func (h *Handler) Handle(
	client *session.Client,
	message protocol.Message,
) error {

	if !client.IsAuthenticated() {
		return fmt.Errorf(
			"authentication required",
		)
	}

	/*
	 * Current Ether-Client sends:
	 *
	 * {
	 *     "type": "chat",
	 *     "data": {
	 *         "message": "hello"
	 *     }
	 * }
	 *
	 * Extract the nested data object.
	 */
	var payload IncomingPacket

	if err := message.Field(
		"data",
		&payload,
	); err != nil {

		/*
		 * Be slightly backwards-compatible and also accept:
		 *
		 * {
		 *     "type": "chat",
		 *     "message": "hello"
		 * }
		 *
		 * This costs us almost nothing and makes protocol
		 * migration easier.
		 */
		if fallbackErr := message.Field(
			"message",
			&payload.Message,
		); fallbackErr != nil {
			return fmt.Errorf(
				"chat packet missing data.message: %w",
				err,
			)
		}
	}

	payload.Message = strings.TrimSpace(
		payload.Message,
	)

	if payload.Message == "" {
		return fmt.Errorf(
			"chat message cannot be empty",
		)
	}

	if !utf8.ValidString(payload.Message) {
		return fmt.Errorf(
			"chat message contains invalid UTF-8",
		)
	}

	if len([]rune(payload.Message)) > MaxMessageLength {
		return fmt.Errorf(
			"chat message exceeds %d characters",
			MaxMessageLength,
		)
	}

	/*
	 * Everything below this point is authoritative backend data.
	 */
	uuid, username := client.Identity()
	presence := client.GetPresence()

	outgoing := Message{
		UUID:     uuid,
		Username: username,
		Message:  payload.Message,

		Server:      presence.Server,
		ServerIP:    presence.ServerIP,
		GameVersion: presence.GameVersion,

		/*
		 * Roles are deliberately backend-controlled.
		 *
		 * These will later come from the account/permission
		 * service instead of being client supplied.
		 */
		IsStaff:      false,
		IsDeveloper:  false,
		IsBetaTester: false,
	}

	data, err := json.Marshal(
		outgoing,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to encode chat message: %w",
			err,
		)
	}

	/*
	 * Important:
	 *
	 * Do NOT use protocol.LegacyMessage("chat", ...).
	 *
	 * The current Ether-Client expects ordinary incoming chat
	 * to deserialize directly into ChatMessage.
	 */
	return h.broadcast.BroadcastRaw(data)
}
