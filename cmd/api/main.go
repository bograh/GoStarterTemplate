package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kovadelivery.com/internal/auth"
	"kovadelivery.com/internal/cache"
	"kovadelivery.com/internal/config"
	"kovadelivery.com/internal/db"
	apphttp "kovadelivery.com/internal/http"
	"kovadelivery.com/internal/mq"
	"kovadelivery.com/internal/ws"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.NewPostgresDB(
		cfg.GetDSN(),
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
	)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("connected to PostgreSQL")

	redisCache, err := cache.NewRedis(
		cfg.GetRedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisCache.Close()
	log.Println("connected to Redis")

	sessionManager := auth.NewSessionManager(
		redisCache,
		cfg.Session.Duration,
		cfg.Session.RefreshThreshold,
	)

	authService := auth.NewService(
		database.DB,
		sessionManager,
		cfg.Security.BCryptCost,
	)

	producer := mq.NewProducer(redisCache.Client(), "delivery-events")

	consumer := mq.NewConsumer(
		redisCache.Client(),
		"delivery-events",
		"delivery-consumer-group",
		"worker-1",
	)

	go func() {
		if err := consumer.Start(context.Background(), handleEvent); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	wsHub := ws.NewHub()
	go wsHub.Run()

	router := apphttp.NewRouter(cfg, authService, sessionManager, producer, wsHub)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}

func handleEvent(event mq.Event) error {
	log.Printf("processing event: %s with payload: %v", event.Type, event.Payload)
	return nil
}
