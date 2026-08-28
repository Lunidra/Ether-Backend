package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	gws "github.com/gorilla/websocket"

	"github.com/Lunidra/Ether-Backend/internal/auth"
	"github.com/Lunidra/Ether-Backend/internal/broadcast"
	"github.com/Lunidra/Ether-Backend/internal/chat"
	"github.com/Lunidra/Ether-Backend/internal/presence"
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
	addr            string
	router          *protocol.Router
	clients         *session.Manager
	authService     *auth.Service
	presenceManager *presence.Manager
	broadcast       *broadcast.Service
	//presenceHandler *presence.Handler
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

	clients := session.NewManager()

	broadcastService := broadcast.NewService(
		clients,
	)

	presenceManager := presence.NewManager(
		clients,
		broadcastService,
	)

	authHandler := auth.NewHandler(
		authService,
		func(client *session.Client) {
			if err := presenceManager.Join(client); err != nil {
				log.Printf(
					"failed to broadcast user join: %v",
					err,
				)
			}
		},
	)

	router.Register(
		"auth_hello",
		authHandler.Identify,
	)

	router.Register(
		"auth_verify",
		authHandler.Verify,
	)

	presenceHandler := presence.NewHandler(
		presenceManager,
	)

	router.Register(
		"presence_request",
		presenceHandler.Request,
	)

	router.Register(
		"presence",
		presenceHandler.Update,
	)

	chatHandler := chat.NewHandler(
		broadcastService,
	)

	router.Register(
		"chat",
		chatHandler.Handle,
	)

	router.RegisterDebugHandlers()

	return &Server{
		addr:            addr,
		router:          router,
		clients:         clients,
		authService:     authService,
		presenceManager: presenceManager,
		broadcast:       broadcastService,
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

	mux.HandleFunc("/dev/broadcast", s.handleDevBroadcast)

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

		if client.Authenticated {
			if err := s.presenceManager.Leave(client); err != nil {
				log.Printf(
					"failed to broadcast user leave: %v",
					err,
				)
			}
		}

		s.clients.Remove(client.ID)

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

// Broadcast sends an Ether broadcast packet to every authenticated client.
//
// This is intentionally server-side only. It is intended for future
// administrative commands, console commands, or an HTTP administration API.
// Broadcast sends an Ether broadcast to every authenticated client.
func (s *Server) Broadcast(
	sender string,
	message string,
) error {
	return s.broadcast.Broadcast(
		sender,
		message,
	)
}

//BROADCAST FUNCTION

type devBroadcastRequest struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

func (s *Server) handleDevBroadcast(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var request devBroadcastRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid JSON",
			http.StatusBadRequest,
		)
		return
	}

	if err := s.Broadcast(
		request.Sender,
		request.Message,
	); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(
		`{"success":true}`,
	))
}
