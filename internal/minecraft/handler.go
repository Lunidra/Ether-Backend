package minecraft

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
	secret  string
}

func NewHandler(service *Service, secret string) *Handler {
	return &Handler{
		service: service,
		secret:  secret,
	}
}

type linkRequest struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type unlinkRequest struct {
	UUID string `json:"uuid"`
}

type linkResponse struct {
	Success bool `json:"success"`
	Linked  bool `json:"linked"`
}

type unlinkResponse struct {
	Success  bool `json:"success"`
	Unlinked bool `json:"unlinked"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (h *Handler) Link(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Success: false,
			Error:   "unauthorized",
		})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var request linkRequest

	decoder := json.NewDecoder(
		http.MaxBytesReader(
			w,
			r.Body,
			4096,
		),
	)

	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if err := h.service.Link(
		request.UUID,
		request.Username,
	); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	log.Printf(
		"[Minecraft] Linked %s (%s)",
		request.Username,
		request.UUID,
	)

	writeJSON(w, http.StatusOK, linkResponse{
		Success: true,
		Linked:  true,
	})
}

func (h *Handler) Unlink(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Success: false,
			Error:   "unauthorized",
		})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var request unlinkRequest

	decoder := json.NewDecoder(
		http.MaxBytesReader(
			w,
			r.Body,
			4096,
		),
	)

	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	unlinked := h.service.Unlink(request.UUID)
	if unlinked {
		log.Printf(
			"[Minecraft] Unlinked %s",
			request.UUID,
		)
	}

	writeJSON(w, http.StatusOK, unlinkResponse{
		Success:  true,
		Unlinked: unlinked,
	})
}

func (h *Handler) authorized(r *http.Request) bool {
	if h.secret == "" {
		return false
	}

	header := r.Header.Get("Authorization")

	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return false
	}

	token := strings.TrimSpace(
		strings.TrimPrefix(header, prefix),
	)

	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(token),
		[]byte(h.secret),
	) == 1
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

type statusResponse struct {
	Success  bool   `json:"success"`
	Linked   bool   `json:"linked"`
	UUID     string `json:"uuid,omitempty"`
	Username string `json:"username,omitempty"`
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Success: false,
			Error:   "unauthorized",
		})
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))

	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Success: false,
			Error:   "missing Minecraft UUID",
		})
		return
	}

	link, linked := h.service.Get(uuid)

	if !linked {
		writeJSON(w, http.StatusOK, statusResponse{
			Success: true,
			Linked:  false,
		})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{
		Success:  true,
		Linked:   true,
		UUID:     link.UUID,
		Username: link.Username,
	})
}
