# Cinema Booking System

A small **cinema seat booking** service with a Go HTTP API, **Redis** for holds and confirmations, and a **single-page static frontend**. Users pick a movie, reserve a seat with a time-limited hold, then confirm or release it.

---

## Features

- Browse movies and cinema layouts (rows × seats per movie).
- **Hold** a seat with automatic TTL (~2 minutes); only one active hold per client session in the UI.
- **Confirm** to finalize a booking or **release** the hold.
- Live seat map polling so multiple browsers see holds from others.
- Session ownership checks (`user_id` must match the Redis-backed session).

---

## Tech Stack

| Layer | Choice |
|--------|--------|
| Language | Go **1.26** |
| HTTP | `net/http` with Go **1.22+** route patterns (`{movieID}`, `{seatID}`, …) |
| Storage | **Redis** (`github.com/redis/go-redis/v9`) |
| Frontend | Vanilla HTML/CSS/JS under `static/` |

---

## Repository Layout

```
CinemaBookingSystem/
├── cmd/main.go              # Entry point: mux, routes, static files, Redis wiring
├── internal/
│   ├── adapters/redis.go   # redis client + ping on startup
│   ├── booking/            # domain, service, HTTP handler, Redis store
│   └── utils/              # shared JSON helpers
├── static/index.html        # Web UI
├── docker-compose.yaml      # Redis + Redis Commander (optional UI for Redis)
├── Makefile                 # run, down, test
├── go.mod
└── README.md
```

---

## Prerequisites

- **Go** 1.26+ (see `go.mod`)
- **Docker** (and Docker Compose) for local Redis, *or* a Redis instance on **`localhost:6379`**

The app **exits on startup** if it cannot connect to Redis (`adapters.NewClient` pings the server).

---

## Quick Start

### 1. Start Redis

```bash
docker compose up -d
```

This starts **Redis** on port **6379** and **Redis Commander** on **8081** (optional, for inspecting keys).

### 2. Run the application

```bash
make run
```

Or manually:

```bash
docker compose up -d
go run ./cmd
```

The server listens on **http://localhost:8080**.

- **Web UI:** open **http://localhost:8080/** (served from `static/`).
- **API base URL:** `http://localhost:8080`

### 3. Stop infrastructure

```bash
make down
```

Stops Docker Compose services. The Go process, if run in the foreground, stop with `Ctrl+C`.

---

## Make Targets

| Target | Description |
|--------|-------------|
| `make run` | `docker compose up -d` then `go run ./cmd` |
| `make down` | `docker compose down` |
| `make test` | `go test ./... -v -count=1` (tests expect Redis on `localhost:6379`) |

---

## HTTP API

All JSON responses use **UTF-8** with `Content-Type: application/json; charset=utf-8` where applicable.

### Movies

#### `GET /movies`

Returns the catalog configured in `cmd/main.go`.

**Response** (`200 OK`): array of objects:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Movie slug (e.g. `limitless`) |
| `title` | string | Display title |
| `rows` | int | Number of seat rows |
| `seats_per_row` | int | Seats per row |

---

### Seats

#### `GET /movies/{movieID}/seats`

Returns **occupied** seats only (held or confirmed). Absence of a seat in this list means “available” on the client.

**Response** (`200 OK`): array of:

| Field | Type | Description |
|-------|------|-------------|
| `seat_id` | string | e.g. `A1`, `B3` (row letter + number) |
| `user_id` | string | Holder’s client id |
| `booked` | bool | Always `true` for entries in this list |
| `confirmed` | bool | `true` if status is confirmed |

---

### Hold a seat

#### `POST /movies/{movieID}/seats/{seatID}/hold`

**Body (JSON):**

```json
{ "user_id": "<opaque client id>" }
```

**Response** (`201 Created`):

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Use for confirm / release |
| `movie_id` | string | Same as `{movieID}` |
| `seat_id` | string | Same as `{seatID}` |
| `expires_at` | string | RFC3339 expiry time for the hold |

If the seat is already taken, booking fails (see server logs; errors may not always be returned as JSON depending on handler paths).

---

### Confirm booking

#### `PUT /sessions/{sessionID}/confirm`

**Body (JSON):**

```json
{ "user_id": "<same user_id as hold>" }
```

**Response** (`200 OK`): session details including updated `status`.

**Response** (`403 Forbidden`): `{"error":"not your session"}` if `user_id` does not match the session owner.

---

### Release hold

#### `DELETE /sessions/{sessionID}`

**Body (JSON):**

```json
{ "user_id": "<same user_id as hold>" }
```

**Response** (`204 No Content`) on success.

**Response** (`403 Forbidden`): same as confirm when `user_id` does not match.

---

## Redis Model (overview)

| Key pattern | Role |
|-------------|------|
| `seat:{movieID}:{seatID}` | JSON booking payload; **NX** set on hold; TTL while held |
| `session:{sessionID}` | Points to the seat key for lookup |

Confirm removes TTL (persistent confirmed booking) as implemented in `RedisStore`. Release deletes the relevant keys.

---

## Configuration Notes

- Redis address is currently **hardcoded** as `localhost:6379` in `cmd/main.go` (`adapters.NewClient`).
- Movies are **hardcoded** in `cmd/main.go`; extend `movies` and redeploy to change the catalog.

---

## Testing

```bash
make test
```

The concurrent booking test uses **live Redis** on `localhost:6379` and asserts that only one goroutine succeeds booking the same seat under contention.

---

## License / Contributing

Not specified in-repo; add a license file if you open-source the project.
