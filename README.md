# StorageVault

A peer-to-peer distributed file storage system built from scratch in Go.
Files stored on one node are automatically encrypted and replicated to all
connected peers over raw TCP. Any node can retrieve any file even if it no
longer has a local copy, by fetching it from the network.

## Architecture

```
Client → Node A

↙      ↘

Node B    Node C
```

- Custom TCP transport layer with a binary application-layer protocol
- Two communication modes: control messages (GOB encoded) and raw file streams
- Stream coordination via WaitGroups to prevent protocol corruption
- AES-256-CTR encryption with a random IV per file
- Content Addressable Storage (CAS) with hierarchical path sharding
- io.Pipe streaming for O(1) memory usage regardless of file size

## Key Engineering Decisions

**io.Pipe streaming (O(1) memory)**
Original design buffered the entire file in a `bytes.Buffer` before streaming.
Refactored to `io.Pipe` so disk write and network stream happen simultaneously.
The pipe blocks the writer goroutine when the reader hasn't consumed, giving
native backpressure. Verified: 1GB+ transfers within a 256Mi memory limit.

**Custom binary protocol over TCP**
Instead of HTTP, a 1-byte signal prefix separates control messages (`0x1`)
from raw file streams (`0x2`) on the same TCP connection — a standard
distributed systems pattern for separating the metadata and data planes.

**Content Addressable Storage**
Files are stored by SHA1 hash of their key, split into 5-character directory
chunks (`18b73/e82a1/...`). Deterministic placement, uniform IDs, and the
same pattern Git uses for object storage.

## Project Structure

```
.

├── main.go              # Entry point, demo setup and GKE env var mode

├── server.go            # FileServer: Store, Get, broadcast, peer management

├── store.go             # CAS disk storage and path transform

├── crypto.go            # AES-256-CTR encrypt/decrypt streaming

├── p2p/

│   ├── transport.go     # Transport and Peer interfaces

│   ├── tcp_transport.go # TCP implementation, connection handling

│   ├── message.go       # RPC struct, IncomingMessage/IncomingStream constants

│   ├── encoding.go      # GOB and default decoder

│   └── handshake.go     # Handshake interface

└── k8s/

└── k8s.yaml         # StatefulSet + Headless Service + PVCs
```

## Running Locally

**Demo mode (3 nodes, automatic peer connections):**

```bash
go run .
```

**Run tests:**

```bash
go test ./...
```

## Kubernetes Deployment (minikube)

```bash
# Start minikube
minikube start --driver=docker --memory=2048 --cpus=2

# Point Docker at minikube's daemon
eval $(minikube docker-env)

# Build image inside minikube
docker build -t storagevault:v1 .

# Deploy
kubectl apply -f k8s/k8s.yaml

# Verify
kubectl get pods
kubectl get pvc
kubectl logs vault-2
```

## What the Kubernetes deployment proves

- **StatefulSet** gives each pod a stable ordinal hostname (`vault-0`, `vault-1`, `vault-2`)
- **Headless Service** (`clusterIP: None`) exposes stable DNS: `vault-0.storage-service`
- **PersistentVolumeClaims** (10Gi per pod) survive pod restarts — data is durable
- **256Mi memory limit** enforces the O(1) streaming design — no OOM possible

## Known Limitations (production improvements)

- `time.Sleep` in `Get()` instead of a proper response channel
- `io.MultiWriter` blocks on slowest peer — production would use per-peer goroutines
- MD5 used for path key hashing — SHA-256 would be correct for production
- No liveness/readiness probes in `k8s.yaml`
