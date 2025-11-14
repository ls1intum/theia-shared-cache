# cache-node Helm chart

Helm chart to deploy the official Gradle Build Cache Node using the image `gradle/build-cache-node`.

Features

- Uses `gradle/build-cache-node` container image (configurable via `values.yaml`).
- Persistent storage for cache data via a PVC (`persistence.enabled`, `persistence.size`, `persistence.storageClass`).
- Service exposing port (default 5071).
- Configurable extra environment variables and resource limits.

Quickstart

1. Install the chart with default persistence enabled:

```bash
helm install my-cache ./
```

2. To override values (example: use 20Gi and a specific storage class):

```bash
helm install my-cache ./ --set persistence.size=20Gi --set persistence.storageClass=fast
```

Port forwarding / usage

If you deployed as ClusterIP (default), port-forward the service to test locally:

```bash
kubectl port-forward svc/my-cache-cache-node 5071:5071
# then configure your Gradle build cache to point at http://127.0.0.1:5071
```

Notes & tips

- If you need a multi-writer volume (shared between nodes) set `persistence.accessMode=ReadWriteMany` and provide a StorageClass which supports it (e.g. NFS, CephFS).
- The chart provides `extraEnv` for setting environment variables (e.g. authentication, JVM options). Sensitive values should be provided via Kubernetes Secrets and referenced via `volumeMounts` / `volumes` or by customizing templates.
- Adjust `persistence.mountPath` if the upstream image expects a different data directory.

Further reading

- Gradle Build Cache Node — Kubernetes deployment recommendations and examples:
- https://docs.gradle.com/develocity/build-cache-node/#kubernetes

Configuration

- See `values.yaml` for all configurable options.
