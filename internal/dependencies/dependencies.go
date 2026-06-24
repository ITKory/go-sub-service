package dependencies

import (
	"log/slog"

	"subscription-service/internal/config"
	"subscription-service/internal/handler"
	"subscription-service/internal/repository"
	"subscription-service/internal/server"
	"subscription-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func WireDependencies(pool *pgxpool.Pool, cfg *config.Config, logger *slog.Logger) *server.Server {
	repo := repository.NewPostgresSubscriptionRepository(pool)
	svc := service.NewSubscriptionService(repo, logger)
	h := handler.NewSubscriptionHandler(svc, logger)
	return server.NewServer(h, cfg.Server.Port, logger)
}
