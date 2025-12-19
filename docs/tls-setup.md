# TLS Setup Guide

This guide explains how to enable TLS/HTTPS for the Gradle Build Cache Server to secure communication between clients and the cache server.

## Table of Contents

- [Why Enable TLS?](#why-enable-tls)
- [Prerequisites](#prerequisites)
- [Option 1: Using cert-manager (Recommended)](#option-1-using-cert-manager-recommended)
- [Option 2: Using Self-Signed Certificates](#option-2-using-self-signed-certificates)
- [Option 3: Using External Certificates](#option-3-using-external-certificates)
- [Verifying TLS Configuration](#verifying-tls-configuration)
- [Client Configuration](#client-configuration)
- [Troubleshooting](#troubleshooting)

## Why Enable TLS?

Without TLS, all communication with the cache server is unencrypted, including:
- HTTP Basic Authentication credentials
- Cached build artifacts
- Metadata about your builds

Enabling TLS provides:
- **Encryption** - All traffic is encrypted in transit
- **Authentication** - Verify you're connecting to the legitimate cache server
- **Integrity** - Prevent tampering with cached artifacts

## Prerequisites

Before enabling TLS, ensure you have:
- Kubernetes cluster (v1.19+)
- Helm 3.x installed
- Cache server already deployed (or ready to deploy)
- One of the following:
  - cert-manager installed (for automatic certificate management)
  - Your own TLS certificates
  - OpenSSL installed (for self-signed certificates)

## Option 1: Using cert-manager (Recommended)

cert-manager automates certificate creation, renewal, and management.

### Step 1: Install cert-manager

If not already installed:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

Verify installation:
```bash
kubectl get pods -n cert-manager
```

### Step 2: Create a ClusterIssuer

For Let's Encrypt (production):

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
```

For self-signed certificates (testing):

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
```

Apply the issuer:
```bash
kubectl apply -f clusterissuer.yaml
```

### Step 3: Configure Helm Values

Update `chart/values.yaml`:

```yaml
cacheServer:
  tls:
    enabled: true
    secretName: gradle-cache-tls
    certManager:
      enabled: true
      issuerName: "selfsigned-issuer"  # or "letsencrypt-prod"
      issuerKind: "ClusterIssuer"
```

### Step 4: Deploy with Helm

```bash
helm upgrade --install theia-cache ./chart
```

cert-manager will automatically:
1. Create a Certificate resource
2. Request a certificate from the issuer
3. Store the certificate in the specified secret
4. Renew the certificate before expiration

### Step 5: Verify Certificate

```bash
kubectl get certificate
kubectl describe certificate theia-cache-tls
kubectl get secret gradle-cache-tls
```

## Option 2: Using Self-Signed Certificates

For development or internal use without cert-manager.

### Step 1: Generate Self-Signed Certificate

```bash
# Generate private key
openssl genrsa -out tls.key 2048

# Generate certificate (valid for 365 days)
openssl req -new -x509 -key tls.key -out tls.crt -days 365 \
  -subj "/CN=theia-cache-cache-server.default.svc.cluster.local" \
  -addext "subjectAltName=DNS:theia-cache-cache-server,DNS:theia-cache-cache-server.default,DNS:theia-cache-cache-server.default.svc,DNS:theia-cache-cache-server.default.svc.cluster.local"
```

**Note**: Replace `default` with your namespace and `theia-cache` with your release name.

### Step 2: Create Kubernetes Secret

```bash
kubectl create secret tls gradle-cache-tls \
  --cert=tls.crt \
  --key=tls.key \
  --namespace=default
```

### Step 3: Configure Helm Values

Update `chart/values.yaml`:

```yaml
cacheServer:
  tls:
    enabled: true
    secretName: gradle-cache-tls
    certManager:
      enabled: false  # Not using cert-manager
```

### Step 4: Deploy with Helm

```bash
helm upgrade --install theia-cache ./chart
```

### Step 5: Trust the Certificate (Clients)

Since the certificate is self-signed, clients will need to either:

**Option A**: Trust the certificate
```bash
# Copy tls.crt to your client machine
sudo cp tls.crt /usr/local/share/ca-certificates/gradle-cache.crt
sudo update-ca-certificates
```

**Option B**: Configure Gradle to skip verification (NOT RECOMMENDED for production)
```groovy
// settings.gradle
buildCache {
    remote(HttpBuildCache) {
        url = 'https://theia-cache-cache-server:8080/cache/'
        allowUntrustedServer = true  // Only for self-signed certs
        credentials {
            username = 'gradle'
            password = 'your-password'
        }
    }
}
```

## Option 3: Using External Certificates

If you have certificates from a trusted CA (e.g., purchased SSL certificate).

### Step 1: Prepare Certificate Files

Ensure you have:
- `tls.crt` - Your certificate (including intermediate certificates)
- `tls.key` - Your private key

### Step 2: Create Kubernetes Secret

```bash
kubectl create secret tls gradle-cache-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  --namespace=default
```

### Step 3: Configure Helm Values

```yaml
cacheServer:
  tls:
    enabled: true
    secretName: gradle-cache-tls
    certManager:
      enabled: false
```

### Step 4: Deploy with Helm

```bash
helm upgrade --install theia-cache ./chart
```

## Verifying TLS Configuration

### Check Server Logs

```bash
kubectl logs -f deployment/theia-cache-cache-server
```

Look for:
```
{"level":"info","addr":":8080","mode":"https","message":"starting server with TLS"}
```

### Test with curl

```bash
# Port-forward the service
kubectl port-forward svc/theia-cache-cache-server 8080:8080

# Test HTTPS endpoint
curl -k https://localhost:8080/ping
# Expected: pong
```

### Test with OpenSSL

```bash
openssl s_client -connect localhost:8080 -showcerts
```

### Check Health Probes

```bash
kubectl describe pod -l app.kubernetes.io/component=cache-server
```

Ensure liveness and readiness probes are succeeding.

## Client Configuration

### Gradle Configuration with TLS

Update your `settings.gradle`:

```groovy
buildCache {
    remote(HttpBuildCache) {
        url = 'https://theia-cache-cache-server:8080/cache/'
        credentials {
            username = 'gradle'
            password = System.getenv('GRADLE_CACHE_PASSWORD')
        }
        push = true
    }
}
```

Or `gradle.properties`:
```properties
org.gradle.caching=true
org.gradle.caching.remote.url=https://theia-cache-cache-server:8080/cache/
org.gradle.caching.remote.username=gradle
org.gradle.caching.remote.password=${GRADLE_CACHE_PASSWORD}
org.gradle.caching.remote.push=true
```

### CI/CD Configuration

**GitHub Actions:**
```yaml
- name: Build with Gradle
  env:
    GRADLE_CACHE_PASSWORD: ${{ secrets.GRADLE_CACHE_PASSWORD }}
  run: ./gradlew build --build-cache
```

**GitLab CI:**
```yaml
build:
  script:
    - ./gradlew build --build-cache
  variables:
    GRADLE_CACHE_PASSWORD: $GRADLE_CACHE_PASSWORD
```

**Jenkins:**
```groovy
withCredentials([string(credentialsId: 'gradle-cache-password', variable: 'GRADLE_CACHE_PASSWORD')]) {
    sh './gradlew build --build-cache'
}
```

## Troubleshooting

### Certificate Not Found

**Error**: `failed to read config file: server.tls.cert_file is required when TLS is enabled`

**Solution**: Ensure the TLS secret exists and is mounted correctly:
```bash
kubectl get secret gradle-cache-tls
kubectl describe pod theia-cache-cache-server-xxx
```

### Health Probes Failing

**Error**: Readiness probe failed: HTTP probe failed

**Solution**: Check if health probes are using the correct scheme (HTTPS):
```bash
kubectl get pod theia-cache-cache-server-xxx -o yaml | grep -A 5 livenessProbe
```

Should show `scheme: HTTPS`.

### Certificate Expired

**Error**: `x509: certificate has expired`

**Solution**:
- If using cert-manager, it should auto-renew. Check cert-manager logs.
- If using manual certificates, regenerate and update the secret.

### Certificate Name Mismatch

**Error**: `x509: certificate is valid for X, not Y`

**Solution**: Regenerate certificate with correct DNS names in SAN field:
```bash
-addext "subjectAltName=DNS:service-name,DNS:service-name.namespace.svc.cluster.local"
```

### Self-Signed Certificate Rejected

**Error**: `unable to get local issuer certificate`

**Solution**: Either:
1. Add certificate to client's trust store
2. Use `allowUntrustedServer = true` in Gradle (development only)

### TLS Handshake Errors

**Error**: `TLS handshake timeout`

**Solution**: Ensure the certificate and key match:
```bash
openssl x509 -noout -modulus -in tls.crt | openssl md5
openssl rsa -noout -modulus -in tls.key | openssl md5
# Should produce identical MD5 hashes
```

## Best Practices

1. **Use cert-manager for production** - Automates certificate renewal
2. **Use Let's Encrypt for public services** - Free, trusted certificates
3. **Use self-signed only for development** - Requires manual trust management
4. **Monitor certificate expiration** - Set up alerts 30 days before expiry
5. **Rotate certificates regularly** - Even if not expired
6. **Secure private keys** - Use Kubernetes RBAC to restrict secret access
7. **Use strong key sizes** - Minimum 2048-bit RSA or 256-bit ECDSA

## Additional Resources

- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Kubernetes TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets)
- [Let's Encrypt](https://letsencrypt.org/)
- [Gradle Build Cache Documentation](https://docs.gradle.org/current/userguide/build_cache.html)
