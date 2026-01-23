package http

import (
	"net/http"

	"kovadelivery.com/internal/auth"
	"kovadelivery.com/internal/config"
	"kovadelivery.com/internal/http/handlers"
	"kovadelivery.com/internal/http/middleware"
	"kovadelivery.com/internal/mq"
	"kovadelivery.com/internal/ws"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router *chi.Mux
}

func NewRouter(
	cfg *config.Config,
	authService *auth.Service,
	sessionManager *auth.SessionManager,
	producer *mq.Producer,
	wsHub *ws.Hub,
) *Router {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.NewCORS(cfg.CORS.AllowedOrigins, cfg.CORS.AllowCredentials))

	rateLimiter := middleware.NewRateLimiter(
		cfg.Security.RateLimitRequests,
		cfg.Security.RateLimitWindow,
	)

	authMiddleware := middleware.NewAuthMiddleware(sessionManager, "session_id")

	authHandler := handlers.NewAuthHandler(authService, cfg, producer)
	userHandler := handlers.NewUserHandler(authService)
	wsHandler := ws.NewHandler(wsHub, sessionManager)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Use(rateLimiter.Limit)
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/me", userHandler.GetMe)
		})

		r.Route("/ws", func(r chi.Router) {
			r.Get("/", wsHandler.HandleWebSocket)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &Router{router: r}
}

func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ro.router.ServeHTTP(w, r)
}
