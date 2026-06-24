# Subscription Service

REST API service for managing subscriptions with total sum calculation.

## 📖 API Documentation

**Swagger UI**: http://xgw8so00sggcg88ogok0s80c.95.79.96.242.sslip.io/swagger/index.html

## 🚀 Quick Start

```bash
git clone https://github.com/ITKory/go-sub-service.git
cd go-sub-service
docker compose up -d
```

Service runs on `http://localhost:8080`

## 📡 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/subscriptions` | Create subscription |
| GET | `/subscriptions` | Get all subscriptions |
| GET | `/subscriptions/{id}` | Get subscription by ID |
| PUT | `/subscriptions/{id}` | Update subscription |
| DELETE | `/subscriptions/{id}` | Delete subscription |
| POST | `/subscriptions/sum` | Calculate total sum |
| GET | `/health` | Health check |

## 🏗 Project Structure

```
subscription-service/
├── cmd/app/
│   └── main.go                    # Application entry point
├── docs/
│   ├── docs.go                    # Swagger docs
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── apperrors/
│   │   └── errors.go              # Custom errors
│   ├── config/
│   │   └── config.go              # Configuration
│   ├── database/
│   │   ├── migrate.go             # DB migrations
│   │   └── postgres.go            # PostgreSQL connection
│   ├── dependencies/
│   │   └── dependencies.go        # DI container
│   ├── handler/
│   │   └── subscription_handler.go # HTTP handlers
│   ├── model/
│   │   ├── dto.go                 # Data transfer objects
│   │   └── subscription.go        # Domain model
│   ├── repository/
│   │   └── subscription_repo.go   # DB repository
│   ├── server/
│   │   └── server.go              # HTTP server setup
│   └── service/
│       └── subscription_service.go # Business logic
├── migrations/                     # SQL migration files
├── .env                            # Environment variables
├── .env.example                    # Example env file
├── config.yaml                     # App configuration
├── docker-compose.yml              # Docker Compose (production)
├── docker-compose.local.yml        # Docker Compose (local)
├── Dockerfile                      # Docker image
├── go.mod                          # Go module
└── go.sum                          # Go dependencies
```

## ⚙️ Configuration

**config.yaml**:
```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "your_password"
  dbname: "subscriptions"
  sslmode: "disable"
```

**.env**:
```env
SERVER_PORT=8080
```

## 🛠 Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/app/main.go

# Regenerate Swagger docs
swag init -g cmd/app/main.go -o docs
```

## 📝 Tech Stack

- Go 1.26
- PostgreSQL
- Docker & Docker Compose
- Swagger (swaggo)
- Structured logging (slog)
- Self-hosted by Coolify
