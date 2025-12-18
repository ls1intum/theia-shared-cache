# theia-shared-cache Helm Chart

Helm chart to deploy a custom Gradle Build Cache server with MinIO backend for Theia IDE deployments in Kubernetes.

## Architecture

This chart deploys two components:

1. **Cache Server** - A custom Go-based Gradle build cache server
2. **MinIO** - S3-compatible object storage for cache data

```
┌─────────────────────────────────────────────────────┐
│                 Kubernetes Cluster                   │
│                                                      │
│  ┌─────────────────┐       ┌─────────────────────┐  │
│  │  Cache Server   │──────▶│       MinIO         │  │
│  │  (Deployment)   │       │   (StatefulSet)     │  │
│  │  Port: 8080     │       │   Port: 9000        │  │
│  └────────┬────────┘       └─────────────────────┘  │
│           │                                          │
│  ┌────────▼────────┐                                │
│  │    Service      │◀──── Gradle Builds             │
│  │  Port: 8080     │                                │
│  └─────────────────┘                                │
└─────────────────────────────────────────────────────┘
```

## Features

- Custom Go-based cache server with MinIO storage backend
- Prometheus metrics endpoint (`/metrics`)
- Health checks (`/ping`, `/health`)
- Basic authentication for cache operations
- Persistent storage for MinIO data
- Configurable resource limits

## Quick Start

### Install with default values

```bash
helm install gradle-cache ./chart
```

### Install with custom values

```bash
helm install gradle-cache ./chart \
  --set cacheServer.auth.password=mysecretpassword \
  --set minio.auth.accessKey=myaccesskey \
  --set minio.auth.secretKey=mysecretkey \
  --set minio.persistence.size=100Gi
```

## Configuration

### Cache Server

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cacheServer.replicaCount` | Number of cache server replicas | `1` |
| `cacheServer.image.repository` | Cache server image repository | `ghcr.io/kevingruber/theia-shared-cache/gradle-cache` |
| `cacheServer.image.tag` | Cache server image tag | `latest` |
| `cacheServer.port` | Cache server port | `8080` |
| `cacheServer.auth.enabled` | Enable authentication | `true` |
| `cacheServer.auth.username` | Cache username | `gradle` |
| `cacheServer.auth.password` | Cache password | `changeme` |
| `cacheServer.config.maxEntrySizeMB` | Max cache entry size in MB | `100` |
| `cacheServer.resources` | Resource limits/requests | See values.yaml |

### MinIO

| Parameter | Description | Default |
|-----------|-------------|---------|
| `minio.enabled` | Deploy MinIO alongside cache server | `true` |
| `minio.image.repository` | MinIO image repository | `minio/minio` |
| `minio.image.tag` | MinIO image tag | `latest` |
| `minio.auth.accessKey` | MinIO access key | `minioadmin` |
| `minio.auth.secretKey` | MinIO secret key | `minioadmin` |
| `minio.persistence.enabled` | Enable persistent storage | `true` |
| `minio.persistence.size` | Storage size | `50Gi` |
| `minio.persistence.storageClass` | Storage class | `""` (default) |
| `minio.resources` | Resource limits/requests | See values.yaml |

### Metrics

| Parameter | Description | Default |
|-----------|-------------|---------|
| `metrics.enabled` | Enable Prometheus metrics | `true` |
| `metrics.serviceMonitor.enabled` | Create ServiceMonitor (Prometheus Operator) | `false` |

### External Secret

| Parameter | Description | Default |
|-----------|-------------|---------|
| `existingSecret` | Use existing secret for credentials | `""` |

If using `existingSecret`, the secret must contain these keys:
- `minio-access-key`
- `minio-secret-key`
- `cache-password`

## Gradle Configuration

Configure your Gradle build to use the cache:

```kotlin
// settings.gradle.kts
buildCache {
    remote<HttpBuildCache> {
        url = uri("http://<release-name>-cache:8080/cache/")
        credentials {
            username = "gradle"
            password = "your-password"
        }
        isPush = true
    }
}
```

Or via environment variables:

```bash
export GRADLE_CACHE_URL=http://<release-name>-cache:8080/cache/
export GRADLE_CACHE_USERNAME=gradle
export GRADLE_CACHE_PASSWORD=your-password
```

## Port Forwarding (Development)

For local testing, port-forward the cache service:

```bash
kubectl port-forward svc/<release-name>-cache 8080:8080

# Test the cache server
curl http://localhost:8080/ping
curl http://localhost:8080/health
```

## Monitoring

If `metrics.enabled` is true, Prometheus metrics are available at `/metrics`:

```bash
curl http://<release-name>-cache:8080/metrics
```

Available metrics:
- `cache_requests_total` - Total cache requests by method and status
- `cache_hits_total` - Cache hit count
- `cache_misses_total` - Cache miss count
- `cache_request_duration_seconds` - Request duration histogram
- `cache_entry_size_bytes` - Cache entry size histogram

## Upgrading

### From 0.1.x to 0.2.0

Version 0.2.0 is a breaking change that replaces the official Gradle cache node with a custom implementation.

Changes:
- New image: custom Go-based cache server instead of `gradle/build-cache-node`
- New storage: MinIO backend instead of local PersistentVolume
- New port: 8080 instead of 5071
- New authentication: Basic auth built into cache server

Migration steps:
1. Back up any important cache data (usually safe to lose)
2. Uninstall the old release: `helm uninstall <release-name>`
3. Install the new version with updated values
4. Update Gradle configuration to use new port and credentials
