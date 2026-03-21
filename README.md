# 🎬 Streaming Backend

A high-performance video streaming backend built in **Go**, following clean architecture principles.

## 🚀 Quick Start

### Prerequisites
- [Go 1.22+](https://go.dev/dl/) installed
- (Later steps) Docker, PostgreSQL, Redis

### Run the server

```bash
# Option 1: Using Go directly
go run ./cmd/api/main.go

# Option 2: Using Make
make run

# Option 3: Build and run the binary
make build
./bin/streaming-backend.exe
```

### Test it

Open your browser or use curl:

```bash
# Health check
curl http://localhost:8080/health

# Welcome endpoint
curl http://localhost:8080/
```

**Expected response from `/health`:**
```json
{
  "message": "Streaming backend is running 🎬",
  "status": "ok",
  "timestamp": "2026-03-12T08:00:00Z",
  "version": "0.1.0"
}
```

### Stop the server
Press `Ctrl+C` — the server will shut down gracefully, finishing any in-progress requests.

---

## 📁 Project Structure (so far)

```
streaming-backend/
├── cmd/
│   └── api/
│       └── main.go         ← Server entry point (you are here!)
├── docs/
│   └── ARCHITECTURE.md     ← Full system architecture
├── .env.example             ← Environment variable template
├── go.mod                   ← Go module definition
├── Makefile                 ← Build/run commands
└── README.md                ← This file
```

## 🗺️ Learning Roadmap

| Step | Topic | Status |
|------|-------|--------|
| 1 | Project Setup & Hello Server | ✅ Done |
| 2 | Router, Middleware & Config | ✅ Done |
| 3 | Database Connection | ✅ Done |
| 4 | User Registration & Login | ✅ Done |
| 5 | Auth Middleware & Profiles | ✅ Done |
| 6 | Video Upload & Metadata | ✅ Done |
| 7 | Video Transcoding | ✅ Done (Simulated Worker) |
| 8 | HLS Streaming | ⏳ Next |
| 9 | Search | 🔲 |
| 10 | Recommendations | 🔲 |
| 11 | Billing | 🔲 |
