package handlers

import (
	"net/http"

	"kovadelivery.com/internal/auth"
	"kovadelivery.com/internal/http/middleware"
	"kovadelivery.com/pkg/utils"
)

type UserHandler struct {
	authService *auth.Service
}

func NewUserHandler(authService *auth.Service) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.ErrorResponse(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.ErrorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	utils.JSONResponse(w, user.ToResponse(), http.StatusOK)
}
