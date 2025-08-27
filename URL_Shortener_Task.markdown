# URL Shortener Service

## Objective
Design and implement a **scalable URL Shortener Service** in Go, similar to Bitly or TinyURL. This task evaluates your ability to architect a production-ready system,
solution should demonstrate expertise in system design, clean code, scalability, and testability.

## Requirements

### Functional Requirements
1. **URL Shortening**: Convert long URLs to unique, fixed-length short codes (e.g., `https://example.com` → `https://short.url/abc123`, 6-8 characters).
2. **Redirection**: Redirect short URLs to original URLs (HTTP 301/302).
3. **Analytics**: Track clicks (count, timestamps) and provide an endpoint to retrieve stats per short URL.
4. **Optional Features** (bonus, not required):
   - Custom aliases (e.g., `https://short.url/mycustom`).
   - URL expiry (e.g., auto-delete after 30 days).
   - JWT-based authentication for private URLs or analytics.

### Non-Functional Requirements
1. **Performance**:
   - Redirect latency: <100ms.
   - Throughput: Handle ~1000 requests/second.
   - Use caching (Redis) for frequent redirects.
2. **Scalability**: Support horizontal scaling and millions of URLs.
3. **Security**:
   - Validate URLs to prevent malicious redirects.
   - Rate-limit API requests.
   - Secure analytics endpoint (e.g., JWT).
4. **Observability**:
   - Structured logging (e.g., Zap/Logrus).
5. **Testing**:
   - 80%+ code coverage.
   - Unit tests for business logic, integration tests for DB/cache, end-to-end tests for API flows.
   - Test edge cases: collisions, invalid URLs, high load.
6. **Deployment**:
   - Containerize with Docker.
   - Provide `docker-compose` for local development.

### Technical Stack
- **Language**: Go (latest stable version).
- **Framework**: Standard library or lightweight (e.g., Gin).
- **Database**: PostgreSQL (or Mysql);
- **Cache**: Redis.
- **Testing**: Go’s `testing` package, `testify` for mocks.
- **Deployment**: Docker.

## Deliverables
1. **Git Repository**:
   - Organized code with clear commit history.
   - Suggested structure:
 
2. **Documentation** (in `README.md` or `docs/architecture.md`):
   - **Architecture Overview**: Explain design choices (e.g., monolith vs. microservices, DB choice).
   - **Setup Instructions**: How to run locally (with `docker-compose`).
   - **API Docs**: List endpoints (e.g., `POST /shorten`, `GET /:shortCode`).
4. **Bonus** (optional):
   - CI/CD pipeline (e.g., GitHub Actions).
   - Load test results (e.g., using Vegeta).
   - Kubernetes manifests.

## Evaluation Criteria
- **Architecture Quality**: Clear separation of concerns, extensibility, scalability.
- **Code Quality**: Clean, idiomatic Go code; proper error handling; minimal comments.
- **Design Decisions**: Well-justified choices (e.g., DB, framework, patterns).
- **Testing**: Comprehensive coverage, edge case handling.
- **Documentation**: Clear explanation of design, setup, and trade-offs.

## Submission Guidelines
- Provide a public or shared Git repository (e.g., GitHub).
- Include a `README.md` with setup instructions and architecture overview.
- Submit within [insert timeline, e.g., 2 days].

## Notes
- Focus on **design first**: Start with architecture docs/diagrams before coding.
- Prioritize simplicity but demonstrate scalability potential.
- Contact [recruiter email] for clarifications.

Good luck! We look forward to reviewing your solution.
