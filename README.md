# Sanguis

Sanguis is a high-performance, production-grade distributed rate limiter built with Go. It provides low-latency rate limiting services via gRPC and ensures global consistency using Redis.

## Key Features

- **Distributed Rate Limiting**: Global state management using Redis with multiple strategies (Sliding Window, Token Bucket).
- **Hybrid L1/L2 Caching**: Architect-level strategy using local in-memory L1 cache for sub-millisecond latency and periodic L2 (Redis) synchronization for global consistency.
- **Resilience**: Fail-open strategy for Redis operations and graceful shutdown support (SIGINT/SIGTERM).
- **Observability**: 
    - Structured logging with `uber-go/zap`.
    - Real-time metrics with Prometheus (Requests total, Redis latency).
    - gRPC Unary Interceptor for request tracking.
- **Infrastructure**: Multi-stage Docker builds and automated CI/CD with GitHub Actions.

## Tech Stack

- **Language**: Go 1.25+
- **Communication**: gRPC
- **Storage**: Redis
- **Observability**: Prometheus & Zap
- **Configuration**: Cleanenv (YAML & Env vars)
- **Deployment**: Docker & Docker Compose

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+ (for local development)

### Running with Docker Compose (Recommended)
The easiest way to start the entire stack (Redis + Sanguis Server + Metrics):

```bash
docker-compose -f deployments/docker-compose.yaml up --build
```

- **gRPC Server**: `localhost:50051`
- **Prometheus Metrics**: `http://localhost:9090/metrics`

### Local Development
1. Start a Redis instance:
   ```bash
   docker run -p 6379:6379 -d redis:7-alpine
   ```
2. Run the server:
   ```bash
   make run
   ```

## Configuration

The application is configured via `configs/config.yaml` or environment variables:

| Env Var | Description | Default |
|---------|-------------|---------|
| `SERVER_PORT` | gRPC server port | `50051` |
| `METRICS_PORT` | Prometheus metrics port | `9090` |
| `LIMITER_TYPE` | `redis`, `hybrid`, `sliding_window`, `token_bucket` | `hybrid` |
| `REDIS_ADDRESS` | Redis connection string | `localhost:6379` |
| `HYBRID_SYNC_INTERVAL` | Interval for L1 -> L2 sync | `500ms` |

## Development Commands

Use the provided `Makefile` for common tasks:

- `make gen`: Generate gRPC code from proto.
- `make test`: Run integration and unit tests.
- `make lint`: Run golangci-lint.
- `make docker-build`: Build the minimal production Docker image.

## License

[MIT](LICENSE)
