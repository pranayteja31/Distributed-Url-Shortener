# 🔗 URL Shortener

A production-oriented URL Shortener backend built with **Go, Gin, PostgreSQL and Redis**.

The project provides URL creation and management, fast Redis-backed redirection, URL expiration, click analytics, rate limiting, application metrics, monitoring and load testing. The entire application stack can be run using Docker Compose.

---

## 🚀 Features

- Create short URLs
- Redirect short URLs to original URLs
- Get URL details
- List URLs
- Update URLs
- Delete URLs
- URL validation
- URL expiration
- Unique short-code generation
- Redis caching with TTL
- PostgreSQL persistence
- Click counting
- Asynchronous click analytics
- Rate limiting
- Prometheus metrics
- Grafana monitoring
- Dockerized application
- PostgreSQL database migrations
- Automated tests
- k6 load testing

---

## 🛠️ Technology Stack

| Technology | Purpose |
|---|---|
| **Go** | Backend programming language |
| **Gin** | HTTP web framework |
| **PostgreSQL** | Persistent database |
| **Redis** | URL caching |
| **Prometheus** | Application metrics |
| **Grafana** | Metrics visualization |
| **Docker** | Containerization |
| **Docker Compose** | Multi-container environment |
| **k6** | Load testing |

---

# 🏗️ Architecture

The application follows a layered backend architecture.

```text
                         Client
                           |
                           v
                    +-------------+
                    | Gin Router  |
                    +-------------+
                           |
                           v
                    +-------------+
                    | Middleware  |
                    +-------------+
                           |
                           v
                    +-------------+
                    |  Handlers   |
                    +-------------+
                           |
                           v
                    +-------------+
                    |  Services   |
                    +-------------+
                       /       \
                      /         \
                     v           v
                +--------+   +------------+
                | Redis  |   | Repository |
                | Cache  |   +------------+
                +--------+          |
                                    v
                              +-----------+
                              |PostgreSQL |
                              +-----------+

                         Analytics
                            |
                            v
                    Background Worker
                            |
                            v
                       url_clicks
```

### Observability

```text
                 Go Application
                       |
                       v
                 /metrics
                       |
                       v
                  Prometheus
                       |
                       v
                    Grafana
```

---

# 📁 Project Structure

```text
URL_SHORTENER/
│
├── .dockerignore
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
│
├── loadtest.js
├── loadtest_hit.js
├── loadmiss-test.js
│
├── cmd/
│   └── url-shortener/
│       └── main.go
│
├── config/
│   ├── local.example.yaml
│   └── local.yaml
│
├── internal/
│   ├── analytics/
│   │   ├── worker.go
│   │   └── worker_test.go
│   │
│   ├── cache/
│   │   ├── cache.go
│   │   ├── cache_test.go
│   │   ├── keys.go
│   │   ├── redis.go
│   │   ├── redis_cache.go
│   │   └── ttl.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── db/
│   │   └── db_config.go
│   │
│   ├── dto/
│   │   └── response.go
│   │
│   ├── handlers/
│   │   ├── admin_handlers.go
│   │   ├── analytics_handler.go
│   │   ├── errors.go
│   │   ├── health_handler.go
│   │   ├── response.go
│   │   ├── urlHandlers.go
│   │   └── url_handler_test.go
│   │
│   ├── metrics/
│   │   ├── metrics.go
│   │   └── metrics_test.go
│   │
│   ├── middlerware/
│   │   ├── metrics.go
│   │   ├── rate_limiter.go
│   │   └── url_middleware_test.go
│   │
│   ├── models/
│   │   ├── click.go
│   │   └── url.go
│   │
│   ├── repository/
│   │   ├── click_repository.go
│   │   ├── url_repository.go
│   │   └── url_repository_test.go
│   │
│   ├── routes/
│   │   └── routes.go
│   │
│   ├── services/
│   │   ├── url_services.go
│   │   └── url_service_test.go
│   │
│   └── utils/
│       ├── shortcode.go
│       └── url_validator.go
│
├── migrations/
│   ├── 000001_create_urls.down.sql
│   ├── 000001_create_urls.up.sql
│   ├── 000002_create_url_clicks.down.sql
│   └── 000002_create_url_clicks.up.sql
│
└── prometheus/
    └── prometheus.yml
```

---

# 🔄 URL Redirection Flow

URL redirection is optimized using Redis.

```text
Client
  |
  | GET /:shortCode
  v
Gin Handler
  |
  v
Redis GET
  |
  +-------------------+
  |                   |
 HIT                 MISS
  |                   |
  v                   v
URL found         PostgreSQL
  |                   |
  |                   v
  |              URL found
  |                   |
  +---------+---------+
            |
            v
      Check expiration
            |
            v
      Increment clicks
            |
            v
     Analytics Worker
            |
            v
       Store in Redis
       (on cache miss)
            |
            v
         Redirect
```

### Cache Hit

When a short URL exists in Redis:

1. Redis returns the cached URL.
2. The application checks whether the URL has expired.
3. The click count is incremented.
4. Click analytics are sent to the background worker.
5. The original URL is returned for redirection.

### Cache Miss

When the URL is not present in Redis:

1. Redis lookup fails.
2. PostgreSQL is queried.
3. The URL expiration is checked.
4. The click count is incremented.
5. Analytics are recorded asynchronously.
6. The URL is stored in Redis using an appropriate TTL.
7. The original URL is returned.

Redis therefore acts as a **caching layer**, while PostgreSQL remains the persistent source of URL data.

---

# 🔌 API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/url/create` | Create a short URL |
| `GET` | `/api/v1/url/get/:id` | Get a URL by ID |
| `GET` | `/api/v1/url/list` | List URLs |
| `PUT` | `/api/v1/url/update/:id` | Update a URL |
| `DELETE` | `/api/v1/url/delete/:id` | Delete a URL |
| `GET` | `/:shortCode` | Redirect to original URL |
| `GET` | `/metrics` | Prometheus metrics |

The repository also contains health, admin and analytics handlers; their actual routes should be checked in `internal/routes/routes.go`.

---

# 📝 Create Short URL

### Request

```http
POST /api/v1/url/create
Content-Type: application/json
```

Example:

```json
{
  "url": "https://example.com",
  "expiration": 30
}
```

The application validates the URL, generates a unique short code and stores the URL in PostgreSQL.

The newly created URL can also be cached in Redis with a TTL corresponding to its expiration.

---

# 🔀 Redirect

```http
GET /:shortCode
```

Example:

```text
GET /aB72xQ
```

The application attempts to retrieve the URL from Redis first.

If the URL is not cached, PostgreSQL is queried and the result is placed into Redis for subsequent requests.

---

# 🗄️ Database

PostgreSQL is used for persistent storage.

The migrations create the following tables:

```text
schema_migrations
urls
url_clicks
```

The main relationship is:

```text
+----------------+
|      urls      |
+----------------+
| id             |
| short_code     |
| original_url   |
| created_at     |
| expires_at     |
| click_count    |
+----------------+
        |
        | 1
        |
        | N
        v
+----------------+
|   url_clicks   |
+----------------+
| id             |
| url_id         |
| clicked_at     |
| ip_address     |
| user_agent     |
| referrer       |
+----------------+
```

`urls` stores the shortened URL information.

`url_clicks` stores click analytics associated with a URL.

Database schema and constraints are defined through the SQL migration files in:

```text
migrations/
```

---

# ⚙️ Configuration

The application supports environment-based configuration.

Important variables include:

```text
ENV
ADDRESS

DB_HOST
DB_PORT
DB_USER
DB_PASSWORD
DB_NAME
DB_SSLMODE

REDIS_HOST
REDIS_PORT
REDIS_PASSWORD
REDIS_DB
REDIS_POOLSIZE

CACHE_ENABLED
```

For Docker Compose, the application communicates with PostgreSQL and Redis using their Compose service names:

```text
DB_HOST=postgres
REDIS_HOST=redis
```

The application listens on:

```text
0.0.0.0:8082
```

---

# 🐳 Docker

The complete application stack can be run using Docker Compose.

The Compose environment contains:

```text
+----------------------+
|   url-shortener-app  |
|        :8082         |
+----------+-----------+
           |
     +-----+-----+
     |           |
     v           v
+---------+   +-------+
|Postgres |   | Redis |
|  :5432  |   | :6379 |
+---------+   +-------+
```

PostgreSQL uses a persistent Docker volume.

Redis also uses a persistent Docker volume.

PostgreSQL has a healthcheck configured so the application waits for the database to become healthy.

---

# 🚀 Running the Project

## Prerequisites

Install:

- Go
- Docker Desktop
- Git

For load testing:

- k6

---

## 1. Clone the repository

```bash
git clone <your-repository-url>
cd URL_SHORTENER
```

---

## 2. Configure environment variables

Create a `.env` file:

```env
DB_PASSWORD=your_password
```

Do not commit `.env` to GitHub.

---

## 3. Start the application

```bash
docker compose up --build
```

The application will be available at:

```text
http://localhost:8082
```

---

## 4. Check running containers

```bash
docker compose ps
```

Expected services:

```text
url-shortener-app
url-shortener-postgres
url-shortener-redis
```

---

## 5. Stop the application

```bash
docker compose down
```

To remove containers and associated volumes:

```bash
docker compose down -v
```

> Removing volumes deletes persisted PostgreSQL/Redis data.

---

# 🗃️ Database Migrations

The project uses SQL migrations.

Migration files:

```text
migrations/
├── 000001_create_urls.up.sql
├── 000001_create_urls.down.sql
├── 000002_create_url_clicks.up.sql
└── 000002_create_url_clicks.down.sql
```

After migration, PostgreSQL contains:

```text
schema_migrations
urls
url_clicks
```

---

# 🧪 Testing

The project contains tests across multiple layers.

Tested areas include:

- Analytics worker
- Redis cache
- Handlers
- Metrics
- Middleware
- Repository
- Services

Run all Go tests:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

---

# 📊 Monitoring

The application exposes Prometheus metrics through:

```text
GET /metrics
```

Example:

```text
http://localhost:8082/metrics
```

Prometheus collects application metrics and Grafana is used to visualize them.

---

# 📈 Prometheus Metrics

Custom application metrics include:

```text
url_shortener_http_requests_total
url_shortener_http_request_latency_seconds

url_shortener_cache_hits_total
url_shortener_cache_misses_total

url_shortener_postgres_lookup_latency_seconds
url_shortener_redis_get_latency_seconds
```

These metrics provide visibility into:

- HTTP traffic
- HTTP latency
- Redis cache performance
- PostgreSQL lookup latency
- Application behavior

Standard Go runtime/process metrics are also exposed through Prometheus.

---

# 📊 Grafana

Grafana is connected to Prometheus and provides dashboards for application monitoring.

Useful metrics include:

```text
HTTP Requests / Second
HTTP Request Latency
P95 Latency
Redis Cache Hits
Redis Cache Misses
Redis Latency
PostgreSQL Lookup Latency
```

The monitoring pipeline is:

```text
Go Application
      |
      | /metrics
      v
 Prometheus
      |
      v
  Grafana
```

---

# ⚡ Load Testing

The project uses **k6** to test application performance.

Load-testing scripts include:

```text
loadtest.js
loadtest_hit.js
loadmiss-test.js
```

The tests evaluate URL redirection under concurrent load and can be used to observe the effect of Redis caching.

---

# 📈 Performance Results

One of the final local Docker benchmark runs used:

```text
Virtual Users: 5
Duration:      30 seconds
```

### Results

| Metric | Result |
|---|---:|
| Total Requests | 16,243 |
| Throughput | 541.34 req/s |
| Average Latency | 9.11 ms |
| Median Latency | 8.12 ms |
| P90 Latency | 13.90 ms |
| P95 Latency | 16.25 ms |
| Failed Requests | 0% |
| Successful Checks | 100% |

Redis metrics after the test:

```text
Cache Hits:   16,243
Cache Misses: 0
```

### Important

These numbers were measured on a **local Docker environment using 5 virtual users for 30 seconds**.

They represent the observed performance under this specific workload and environment. They should not be interpreted as the maximum production capacity of the application.

---

# 🔬 Cache Performance Experiment

The application was also tested with Redis caching enabled and disabled.

### Redis Enabled

```text
Requests:       14,766
Throughput:     492.14 req/s
Average:        10.02 ms
Median:          8.85 ms
P90:            15.13 ms
P95:            17.88 ms
Failures:        0%
```

### Redis Disabled

```text
Requests:       17,063
Throughput:     568.68 req/s
Average:         8.67 ms
Median:          7.74 ms
P90:            13.02 ms
P95:            15.27 ms
Failures:        0%
```

This comparison should not be interpreted as Redis being slower than PostgreSQL.

The redirect operation still performs PostgreSQL click-count updates even when the URL itself is served from Redis. Therefore, the benchmark measures the performance of the **complete application request path**, not an isolated Redis-vs-PostgreSQL lookup.

---

# 📦 Docker Services

The project uses three main containers:

### Application

```text
url-shortener-app
Port: 8082
```

### PostgreSQL

```text
url-shortener-postgres
Port: 5432
```

### Redis

```text
url-shortener-redis
Port: 6379
```

Docker Compose provides networking between the services.

Inside the Docker network:

```text
postgres
redis
```

are used as service hostnames.

---

# 🔐 Security & Reliability

The project includes several basic security and reliability mechanisms:

- URL validation
- Rate limiting
- Environment-based database credentials
- Parameterized database operations
- Expiration checks
- PostgreSQL persistence
- Redis TTL
- Docker isolation
- Structured error handling

Secrets such as database passwords should never be committed to the repository.

---

# 🧠 Design Decisions

### Why Go?

Go provides:

- Simple concurrency
- Low runtime overhead
- Strong standard library
- Fast HTTP services
- Easy containerization

### Why PostgreSQL?

PostgreSQL provides reliable persistent storage and relational data management for URLs and click analytics.

### Why Redis?

Redis provides fast in-memory access for frequently requested short URLs, reducing repeated database lookups.

### Why Docker?

Docker provides a consistent environment for running:

```text
Application
PostgreSQL
Redis
```

without requiring each dependency to be manually configured.

### Why Prometheus and Grafana?

Prometheus provides application metrics while Grafana provides visualization and monitoring dashboards.

### Why asynchronous analytics?

Click analytics do not need to block the redirect response. A background worker allows analytics processing to happen asynchronously.

---

# ⚠️ Current Limitations

The current project is focused on backend functionality.

Some potential future improvements include:

- Authentication and authorization
- API documentation using OpenAPI/Swagger
- Cloud deployment
- CI/CD pipeline
- Distributed deployment
- Advanced analytics dashboards
- Horizontal scaling
- More advanced cache invalidation strategies

These are future improvements and are not part of the current implementation.

---

# 📌 Project Highlights

This project demonstrates practical backend engineering concepts:

- Layered architecture
- REST API development
- PostgreSQL persistence
- Redis caching
- Cache hit/miss monitoring
- URL expiration
- Rate limiting
- Asynchronous processing
- Database migrations
- Docker containerization
- Prometheus observability
- Grafana monitoring
- Automated testing
- k6 performance testing

---

# 📚 What This Project Demonstrates

The project goes beyond a basic CRUD URL shortener by implementing an application architecture containing:

```text
API Layer
     ↓
Business Logic
     ↓
Persistence Layer
     ↓
Caching Layer
     ↓
Analytics
     ↓
Observability
     ↓
Performance Testing
```

This makes the project suitable as a backend/system-oriented portfolio project.

---

# 👨‍💻 Project Status

**Status: Completed**

The current implementation includes the core URL-shortening functionality, persistence, caching, analytics, observability, containerization, testing and performance testing.

---

# 📄 License

Add an appropriate open-source license before distributing the project publicly.