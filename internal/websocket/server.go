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
	return NewServerWithService(
		addr,
		auth.NewService(),
	)
}

func NewServerWithService(
	addr string,
	authService *auth.Service,
) *Server {

	router := protocol.NewRouter()

	authHandler :=
		auth.NewHandler(
			authService,
		)

	router.Register(
		"auth_hello",
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

// DEV SERVER
func NewDevelopmentServer(addr string) *Server {
	return NewServerWithService(
		addr,
		auth.NewServiceWithVerifier(
			auth.NewDevelopmentVerifier(),
		),
	)
}

// Handler returns the HTTP handler used by the backend.
//
// This is useful for integration tests and also gives us a clean
// boundary for eventually putting the backend behind nginx.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/ws",
		s.handleWebSocket,
	)

	return mux
}

func (s *Server) Start() error {

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
		s.Handler(),
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
