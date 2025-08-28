# Go URL Shortener Service

A scalable and efficient URL shortening service built with **Golang**, designed for performance, simplicity, and easy extensibility.  
It provides a REST API for generating, retrieving, and managing shortened URLs, with support for caching and persistence.

---

## ✨ Features

- Shorten long URLs into unique short links
- Redirect users from short links to original URLs
- REST API built with [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL for persistence
- Redis for caching frequently accessed URLs
- Configurable via YAML
- Unit and integration tests included
- Structured logging with [Zap](https://github.com/uber-go/zap)

---

## 🛠 Tech Stack

- **Language:** Go (>= 1.22)
- **Framework:** Gin Web Framework
- **Database:** PostgreSQL
- **Cache:** Redis
- **Logging:** Zap

---

## 🚀 Getting Started

### Prerequisites

Make sure you have the following installed:
- Go (>= 1.22)
- PostgreSQL
- Redis
- Docker (optional, for containerized setup)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/mohammedrefaat/Go-URL-Shortener-Service.git
   cd Go-URL-Shortener-Service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Create a `config.yaml` file (see [Configuration](#-configuration)).

4. Run the application:
   ```bash
   go run ./cmd
   ```

The server will start on `http://localhost:8080` (default).

---

## ⚙️ Configuration

The application is configured using `config.yaml`. Example:

```yaml
server:
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: url_shortener

cache:
  host: localhost
  port: 6379
```

---

## 📡 API Endpoints

### Health Check
```
GET /health
```
Response: `{"status":"ok"}`

### Shorten URL
```
POST /shorten
```
Request Body:
```json
{
  "url": "https://example.com/very/long/link"
}
```

Response:
```json
{
  "short_url": "http://localhost:8080/abc123"
}
```

### Redirect
```
GET /:short_id
```
Redirects to the original URL.

---

## 🧪 Testing

Run all tests:
```bash
go test ./...
```

---

## 🐳 Docker Setup

To run with Docker Compose:

```bash
docker-compose up --build
```

This will start:
- The API service
- PostgreSQL
- Redis

---

## 📂 Project Structure

```
Go-URL-Shortener-Service/
├── cmd/                # Application entrypoint
├── internal/
│   ├── domain/         # Domain models
│   ├── repository/     # Database and cache repositories
│   ├── service/        # Business logic
│   ├── handler/        # HTTP handlers (Gin)
├── tests/              # Unit and integration tests
├── config/             # Configuration files
└── docker-compose.yml  # Docker setup
```

---

## 📜 License

This project is licensed under the MIT License.
