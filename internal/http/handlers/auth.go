package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kovadelivery.com/internal/auth"
	"kovadelivery.com/internal/config"
	"kovadelivery.com/internal/models"
	"kovadelivery.com/internal/mq"
	"kovadelivery.com/pkg/utils"
)

type AuthHandler struct {
	authService *auth.Service
	config      *config.Config
	producer    *mq.Producer
}

func NewAuthHandler(authService *auth.Service, cfg *config.Config, producer *mq.Producer) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		config:      cfg,
		producer:    producer,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSONResponse(w, user.ToResponse(), http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, sessionID, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.setSessionCookie(w, sessionID)

	h.producer.Publish(r.Context(), mq.Event{
		Type: mq.EventUserLoggedIn,
		Payload: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})

	utils.JSONResponse(w, map[string]interface{}{
		"user":    user.ToResponse(),
		"message": "login successful",
	}, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.ErrorResponse(w, "no session found", http.StatusBadRequest)
		return
	}

	if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
		utils.ErrorResponse(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	h.clearSessionCookie(w)

	utils.JSONResponse(w, map[string]string{
		"message": "logout successful",
	}, http.StatusOK)
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Domain:   h.config.Cookie.Domain,
		MaxAge:   int(h.config.Session.Duration.Seconds()),
		Secure:   h.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Domain:   h.config.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}
