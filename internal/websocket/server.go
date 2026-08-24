package websocket

import (
	"log"
	"net/http"

	gws "github.com/gorilla/websocket"

	"github.com/Lunidra/Ether-Backend/internal/auth"
	"github.com/Lunidra/Ether-Backend/internal/protocol"
	"github.com/Lunidra/Ether-Backend/internal/session"
)

const (
	readBufferSize  = 4096
	writeBufferSize = 4096
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,

	// Development only.
	// This must be restricted before production deployment.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	addr        string
	router      *protocol.Router
	clients     *session.Manager
	authService *auth.Service
}

func NewServer(addr string) *Server {

	router := protocol.NewRouter()

	authService := auth.NewService()

	authHandler :=
		auth.NewHandler(
			authService,
		)

	router.Register(
		"auth_identify",
		authHandler.Identify,
	)

	router.Register(
		"auth_verify",
		authHandler.Verify,
	)

	router.RegisterDebugHandlers()

	return &Server{
		addr:        addr,
		router:      router,
		clients:     session.NewManager(),
		authService: authService,
	}
}

func (s *Server) Start() error {

	http.HandleFunc(
		"/ws",
		s.handleWebSocket,
	)

	log.Printf(
		"Ether Backend listening on %s",
		s.addr,
	)

	log.Printf(
		"WebSocket endpoint: ws://%s/ws",
		s.addr,
	)

	return http.ListenAndServe(
		s.addr,
		nil,
	)
}

func (s *Server) handleWebSocket(
	w http.ResponseWriter,
	r *http.Request,
) {

	conn, err :=
		upgrader.Upgrade(
			w,
			r,
			nil,
		)

	if err != nil {
		log.Printf(
			"websocket upgrade failed: %v",
			err,
		)

		return
	}

	client :=
		session.NewClient(conn)

	s.clients.Add(client)

	log.Printf(
		"client connected: %s",
		client.ID,
	)

	defer func() {

		s.clients.Remove(
			client.ID,
		)

		log.Printf(
			"client disconnected: %s",
			client.ID,
		)

		conn.Close()
	}()

	for {

		messageType, data, err :=
			conn.ReadMessage()

		if err != nil {
			return
		}

		if messageType != gws.TextMessage {
			continue
		}

		message, err :=
			protocol.Decode(data)

		if err != nil {

			log.Printf(
				"invalid packet from %s: %v",
				client.ID,
				err,
			)

			continue
		}

		if err := s.router.Handle(
			client,
			message,
		); err != nil {

			log.Printf(
				"packet rejected from %s: %v",
				client.ID,
				err,
			)

			continue
		}
	}
}