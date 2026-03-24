package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
	"github.com/cybozu/neco-containers/spiffe-test/internal/service"
)

type HelloResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Handler struct {
	authenticator auth.Authenticator
	service       service.HelloService
}

func NewHandler(authenticator auth.Authenticator, svc service.HelloService) *Handler {
	return &Handler{
		authenticator: authenticator,
		service:       svc,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hello", h.handleHello)
	mux.HandleFunc("/health", h.handleHealth)
}

func (h *Handler) handleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	callerID, err := h.authenticator.GetCallerID(r)
	if err != nil {
		slog.Warn("Authentication failed", "error", err)
		if errors.Is(err, auth.ErrNoClientCert) || errors.Is(err, auth.ErrInvalidToken) {
			h.writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		h.writeError(w, http.StatusForbidden, "access denied")
		return
	}

	slog.Info("Request received", "callerID", callerID.String())

	message, err := h.service.SayHello(r.Context(), callerID)
	if err != nil {
		slog.Warn("Service error", "error", err, "callerID", callerID.String())
		if errors.Is(err, service.ErrUnauthorized) {
			h.writeError(w, http.StatusForbidden, "access denied")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.writeJSON(w, http.StatusOK, HelloResponse{Message: message})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, HelloResponse{Error: message})
}
