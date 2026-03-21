# 🎬 Streaming Website Backend — Architecture & Documentation

## 1. Overview

A high-performance, scalable **video streaming backend** built in **Go (Golang)**. The system supports user authentication, video upload & transcoding, adaptive bitrate streaming (HLS/DASH), content management, search, recommendations, and real-time analytics.

---

## 2. High-Level Architecture

```mermaid
graph TB
    subgraph Clients
        WEB[Web App]
        MOB[Mobile App]
        TV[Smart TV / OTT]
    end

    subgraph API_GATEWAY["API Gateway (Go)"]
        GW[Rate Limiting · Auth · Routing]
    end

    subgraph MICROSERVICES["Backend Microservices (Go)"]
        AUTH[Auth Service]
        USER[User Service]
        VIDEO[Video Service]
        STREAM[Streaming Service]
        SEARCH[Search Service]
        REC[Recommendation Service]
        NOTIFY[Notification Service]
        BILLING[Billing Service]
    end

    subgraph DATA_LAYER["Data Layer"]
        PG[(PostgreSQL)]
        REDIS[(Redis Cache)]
        ES[(Elasticsearch)]
        MONGO[(MongoDB — Analytics)]
    end

    subgraph INFRA["Infrastructure"]
        S3[Object Storage — S3/MinIO]
        CDN[CDN — CloudFront/Nginx]
        MQ[Message Queue — NATS/Kafka]
        FFMPEG[Transcoding Workers]
    end

    WEB & MOB & TV --> GW
    GW --> AUTH & USER & VIDEO & STREAM & SEARCH & REC & NOTIFY & BILLING
    AUTH & USER & BILLING --> PG
    VIDEO --> PG & S3 & MQ
    STREAM --> S3 & CDN & REDIS
    SEARCH --> ES
    REC --> REDIS & MONGO
    NOTIFY --> MQ
    MQ --> FFMPEG
    FFMPEG --> S3
```

---

## 3. Microservices Breakdown

| Service | Responsibility | Port |
|---|---|---|
| **API Gateway** | Rate limiting, JWT validation, request routing, CORS | `:8080` |
| **Auth Service** | Registration, login, OAuth2, JWT token issuance/refresh | `:8081` |
| **User Service** | Profiles, preferences, watch history, subscriptions | `:8082` |
| **Video Service** | Upload, metadata CRUD, transcoding dispatch | `:8083` |
| **Streaming Service** | HLS/DASH manifest generation, chunk serving, DRM | `:8084` |
| **Search Service** | Full-text search, filters, trending content | `:8085` |
| **Recommendation Service** | Personalized feeds, trending, "continue watching" | `:8086` |
| **Notification Service** | Email, push notifications, in-app alerts | `:8087` |
| **Billing Service** | Subscriptions, payments, invoices (Stripe/Razorpay) | `:8088` |

---

## 4. Video Processing Pipeline

```mermaid
sequenceDiagram
    participant Client
    participant VideoSvc as Video Service
    participant ObjStore as Object Storage
    participant Queue as Message Queue
    participant Worker as Transcoding Worker
    participant CDN

    Client->>VideoSvc: Upload video (multipart)
    VideoSvc->>ObjStore: Store raw file
    VideoSvc->>Queue: Publish "transcode" event
    Queue->>Worker: Consume event
    Worker->>ObjStore: Download raw file
    Worker->>Worker: FFmpeg transcode (240p, 360p, 480p, 720p, 1080p)
    Worker->>ObjStore: Upload HLS segments + playlist
    Worker->>Queue: Publish "transcode.complete"
    Queue->>VideoSvc: Update video status → "ready"
    Client->>CDN: Request stream
    CDN->>ObjStore: Fetch HLS segments (cached)
    CDN->>Client: Stream adaptive bitrate
```

### Transcoding Output Structure
```
/videos/{video_id}/
├── master.m3u8            # Master HLS playlist
├── 1080p/
│   ├── playlist.m3u8
│   └── segment_000.ts ... segment_N.ts
├── 720p/
│   ├── playlist.m3u8
│   └── segment_000.ts ... segment_N.ts
├── 480p/
│   ├── playlist.m3u8
│   └── segment_000.ts ... segment_N.ts
├── 360p/
│   └── ...
├── 240p/
│   └── ...
└── thumbnail.jpg
```

---

## 5. Database Schema (PostgreSQL)

```mermaid
erDiagram
    USERS ||--o{ SUBSCRIPTIONS : has
    USERS ||--o{ WATCH_HISTORY : has
    USERS ||--o{ PLAYLISTS : creates
    VIDEOS ||--o{ WATCH_HISTORY : appears_in
    VIDEOS }o--|| CHANNELS : belongs_to
    VIDEOS ||--o{ VIDEO_TAGS : has
    TAGS ||--o{ VIDEO_TAGS : has
    CHANNELS }o--|| USERS : owned_by
    PLAYLISTS ||--o{ PLAYLIST_ITEMS : contains
    PLAYLIST_ITEMS }o--|| VIDEOS : references

    USERS {
        uuid id PK
        string email UK
        string username UK
        string password_hash
        string avatar_url
        string role
        timestamp created_at
        timestamp updated_at
    }

    VIDEOS {
        uuid id PK
        uuid channel_id FK
        string title
        text description
        string status
        int duration_seconds
        string thumbnail_url
        string storage_path
        bigint view_count
        string visibility
        timestamp published_at
        timestamp created_at
    }

    CHANNELS {
        uuid id PK
        uuid owner_id FK
        string name UK
        text description
        string banner_url
        bigint subscriber_count
        timestamp created_at
    }

    SUBSCRIPTIONS {
        uuid id PK
        uuid user_id FK
        string plan
        string status
        timestamp start_date
        timestamp end_date
    }

    WATCH_HISTORY {
        uuid id PK
        uuid user_id FK
        uuid video_id FK
        int progress_seconds
        timestamp watched_at
    }

    TAGS {
        int id PK
        string name UK
    }

    VIDEO_TAGS {
        uuid video_id FK
        int tag_id FK
    }

    PLAYLISTS {
        uuid id PK
        uuid user_id FK
        string title
        string visibility
        timestamp created_at
    }

    PLAYLIST_ITEMS {
        uuid playlist_id FK
        uuid video_id FK
        int position
    }
```

---

## 6. API Design (RESTful)

### Auth
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/auth/register` | Register a new user |
| `POST` | `/api/v1/auth/login` | Login, returns JWT |
| `POST` | `/api/v1/auth/refresh` | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Invalidate token |
| `POST` | `/api/v1/auth/oauth/{provider}` | OAuth2 login (Google, GitHub) |

### Users
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/users/me` | Get current user profile |
| `PUT` | `/api/v1/users/me` | Update profile |
| `GET` | `/api/v1/users/{id}` | Get public profile |
| `GET` | `/api/v1/users/me/history` | Watch history |
| `DELETE` | `/api/v1/users/me/history` | Clear watch history |

### Videos
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/videos` | Upload video (multipart) |
| `GET` | `/api/v1/videos/{id}` | Get video metadata |
| `PUT` | `/api/v1/videos/{id}` | Update metadata |
| `DELETE` | `/api/v1/videos/{id}` | Delete video |
| `GET` | `/api/v1/videos` | List/filter videos |
| `GET` | `/api/v1/videos/trending` | Trending videos |
| `POST` | `/api/v1/videos/{id}/view` | Record a view |

### Streaming
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/stream/{video_id}/manifest` | Get HLS master playlist |
| `GET` | `/api/v1/stream/{video_id}/{quality}/playlist` | Quality-specific playlist |
| `GET` | `/api/v1/stream/{video_id}/{quality}/{segment}` | Get video segment |

### Channels
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/channels` | Create channel |
| `GET` | `/api/v1/channels/{id}` | Get channel info |
| `PUT` | `/api/v1/channels/{id}` | Update channel |
| `POST` | `/api/v1/channels/{id}/subscribe` | Subscribe |
| `DELETE` | `/api/v1/channels/{id}/subscribe` | Unsubscribe |

### Search & Recommendations
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/search?q={query}` | Full-text search |
| `GET` | `/api/v1/recommendations` | Personalized feed |
| `GET` | `/api/v1/recommendations/continue` | Continue watching |

### Billing
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/billing/subscribe` | Start subscription |
| `POST` | `/api/v1/billing/cancel` | Cancel subscription |
| `GET` | `/api/v1/billing/invoices` | Get invoices |
| `POST` | `/api/v1/billing/webhook` | Payment webhook |

---

## 7. Technology Stack

| Layer | Technology | Why |
|---|---|---|
| **Language** | Go 1.22+ | High concurrency, fast compilation, excellent stdlib |
| **HTTP Router** | [Chi](https://github.com/go-chi/chi) or Gin | Lightweight, idiomatic middleware |
| **Database** | PostgreSQL 16 | ACID compliance, JSONB, full-text search |
| **Cache** | Redis 7 | Session store, rate limiting, hot data |
| **Search** | Elasticsearch 8 | Full-text search, autocomplete, analytics |
| **Object Storage** | MinIO / AWS S3 | Scalable blob storage for video files |
| **Message Queue** | NATS JetStream / Kafka | Async event processing, transcoding pipeline |
| **Transcoding** | FFmpeg | Industry-standard video processing |
| **CDN** | CloudFront / Nginx | Edge caching, low-latency delivery |
| **Auth** | JWT + bcrypt | Stateless authentication |
| **ORM/Query** | [sqlc](https://sqlc.dev/) | Type-safe SQL, compile-time checked queries |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | Version-controlled schema migrations |
| **Config** | [Viper](https://github.com/spf13/viper) | Env + file-based configuration |
| **Logging** | [Zap](https://github.com/uber-go/zap) | Structured, high-perf logging |
| **Observability** | OpenTelemetry + Prometheus + Grafana | Traces, metrics, dashboards |
| **Containerization** | Docker + Docker Compose | Local dev & deployment |
| **Orchestration** | Kubernetes (optional) | Production scaling |

---

## 8. Project Structure

```
streaming-backend/
├── cmd/                           # Application entrypoints
│   ├── api-gateway/
│   │   └── main.go
│   ├── auth-service/
│   │   └── main.go
│   ├── user-service/
│   │   └── main.go
│   ├── video-service/
│   │   └── main.go
│   ├── streaming-service/
│   │   └── main.go
│   ├── search-service/
│   │   └── main.go
│   ├── recommendation-service/
│   │   └── main.go
│   ├── notification-service/
│   │   └── main.go
│   ├── billing-service/
│   │   └── main.go
│   └── transcoding-worker/
│       └── main.go
│
├── internal/                      # Private application code
│   ├── auth/
│   │   ├── handler.go             # HTTP handlers
│   │   ├── service.go             # Business logic
│   │   ├── repository.go          # Data access
│   │   └── middleware.go          # Auth middleware
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   ├── video/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── model.go
│   │   └── transcoder.go
│   ├── streaming/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── hls.go                 # HLS manifest/segment logic
│   ├── search/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── indexer.go
│   ├── recommendation/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── engine.go
│   ├── notification/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── providers/             # Email, push, SMS
│   │       ├── email.go
│   │       └── push.go
│   ├── billing/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── stripe.go
│   └── common/                    # Shared utilities
│       ├── config/
│       │   └── config.go
│       ├── database/
│       │   ├── postgres.go
│       │   └── redis.go
│       ├── logger/
│       │   └── logger.go
│       ├── middleware/
│       │   ├── cors.go
│       │   ├── ratelimit.go
│       │   └── logging.go
│       ├── storage/
│       │   └── s3.go
│       ├── queue/
│       │   └── nats.go
│       └── errors/
│           └── errors.go
│
├── pkg/                           # Public reusable packages
│   ├── jwt/
│   │   └── jwt.go
│   ├── validator/
│   │   └── validator.go
│   └── pagination/
│       └── pagination.go
│
├── migrations/                    # SQL migration files
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_channels.up.sql
│   └── ...
│
├── deployments/                   # Deployment configs
│   ├── docker/
│   │   ├── Dockerfile.api-gateway
│   │   ├── Dockerfile.auth
│   │   ├── Dockerfile.video
│   │   └── ...
│   ├── docker-compose.yml
│   └── k8s/
│       ├── namespace.yaml
│       ├── api-gateway.yaml
│       └── ...
│
├── scripts/                       # Dev & CI scripts
│   ├── setup.sh
│   ├── migrate.sh
│   └── seed.sh
│
├── docs/                          # Additional documentation
│   └── api-reference.md
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 9. Key Design Patterns

### 9.1 Clean Architecture (per service)

```mermaid
graph LR
    H[Handler — HTTP Layer] --> S[Service — Business Logic]
    S --> R[Repository — Data Access]
    R --> DB[(Database)]

    style H fill:#4ecdc4,color:#000
    style S fill:#45b7d1,color:#000
    style R fill:#f7dc6f,color:#000
    style DB fill:#e74c3c,color:#fff
```

Each service follows:
- **Handler** — Parses HTTP requests, calls service, writes response
- **Service** — Pure business logic, no HTTP or DB awareness
- **Repository** — Data access via interfaces (easily swappable/mockable)

### 9.2 Dependency Injection

```go
// Repository interface (internal/video/repository.go)
type VideoRepository interface {
    Create(ctx context.Context, video *Video) error
    GetByID(ctx context.Context, id uuid.UUID) (*Video, error)
    List(ctx context.Context, filter VideoFilter) ([]Video, int, error)
    Update(ctx context.Context, video *Video) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// Service depends on interface, not concrete impl
type VideoService struct {
    repo    VideoRepository
    storage storage.ObjectStore
    queue   queue.Publisher
}
```

### 9.3 Event-Driven Communication

Services communicate asynchronously via NATS/Kafka for operations like:
- **Video uploaded** → triggers transcoding workers
- **Transcoding complete** → updates video status, notifies uploader
- **New subscriber** → sends welcome email
- **Video published** → indexes in Elasticsearch

---

## 10. Security

| Concern | Solution |
|---|---|
| **Authentication** | JWT access tokens (15 min) + refresh tokens (7 days) |
| **Password Storage** | bcrypt with cost factor 12 |
| **Authorization** | Role-based (admin, creator, viewer) + resource ownership checks |
| **Rate Limiting** | Token bucket per IP/user via Redis |
| **Input Validation** | Struct tag validation with `go-playground/validator` |
| **CORS** | Configurable allowed origins |
| **HTTPS** | TLS termination at load balancer / reverse proxy |
| **Upload Security** | File type validation, size limits, virus scanning |
| **SQL Injection** | Parameterized queries via sqlc |
| **Secrets** | Environment variables, never in code |

---

## 11. Scalability Strategy

```mermaid
graph TB
    LB[Load Balancer] --> GW1[API Gateway 1]
    LB --> GW2[API Gateway 2]
    LB --> GW3[API Gateway N]

    GW1 & GW2 & GW3 --> SVC[Service Mesh]

    SVC --> DB_PRIMARY[(PostgreSQL Primary)]
    DB_PRIMARY --> DB_READ1[(Read Replica 1)]
    DB_PRIMARY --> DB_READ2[(Read Replica 2)]

    SVC --> REDIS_CLUSTER[Redis Cluster]
    SVC --> ES_CLUSTER[Elasticsearch Cluster]
    SVC --> NATS_CLUSTER[NATS Cluster]

    NATS_CLUSTER --> W1[Worker 1]
    NATS_CLUSTER --> W2[Worker 2]
    NATS_CLUSTER --> W3[Worker N]
```

| Strategy | Implementation |
|---|---|
| **Horizontal scaling** | Stateless services behind a load balancer |
| **Database read replicas** | Separate read/write connections |
| **Caching** | Redis for sessions, hot metadata, rate limits |
| **CDN** | Edge-cached video segments, thumbnails |
| **Async processing** | Transcoding, notifications via message queue |
| **Connection pooling** | `pgxpool` for PostgreSQL |
| **Graceful shutdown** | `context.Context` propagation, `os.Signal` handling |

---

## 12. Deployment

### Docker Compose (Development)
```yaml
# docker-compose.yml (simplified)
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: streaming
      POSTGRES_USER: streaming
      POSTGRES_PASSWORD: secret
    ports: ["5432:5432"]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    ports: ["9000:9000", "9001:9001"]

  nats:
    image: nats:latest
    command: ["--jetstream"]
    ports: ["4222:4222"]

  elasticsearch:
    image: elasticsearch:8.12.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports: ["9200:9200"]

  api-gateway:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile.api-gateway
    ports: ["8080:8080"]
    depends_on: [postgres, redis]

  # ... other services follow the same pattern
```

### Makefile
```makefile
.PHONY: build run test migrate lint docker-up docker-down

build:
	go build -o bin/ ./cmd/...

run:
	docker-compose up -d postgres redis minio nats elasticsearch
	go run ./cmd/api-gateway

test:
	go test -race -cover ./...

migrate-up:
	migrate -path migrations -database "postgres://streaming:secret@localhost:5432/streaming?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://streaming:secret@localhost:5432/streaming?sslmode=disable" down 1

lint:
	golangci-lint run ./...

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down -v
```

---

## 13. Getting Started (Quick Start)

```bash
# 1. Clone the repository
git clone https://github.com/your-org/streaming-backend.git
cd streaming-backend

# 2. Copy environment configuration
cp .env.example .env

# 3. Start infrastructure
docker-compose up -d postgres redis minio nats elasticsearch

# 4. Run database migrations
make migrate-up

# 5. Run the API gateway
go run ./cmd/api-gateway

# 6. (Optional) Run individual services
go run ./cmd/auth-service
go run ./cmd/video-service
```

---

## 14. Configuration (`.env.example`)

```env
# Server
APP_ENV=development
API_GATEWAY_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=streaming
DB_PASSWORD=secret
DB_NAME=streaming
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-256-bit-secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# MinIO / S3
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=videos
S3_USE_SSL=false

# NATS
NATS_URL=nats://localhost:4222

# Elasticsearch
ES_URL=http://localhost:9200

# Stripe (Billing)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

---

> [!TIP]
> **Recommended development order**: Auth → User → Video (upload + metadata) → Transcoding Worker → Streaming → Search → Recommendations → Billing → Notifications
