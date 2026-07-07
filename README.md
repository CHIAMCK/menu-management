# Menu Management API

A Go app using [Gin](https://github.com/gin-gonic/gin) and PostgreSQL for menu management and order placement.

## Requirements

- Go 1.21+
- PostgreSQL 14+

## Data Model

```
user (1) ──< orders (many) >── merchant (1)
merchant (1) ──< menus (many, only 1 ACTIVE per merchant)
menu (1) ──< categories (many)
category (1) ──< items (many, denormalized merchant_id)
order (1) ──< order_items (many) >── items (many)
```

- One merchant can have many menus, but only **one active menu** at a time (enforced by partial unique index).
- One menu has many categories; each category has many items.
- `items.merchant_id` is denormalized for fast order validation; set it in application code from `category → menu → merchant`.
- One order has many items via `order_items`; duplicate items in the same order are prevented by `UNIQUE (order_id, item_id)`.

## Database Schema

| Table         | Key columns |
|---------------|-------------|
| `users`       | `name`, `email` |
| `merchants`   | `name`, `status` (ACTIVE/INACTIVE) |
| `menus`       | `merchant_id`, `name`, `status` (ACTIVE/INACTIVE) |
| `categories`  | `menu_id`, `name`, `status` |
| `items`       | `merchant_id`, `category_id`, `name`, `price`, `status`, `availability` |
| `orders`      | `user_id`, `merchant_id`, `status`, `total_amount` |
| `order_items` | `order_id`, `item_id`, `quantity`, `unit_price` (unique per order) |

Migrations live in `internal/db/migrations/` and run automatically on startup.

`000002_seed_data.up.sql` loads mock merchants, users, menus, categories, items, and orders.

**Merchants**
| ID | Name | Status |
|----|------|--------|
| 1 | Joe's Pizza | ACTIVE |
| 2 | Cafe Bloom | ACTIVE |
| 3 | Sushi Zen | INACTIVE |

**Users:** Alice, Bob, Carol

**Menus:** Joe's Pizza has an active "Main Menu" + inactive "Winter Specials"; Cafe Bloom has "All Day Menu"

**Orders:** 3 sample orders across merchants (COMPLETED, PENDING, CONFIRMED)

The active menu is scoped to the merchant configured via the `MERCHANT_ID` environment variable (default `1` in Docker Compose).

```bash
curl http://localhost:8080/v1/menu
```

To reset and re-apply migrations: `docker compose down -v && docker compose up --build`

## Quick Start (Docker)

```bash
docker compose up --build
```

- App: http://localhost:8080
- PostgreSQL: `localhost:5432` (user `postgres`, password `postgres`, db `menu_management`)

```bash
curl http://localhost:8080/
curl http://localhost:8080/v1/menu
```

Stop and remove containers:

```bash
docker compose down
```

Remove the database volume as well:

```bash
docker compose down -v
```

## Quick Start (Local)

**1. Start PostgreSQL** (example with Docker):

```bash
docker run --name menu-pg \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=menu_management \
  -p 5432:5432 \
  -d postgres:16
```

**2. Run the app:**

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/menu_management?sslmode=disable"
export MERCHANT_ID=1
go mod tidy
go run .
```

The server starts on `http://localhost:8080` (override with `PORT` env var). Set `MERCHANT_ID` to choose which merchant's active menu is served.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/`  | Hello World |
| GET    | `/v1/menu` | Active menu for the merchant configured via `MERCHANT_ID` |
| GET    | `/v1/menu/items/{id}` | Single menu item by ID |

```bash
curl http://localhost:8080/
curl http://localhost:8080/v1/menu
curl http://localhost:8080/v1/menu/items/1
```

## Project Structure

```
.
├── main.go
├── Dockerfile
├── docker-compose.yml
├── internal/
│   ├── db/
│   │   ├── migrations/   # SQL schema migrations
│   │   ├── postgres.go   # Connection pool
│   │   └── migrate.go    # Migration runner
│   ├── models/           # Domain structs matching DB tables
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic layer
│   ├── handlers/         # HTTP handlers
│   └── routes/
└── README.md
```
