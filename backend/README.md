# Personal Finance Tracker - Backend

Backend API for the Personal Finance Tracker mobile app.

## Tech Stack

- **Language**: Golang
- **Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration loading
│   ├── domain/
│   │   └── transaction.go    # Transaction entity
│   ├── handler/
│   │   ├── webhook.go        # Webhook for iOS app
│   │   └── analytics.go      # Dashboard endpoints
│   ├── service/
│   │   └── transaction.go    # Business logic
│   ├── repository/
│   │   └── transaction.go    # Database operations
│   └── middleware/
│   │       └── apikey.go     # API key validation
├── migrations/
│   └── 000001_transactions.up.sql
├── Dockerfile
├── go.mod
└── .env.example
```

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Docker (optional)

### Local Development

1. **Copy environment variables:**
   ```bash
   cp backend/.env.example .env
   ```

2. **Update `.env` with your values:**
   ```bash
   API_KEY=your-secret-api-key-here
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=finance
   DB_PASSWORD=finance
   DB_NAME=finance_tracker
   ```

3. **Install dependencies:**
   ```bash
   cd backend
   go mod download
   ```

4. **Run the application:**
   ```bash
   go run cmd/api/main.go
   ```

### Using Docker Compose

```bash
# Start all services (PostgreSQL + Backend)
docker-compose -f deploy/docker-compose.yaml up -d

# View logs
docker-compose -f deploy/docker-compose.yaml logs -f

# Stop services
docker-compose -f deploy/docker-compose.yaml down
```

Or use the Makefile from the backend directory:
```bash
cd backend
make docker-run
make docker-stop
```

## API Endpoints

### Webhook (iOS App → Backend)

Requires `X-API-Key` header.

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/webhook/transaction` | Create single transaction |
| POST | `/api/v1/webhook/transactions/batch` | Create batch transactions |

### Analytics (Dashboard)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/analytics/summary` | Total in/out, balance |
| GET | `/api/v1/analytics/trends?period=daily` | Trends (daily/weekly/monthly) |
| GET | `/api/v1/analytics/by-source` | Breakdown by bank/wallet |
| GET | `/api/v1/analytics/by-category` | Breakdown by category |

### Transactions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/transactions` | List with pagination |
| GET | `/api/v1/transactions/:id` | Get single transaction |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health check |

## Example Requests

### Create Transaction (Webhook)

```bash
curl -X POST http://localhost:8080/api/v1/webhook/transaction \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{
    "amount": 100.50,
    "type": "out",
    "category": "Food",
    "description": "Lunch",
    "source": "Bank ABC",
    "source_account": "1234****5678",
    "transaction_date": "2026-01-15T12:00:00Z"
  }'
```

### Get Summary

```bash
curl http://localhost:8080/api/v1/analytics/summary
```

## Deployment

### Kubernetes

```bash
# Apply all manifests
kubectl apply -f deploy/k8s/

# Check deployment status
kubectl get pods -l app=finance-tracker
```

### GitHub Actions

The CI/CD pipeline automatically:
1. Lints code
2. Runs tests
3. Builds Docker image
4. Pushes to container registry
5. Deploys to Kubernetes (on main branch)
