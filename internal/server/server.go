package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscription-service/internal/handler"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func NewServer(handler *handler.SubscriptionHandler, port string, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /subscriptions", handler.CreateSubscription)
	mux.HandleFunc("GET /subscriptions/{id}", handler.GetSubscription)
	mux.HandleFunc("PUT /subscriptions/{id}", handler.UpdateSubscription)
	mux.HandleFunc("GET /subscriptions", handler.ListSubscriptions)
	mux.HandleFunc("DELETE /subscriptions/{id}", handler.DeleteSubscription)
	mux.HandleFunc("POST /subscriptions/sum", handler.CalculateTotalSum)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	if port == "" {
		port = "8080"
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) swaggerURL() string {
	if url := os.Getenv("COOLIFY_URL"); url != "" {
		return url + "/swagger/index.html"
	}

	if fqdn := os.Getenv("COOLIFY_FQDN"); fqdn != "" {
		return "https://" + fqdn + "/swagger/index.html"
	}

	return fmt.Sprintf("http://localhost%s/swagger/index.html", s.httpServer.Addr)
}

func (s *Server) Start() error {
	s.logger.Info("server starting",
		"addr", s.httpServer.Addr,
	)

	s.logger.Info("swagger ui available",
		"url", s.swaggerURL(),
	)

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-quit:
	}

	s.logger.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}
