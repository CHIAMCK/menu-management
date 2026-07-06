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

## Quick Start

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
go mod tidy
go run .
```

The server starts on `http://localhost:8080` (override with `PORT` env var).

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/`  | Hello World |

```bash
curl http://localhost:8080/
```

## Project Structure

```
.
├── main.go
├── internal/
│   ├── db/
│   │   ├── migrations/   # SQL schema migrations
│   │   ├── postgres.go   # Connection pool
│   │   └── migrate.go    # Migration runner
│   ├── models/           # Domain structs matching DB tables
│   └── routes/
└── README.md
```
