# Menu Management API

A Go app using [Gin](https://github.com/gin-gonic/gin) and PostgreSQL for menu management and order placement.

## Requirements

- Go 1.21+
- PostgreSQL 14+

## Quick Start

### Docker

```bash
docker compose up --build
```

| Service | Address |
|---------|---------|
| API | http://localhost:8080 |
| PostgreSQL | `localhost:5432` |
| RabbitMQ | `localhost:5672` |
| RabbitMQ UI | http://localhost:15672 |

Verify with `curl http://localhost:8080/`. See [Endpoints](#endpoints) for more examples. After placing an order, check kitchen-display logs with `docker compose logs app`.

```bash
docker compose down      # stop containers
docker compose down -v   # stop and remove database volume
```

### Local

Start dependencies (Docker examples):

```bash
# PostgreSQL
docker run --name menu-pg \
  -e POSTGRES_PASSWORD \
  -e POSTGRES_DB=menu_management \
  -p 5432:5432 -d postgres:16

# RabbitMQ
docker run --name menu-rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  -d rabbitmq:3-management-alpine
```

Run the app:

```bash
cp .env.example .env
# Set DATABASE_URL and RABBITMQ_URL in .env
go mod tidy && go run .
```

Server listens on http://localhost:8080 (`PORT` to override). `MERCHANT_ID` selects which merchant's menu is served.

## Order Events

When `POST /v1/orders` succeeds, the API publishes a durable JSON message to the `order.placed` queue (configurable via `ORDER_QUEUE_NAME`). Publishing is best-effort: if RabbitMQ is unavailable, the order is still created and the API returns `201 Created`.

A background goroutine in the app process consumes those messages and logs a structured JSON summary simulating a kitchen display notification.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/` | Hello World |
| GET    | `/v1/menu` | Active menu for the merchant configured via `MERCHANT_ID` |
| GET    | `/v1/menu/items/{id}` | Single menu item by ID |
| PATCH  | `/v1/menu/items/{id}` | Update item availability (`AVAILABLE` or `OUT_OF_STOCK`) |
| POST   | `/v1/orders` | Create an order |
| GET    | `/v1/orders/{id}` | Get order by ID |
| PATCH  | `/v1/orders/{id}/status` | Update order status (`RECEIVED`, `PREPARING`, `READY`, or `COMPLETED`) |

```bash
# Health check
curl http://localhost:8080/

# Menu
curl http://localhost:8080/v1/menu
curl http://localhost:8080/v1/menu/items/1

# Update item availability
curl -X PATCH http://localhost:8080/v1/menu/items/1 \
  -H "Content-Type: application/json" \
  -d '{"availability":"OUT_OF_STOCK"}'

# Orders
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"merchant_id":1,"items":[{"item_id":1,"quantity":2}]}'

curl http://localhost:8080/v1/orders/1

curl -X PATCH http://localhost:8080/v1/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"PREPARING"}'
```

## Unit Tests

Run all unit tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

## Design Questions

### 1. API contract decisions

What was one non-obvious design decision you made in the API surface — a naming choice, a response shape, a status code — and why did you make it?

One non-obvious API contract decision I made was returning HTTP 409 Conflict when detecting duplicate order submissions.

To prevent duplicate orders caused by users tapping the "Place Order" button multiple times, I temporarily lock order creation for the same user for a short period. If another create order request arrives while an order is already being processed, the API returns 409 Conflict

### 2. Versioning

If a mobile client were already consuming `GET /menu` and you needed to change the response shape in a breaking way, how would you handle that?

I will implement API versioning. Add a new endpoint, `GET /v2/menu` with the new response shape. Migrate the endpoint on mobile client to v2 on new app version, old version still calling v1. Monitor app version adoption rate and deprecate v1 when all clients have upgraded.

### 3. What you'd do differently with more time

Name one thing you would change or add if you had another two hours. Be specific.

### 4. Production gap

What is the most significant thing missing from this service that would concern you before shipping it to real users?

The most significant production gap is authentication and authorization. Without authentication, anyone can place orders or access merchant APIs. Without authorization, any authenticated user could potentially modify item availability or change order status, leading to security issues.
