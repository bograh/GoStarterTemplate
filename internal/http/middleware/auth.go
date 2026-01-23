package middleware

import (
	"context"
	"net/http"

	"kovadelivery.com/internal/auth"
	"kovadelivery.com/pkg/utils"
)

type contextKey string

const (
	SessionKey contextKey = "session"
	UserIDKey  contextKey = "userID"
)

type AuthMiddleware struct {
	sessionManager *auth.SessionManager
	cookieName     string
}

func NewAuthMiddleware(sessionManager *auth.SessionManager, cookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
		cookieName:     cookieName,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.cookieName)
		if err != nil {
			utils.ErrorResponse(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sessionID := cookie.Value
		session, err := m.sessionManager.GetSession(r.Context(), sessionID)
		if err != nil {
			utils.ErrorResponse(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), SessionKey, session)
		ctx = context.WithValue(ctx, UserIDKey, session.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
