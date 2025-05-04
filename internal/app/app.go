package app

import (
	"context"
	"database/sql"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	_ "auth-service/docs" // Import generated docs
	"auth-service/internal/app/config"
	"auth-service/internal/clients/kafka"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/repositories/user"
	"auth-service/internal/services/auth"
)

func Run(envFiles ...string) {
	cfg, err := config.New(envFiles...)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DB.URL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	kafkaProducer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.UserTopic)
	defer kafkaProducer.Close()

	userRepo := user.NewRepository(db)

	authService := auth.NewAuthService(userRepo, kafkaProducer, cfg.JWT.Secret, cfg.Email.From)

	authHandler := handlers.NewAuthHandler(authService)

	router := mux.NewRouter()

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.Use(middleware.JWTAuth(cfg.JWT.Secret))

	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/verify", authHandler.VerifyCode).Methods("POST")
	router.HandleFunc("/api/auth/logout", authHandler.Logout).Methods("POST")

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on %s", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
