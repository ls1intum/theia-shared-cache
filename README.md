# Gradle Build Cache Server

A high-performance, Kubernetes-native HTTP build cache server for Gradle builds. Built with Go and backed by MinIO for reliable, persistent storage.


## Overview

This project provides a lightweight, self-hosted Gradle Build Cache server implementation that can be deployed in Kubernetes clusters to accelerate build times across teams and CI/CD pipelines. It implements the Gradle HTTP Build Cache API and uses MinIO as a persistent storage backend.

### Key Features

- **Gradle HTTP Build Cache API** - Fully compatible with Gradle's remote cache protocol
- **Persistent Storage** - Uses MinIO (S3-compatible) for reliable, scalable artifact storage
- **Authentication** - HTTP Basic Authentication for access control
- **Kubernetes-Native** - Designed for containerized deployments with production-ready Helm charts
- **Observability** - Built-in Prometheus metrics and structured logging (JSON/text)
- **Lightweight** - Minimal resource footprint (~256Mi RAM, ~100m CPU)
- **Health Checks** - Kubernetes-ready liveness and readiness probes
- **Easy Deployment** - Single Helm command to get started

### Use Cases

- **Team Development** - Share build cache across development teams
- **CI/CD Pipelines** - Accelerate build times in Jenkins, GitLab CI, GitHub Actions, etc.
- **Monorepos** - Reduce build times for large multi-module projects
- **Multi-Environment Builds** - Share cache between dev, staging, and production builds
- **Cost Optimization** - Reduce build server compute costs by avoiding redundant work

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                     │
│                                                          │
│  ┌─────────────────┐       ┌─────────────────────┐      │
│  │  Cache Server   │──────▶│       MinIO         │      │
│  │  (Deployment)   │       │   (StatefulSet)     │      │
│  │  Port: 8080     │       │   Port: 9000        │      │
│  └────────┬────────┘       └─────────────────────┘      │
│           │                                              │
│  ┌────────▼────────┐                                    │
│  │    Service      │◀──── Developer Workstations        │
│  │  Port: 8080     │◀──── CI/CD Pipelines              │
│  └─────────────────┘◀──── Build Agents                 │
└─────────────────────────────────────────────────────────┘
```

### Components

| Component | Description | Technology |
|-----------|-------------|------------|
| **Cache Server** | HTTP API server implementing Gradle Build Cache protocol | Go, Gin Framework |
| **Storage Backend** | S3-compatible object storage for cache artifacts | MinIO |
| **Kubernetes Service** | ClusterIP service for internal access | Kubernetes |
| **Helm Chart** | Declarative deployment configuration | Helm 3 |

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.19+)
- Helm 3.x
- kubectl configured to access your cluster
- (Optional) Domain name and TLS certificates for production deployments

### Installation

1. Clone the repository:
```bash
git clone https://github.com/kevingruber/theia-shared-cache.git
cd theia-shared-cache
```

2. Configure your deployment by editing `chart/values.yaml`:
```yaml
cacheServer:
  auth:
    username: "gradle"
    password: "your-secure-password"  # CHANGE THIS!

minio:
  auth:
    accessKey: "your-access-key"      # CHANGE THIS!
    secretKey: "your-secret-key"      # CHANGE THIS!
```

3. Deploy using Helm:
```bash
helm install theia-cache ./chart
```

4. Verify the deployment:
```bash
kubectl get pods
kubectl logs -f deployment/theia-cache
```

### Testing the Cache

Port-forward the service to your local machine:
```bash
kubectl port-forward svc/theia-cache 8080:8080
```

Test the health endpoint:
```bash
curl http://localhost:8080/ping
# Expected response: pong
```

Test cache operations:
```bash
# Store a cache entry
echo "test data" | curl -u gradle:your-password \
  -X PUT \
  -H "Content-Type: application/octet-stream" \
  --data-binary @- \
  http://localhost:8080/cache/test-key

# Retrieve the cache entry
curl -u gradle:your-password http://localhost:8080/cache/test-key
```

## Configuration

### Gradle Configuration

Configure your Gradle builds to use the remote cache by adding to `settings.gradle`:

```groovy
buildCache {
    remote(HttpBuildCache) {
        url = 'http://theia-cache:8080/cache/'
        credentials {
            username = 'gradle'
            password = 'your-password'
        }
        push = true
    }
}
```

Or via `gradle.properties`:
```properties
org.gradle.caching=true
org.gradle.caching.remote.url=http://theia-cache:8080/cache/
org.gradle.caching.remote.username=gradle
org.gradle.caching.remote.password=your-password
org.gradle.caching.remote.push=true
```

### Helm Chart Configuration

Key configuration options in `chart/values.yaml`:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cacheServer.replicaCount` | Number of cache server replicas | `1` |
| `cacheServer.image.repository` | Container image repository | `ghcr.io/kevingruber/theia-shared-cache/gradle-cache` |
| `cacheServer.image.tag` | Container image tag | `latest` |
| `cacheServer.tls.enabled` | Enable TLS/HTTPS | `false` |
| `cacheServer.tls.secretName` | TLS certificate secret name | `""` |
| `cacheServer.tls.certManager.enabled` | Use cert-manager for certificates | `false` |
| `cacheServer.auth.username` | Authentication username | `gradle` |
| `cacheServer.auth.password` | Authentication password | `changeme` |
| `cacheServer.config.maxEntrySizeMB` | Maximum cache entry size | `100` |
| `minio.enabled` | Deploy MinIO with the chart | `true` |
| `minio.persistence.size` | MinIO storage size | `50Gi` |
| `minio.auth.accessKey` | MinIO access key | `minioadmin` |
| `minio.auth.secretKey` | MinIO secret key | `minioadmin` |

For a complete list of configuration options, see the [Helm chart documentation](chart/README.md).

### Enabling TLS/HTTPS

To secure communication with TLS, you can use cert-manager or provide your own certificates. Here's a quick example:

**With cert-manager (recommended):**
```yaml
cacheServer:
  tls:
    enabled: true
    secretName: gradle-cache-tls
    certManager:
      enabled: true
      issuerName: "letsencrypt-prod"
      issuerKind: "ClusterIssuer"
```

**With self-signed certificates:**
```bash
# Generate certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt \
  -subj "/CN=cache-server.default.svc.cluster.local"

# Create Kubernetes secret
kubectl create secret tls gradle-cache-tls --cert=tls.crt --key=tls.key

# Update values.yaml
cacheServer:
  tls:
    enabled: true
    secretName: gradle-cache-tls
```

For detailed TLS setup instructions, see the [TLS Setup Guide](docs/tls-setup.md).

### Environment Variables

The cache server can be configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_READ_TIMEOUT` | Request read timeout | `30s` |
| `SERVER_WRITE_TIMEOUT` | Request write timeout | `120s` |
| `MINIO_ENDPOINT` | MinIO server endpoint | `minio:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | - |
| `MINIO_SECRET_KEY` | MinIO secret key | - |
| `CACHE_PASSWORD` | Cache authentication password | - |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json, text) | `json` |

## API Reference

### Endpoints

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/ping` | GET | Health check (liveness probe) | No |
| `/health` | GET | Storage health check (readiness probe) | No |
| `/metrics` | GET | Prometheus metrics | No |
| `/cache/:key` | GET | Retrieve cache entry | Yes |
| `/cache/:key` | PUT | Store cache entry | Yes |
| `/cache/:key` | HEAD | Check if cache entry exists | Yes |

### HTTP Status Codes

| Code | Description |
|------|-------------|
| `200 OK` | Cache hit (GET), entry exists (HEAD) |
| `201 Created` | Cache entry stored successfully (PUT) |
| `204 No Content` | Entry does not exist (HEAD) |
| `401 Unauthorized` | Authentication failed |
| `404 Not Found` | Cache miss (GET) |
| `413 Payload Too Large` | Entry exceeds maximum size |
| `500 Internal Server Error` | Server or storage error |

## Development

### Building from Source

#### Prerequisites
- Go 1.24+
- Docker (optional, for containerization)

#### Build the binary
```bash
cd src
go build -o bin/cache-server ./cmd/server
```

#### Run locally
```bash
# Set required environment variables
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin
export CACHE_PASSWORD=changeme

# Run the server
./bin/cache-server
```

#### Build Docker image
```bash
cd src
docker build -t theia-cache:dev .
```

### Testing

Run unit tests:
```bash
cd src
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

### Project Structure

```
.
├── chart/                  # Helm chart for Kubernetes deployment
│   ├── templates/         # Kubernetes manifests
│   ├── values.yaml        # Default configuration
│   └── Chart.yaml         # Chart metadata
├── src/                   # Go source code
│   ├── cmd/
│   │   └── server/        # Main application entry point
│   ├── internal/
│   │   ├── config/        # Configuration management
│   │   ├── middleware/    # HTTP middleware (auth, logging)
│   │   ├── server/        # HTTP server and routes
│   │   └── storage/       # MinIO storage backend
│   ├── Dockerfile         # Multi-stage Docker build
│   └── go.mod             # Go module dependencies
└── .github/
    └── workflows/         # GitHub Actions CI/CD
```

## Monitoring

### Prometheus Metrics

The cache server exposes Prometheus metrics at `/metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `cache_requests_total` | Counter | Total number of cache requests |
| `cache_hits_total` | Counter | Number of cache hits |
| `cache_misses_total` | Counter | Number of cache misses |
| `cache_errors_total` | Counter | Number of cache errors |
| `http_request_duration_seconds` | Histogram | HTTP request latency |

### ServiceMonitor

If using Prometheus Operator, enable the ServiceMonitor in `values.yaml`:
```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

## Security Considerations

### Current Limitations

- **TLS Optional** - TLS/HTTPS is disabled by default. Enable it for production deployments.
- **Basic Authentication** - Simple username/password authentication.
- **Single User** - Only one set of credentials supported.
- **Single Replica** - No built-in high availability (single point of failure).

### Recommendations for Production

1. **Enable TLS** - Encrypt all traffic (see [TLS Setup Guide](docs/tls-setup.md))
2. **Change default credentials** - Use strong, randomly-generated passwords
3. **Network Policies** - Restrict access to authorized pods only
4. **Use Kubernetes Secrets** - Never commit credentials to version control
5. **Monitor certificate expiration** - Set up alerts for certificate renewal
6. **Regular updates** - Keep dependencies and base images up to date

## Troubleshooting

### Common Issues

**Pods not starting**
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

**MinIO connection errors**
- Verify MinIO is running: `kubectl get pods -l app=minio`
- Check MinIO logs: `kubectl logs -l app=minio`
- Verify credentials match in both cache and MinIO configurations

**Authentication failures**
- Ensure credentials in Gradle match those in `values.yaml`
- Check for special characters that need URL encoding

**Out of storage**
- Increase MinIO PVC size in `values.yaml`
- Clean old cache entries manually via MinIO console