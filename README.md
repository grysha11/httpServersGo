# Chirpy Backend API (httpServersGo)

A robust, modular RESTful API built in Go. This serves as the backend for "Chirpy" (a social microblogging platform), featuring user authentication, post (chirp) creation, and admin metrics, all backed by a PostgreSQL database.

## Features

* **User Management:** Secure user registration and updates using Argon2id password hashing.
* **Authentication:** Robust JWT-based authentication with short-lived access tokens and long-lived refresh tokens.
* **Chirps (Posts):** Users can create, view, and delete their chirps. Includes an automatic profanity filter.
* **Webhooks:** Integration with third-party services (Polka) to upgrade users to "Chirpy Red" status.
* **Admin Dashboard:** Metrics tracking and database reset capabilities for development environments.

## Tech Stack

* **Language:** Go (Golang).
* **Database:** PostgreSQL.
* **Database Tools:** * [Goose](https://github.com/pressly/goose) for database migrations.
    * [SQLC](https://sqlc.dev/) for type-safe database query generation.
* **Authentication:** `golang-jwt/jwt/v5` and `alexedwards/argon2id`.

## Project Structure

This project follows a domain-driven, standard Go application layout:

`├── cmd/server/` - Application entrypoint
`├── internal/` - Private application code
`│   ├── auth/` - Hashing and JWT logic
`│   ├── config/` - Shared application state (DB pool, secrets)
`│   ├── database/` - SQLC-generated database access code
`│   ├── handler/` - HTTP handlers grouped by domain (users, chirps, etc.)
`│   ├── middleware/` - Reusable HTTP request filters (auth, metrics)
`│   └── router/` - API route definitions
`├── sql/` - Database schemas and queries for SQLC/Goose
`└── index.html` - Static frontend entry file

## Local Development Setup

### 1. Prerequisites
* Go 1.22 or higher installed.
* PostgreSQL installed and running.

### 2. Environment Variables
Create a `.env` file in the root directory with the following variables:
```env
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
JWT_SECRET="your-super-secret-jwt-key"
POLKA_KEY="your-polka-webhook-api-key"

```

### 3. Database Setup

Ensure your PostgreSQL database is running, then apply the migrations using Goose:

```bash
cd sql/schema
goose postgres <YOUR_DB_URL> up

```

### 4. Running the Server

Start the server from the root of the project:

```bash
go run cmd/server/main.go

```

The server will start on `http://localhost:8080`.

## 🗺️ API Endpoints Reference

### Public Routes

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/api/healthz` | Check API health |
| `GET` | `/api/chirps` | Get all chirps (supports `?author_id=` and `?sort=desc`) |
| `GET` | `/api/chirps/{id}` | Get a specific chirp by ID |
| `POST` | `/api/users` | Register a new user |
| `POST` | `/api/login` | Authenticate user and get tokens |
| `POST` | `/api/refresh` | Exchange a refresh token for a new access token |
| `POST` | `/api/revoke` | Revoke a refresh token |

### Protected Routes (Requires Bearer Token)

| Method | Endpoint | Description |
| --- | --- | --- |
| `PUT` | `/api/users` | Update user email/password |
| `POST` | `/api/chirps` | Create a new chirp |
| `DELETE` | `/api/chirps/{id}` | Delete a chirp (must be the author) |

### Webhooks & Admin

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/api/polka/webhooks` | Upgrade user account (Requires `ApiKey` header) |
| `GET` | `/admin/metrics` | View server hit count (HTML) |
| `POST` | `/admin/reset` | Purge database (Only works if `PLATFORM=dev`) |
