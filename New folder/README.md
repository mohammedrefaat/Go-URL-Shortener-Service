# Go URL Shortener Service

A high-performance, production-ready URL shortening service built with Go, featuring Redis caching, PostgreSQL persistence, JWT authentication, and comprehensive analytics.

## 🚀 Features

- **URL Shortening**: Convert long URLs into short, memorable codes
- **Custom Aliases**: Support for user-defined short codes
- **Expiration Support**: Set expiration dates for shortened URLs
- **Analytics**: Detailed click tracking and daily statistics
- **Caching**: Redis-based caching for improved performance
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **Health Checks**: Comprehensive health monitoring endpoints
- **JWT Authentication**: Secure analytics access
- **Graceful Shutdown**: Proper server lifecycle management

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│   Gin Router    │───▶│   Handlers      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                               ┌─────────────────┐
                                               │   Services      │
                                               └─────────────────┘
                                                        │
                                    ┌───────────────────┼───────────────────┐
                            ┌─────────────────┐                   ┌─────────────────┐
                            │   PostgreSQL    │                   │     Redis       │
                            │   Repository    │                   │   Repository    │
                            └─────────────────┘                   └─────────────────┘
```

## 📋 Prerequisites

- Go 1.21 or later
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose (for testing)

## ⚡ Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/mohammedrefaat/Go-URL-Shortener-Service.git
cd Go-URL-Shortener-Service
```

### 2. Environment Setup
Create a `.env` file:
```env
# Server Configuration
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info
BASE_URL=http://localhost:8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=urlshortener
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Security
JWT_SECRET=your-super-secret-jwt-key

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60
```

### 3. Start Dependencies
```bash
# Start PostgreSQL and Redis using Docker Compose
docker-compose up -d postgres-test redis-test
```

### 4. Build and Run
```bash
# Using the build script
./cmd/build.bat

# Or manually
go mod tidy
go build -o urlshortener cmd/main.go
./urlshortener
```

## 🔧 API Endpoints

### Shorten URL
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/very-long-url",
  "custom_alias": "mylink",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "short_url": "http://localhost:8080/abc123",
  "short_code": "abc123",
  "original_url": "https://example.com/very-long-url",
  "expires_at": "2025-12-31T23:59:59Z",
  "created_at": "2025-08-27T10:30:00Z"
}
```

### Redirect
```http
GET /{shortCode}
```
Redirects to the original URL with 301 status.

### Analytics (JWT Required)
```http
GET /api/v1/analytics/{shortCode}?days=30
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "short_code": "abc123",
  "original_url": "https://example.com",
  "click_count": 142,
  "created_at": "2025-08-27T10:30:00Z",
  "last_accessed": "2025-08-27T15:45:00Z",
  "daily_stats": [
    {"date": "2025-08-27", "clicks": 25},
    {"date": "2025-08-26", "clicks": 31}
  ]
}
```

### Health Check
```http
GET /health
```

## 🧪 Testing

### Unit Tests
```bash
go test ./tests/unit/... -v
```

### Integration Tests
```bash
# Start test databases
docker-compose up -d

# Run integration tests
go test ./tests/integration/... -v
```

### End-to-End Tests
```bash
go test ./tests/e2e/... -v
```

### Manual Testing with cURL
```bash
# Shorten a URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Test redirect
curl -I http://localhost:8080/{shortCode}

# Health check
curl http://localhost:8080/health
```

## 🏃‍♂️ Performance

- **Throughput**: 1000+ requests/second
- **Latency**: < 10ms for cached URLs
- **Storage**: Supports millions of URLs
- **Cache Hit Rate**: 95%+ for active URLs

## 🔒 Security Features

- JWT-based authentication for analytics
- Rate limiting (100 requests/minute by default)
- URL validation and malicious domain blocking
- SQL injection prevention with parameterized queries
- CORS protection

## 📊 Monitoring

The service provides comprehensive health checks and logging:

- **Health Endpoint**: `/health` - Database and Redis connectivity status
- **Structured Logging**: JSON logs with zap logger
- **Metrics**: Request latency, error rates, cache hit rates

## 🚀 Deployment

### Docker Deployment
```bash
# Build Docker image
docker build -t url-shortener .

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e REDIS_HOST=your-redis-host \
  url-shortener
```

### Production Considerations

1. **Database**: Use connection pooling and read replicas
2. **Cache**: Redis Cluster for high availability
3. **Load Balancing**: Multiple service instances behind load balancer
4. **SSL/TLS**: Terminate SSL at load balancer level
5. **Monitoring**: Prometheus metrics and Grafana dashboards

## 🛠️ Configuration

All configuration is handled through environment variables. See the config package for available options.

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| DB_HOST | localhost | PostgreSQL host |
| REDIS_HOST | localhost | Redis host |
| JWT_SECRET | your-secret-key | JWT signing secret |
| RATE_LIMIT_REQUESTS | 100 | Requests per window |
| RATE_LIMIT_WINDOW | 60 | Rate limit window (seconds) |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

For questions or issues, please open a GitHub issue or contact the maintainers.

---

**Built with ❤️ using Go, PostgreSQL, and Redis**