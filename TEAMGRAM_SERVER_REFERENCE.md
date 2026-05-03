# Teamgram Server — Complete Reference

> Source: https://github.com/teamgram/teamgram-server (commit master, cloned May 2026)
> Unofficial open-source MTProto 2.0 server written in Go, compatible with Telegram clients.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Services in Detail](#3-services-in-detail)
4. [Infrastructure Dependencies](#4-infrastructure-dependencies)
5. [Request Flow](#5-request-flow)
6. [Startup Order](#6-startup-order)
7. [Configuration](#7-configuration)
8. [MTProto Protocol Implementation](#8-mtproto-protocol-implementation)
9. [Codec Layer](#9-codec-layer)
10. [Handshake Process](#10-handshake-process)
11. [Deployment](#11-deployment)
12. [Observability](#12-observability)
13. [Build System](#13-build-system)
14. [Project Structure](#14-project-structure)
15. [Enterprise vs Community](#15-enterprise-vs-community)
16. [Contributing & Release Process](#16-contributing--release-process)

---

## 1. Overview

**Teamgram Server** is an unofficial, open-source implementation of Telegram's **MTProto 2.0** protocol written entirely in Go. It allows self-hosting a Telegram-compatible messaging backend. Key facts:

- **Language:** Go 1.23+
- **API Layer:** 222
- **Framework:** Built on [go-zero](https://github.com/zeromicro/go-zero) v1.10.0 (microservice framework with RPC, service discovery, tracing)
- **Networking:** [gnet v2](https://github.com/panjf2000/gnet) (event-loop, non-blocking TCP)
- **Protocols:** [teamgram/proto](https://github.com/teamgram/proto) v0.223.1 — generated MTProto TL types
- **Core lib:** [teamgram/marmota](https://github.com/teamgram/marmota) v0.2.0 — shared utilities
- **License:** Apache 2.0
- **Stars:** ~2.2k | **Forks:** ~427
- **Latest release:** v0.90.4 (Mar 2023)

---

## 2. Architecture

The server is decomposed into **11 microservices** organized in 4 layers:

```
┌─────────────────────────────────────────────────────────────┐
│                   CLIENT (MTProto 2.0)                      │
│        Telegram Android / iOS / TDesktop (patched)          │
└─────────────────────────┬───────────────────────────────────┘
                          │ TCP/WS (10443, 5222, 11443)
┌─────────────────────────▼───────────────────────────────────┐
│  GATEWAY LAYER                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │ gnetway  │  │ gnetway  │  │ gnetway  │  (horizontal     │
│  │ (TCP)    │  │ (WS)     │  │ (HTTP)   │   scale)         │
│  └────┬─────┘  └────┬─────┘  └──────────┘                   │
└───────┼─────────────┼───────────────────────────────────────┘
        │             │
┌───────▼─────────────▼───────────────────────────────────────┐
│  SESSION LAYER                                              │
│  ┌──────────────┐  routes MTProto RPC to BFF               │
│  │   session    │  manages connection state, auth          │
│  └──────┬───────┘                                          │
└─────────┼───────────────────────────────────────────────────┘
          │
┌─────────▼───────────────────────────────────────────────────┐
│  BFF LAYER (Backend-for-Frontend)                           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  bff     translates MTProto RPC → internal gRPC     │   │
│  │  Contains modules: account, authorization, chats,   │   │
│  │  chatinvites, notification, privacy, premium, etc.  │   │
│  └───┬──┬──┬──┬──┬──┬──┬──┬────────────────────────────┘   │
└──────┼──┼──┼──┼──┼──┼──┼──┼────────────────────────────────┘
       │  │  │  │  │  │  │  │
┌──────▼──▼──▼──▼──▼──▼──▼──▼────────────────────────────────┐
│  BACKEND SERVICES                                          │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌─────────┐ ┌──────┐ ┌─────┐  │
│  │ biz  │ │ msg  │ │ sync │ │authsess.│ │ media│ │ dfs │  │
│  │20020 │ │20030 │ │20420 │ │ 20450   │ │20650 │ │20640│  │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──┬──────┘ └──┬───┘ └──┬──┘  │
│  ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴──────┐           ┌──┴──┐ │
│  │idgen │ │status│ │      │ │         │           │     │ │
│  │20660 │ │20670 │ │      │ │         │           │     │ │
│  └──────┘ └──────┘ └──────┘ └─────────┘           └─────┘ │
└────────────────────────────────────────────────────────────┘
```

### Service Ports

| Service | RPC/HTTP Listen | Client-facing | Etcd Key |
|---------|----------------|---------------|----------|
| **gnetway** | 127.0.0.1:20110 | **10443** (TCP), **5222** (TCP), **11443** (WebSocket) | `interface.gateway` |
| **session** | 127.0.0.1:20120 | — | `interface.session` |
| **bff** | 127.0.0.1:20010 | — | `bff.bff` |
| **biz** | 127.0.0.1:20020 | — | `service.biz_service` |
| **msg** | 127.0.0.1:20030 | — | `messenger.msg` |
| **sync** | 127.0.0.1:20420 | — | `messenger.sync` |
| **authsession** | 127.0.0.1:20450 | — | `service.authsession` |
| **dfs** | 127.0.0.1:20640 (gRPC), 0.0.0.0:11701 (HTTP) | — | `service.dfs` |
| **media** | 127.0.0.1:20650 | — | `service.media` |
| **idgen** | 127.0.0.1:20660 | — | `service.idgen` |
| **status** | 127.0.0.1:20670 | — | `service.status` |
| **httpserver** (opt) | — | 8801 (HTTP) | — |

---

## 3. Services in Detail

### 3.1 gnetway (Gateway)
- **Path:** `app/interface/gnetway/`
- **Entry:** `cmd/gnetway/main.go` → calls `gnetway_helper.Server`
- **Server:** `internal/server/server.go` — initializes config, gnet server, gRPC server
- **Networking:** `internal/server/gnet/server.go` — gnet `BuiltinEventEngine`, Goroutine pool, LRU auth key cache, timeout wheel
- **Transports:** TCP, WebSocket, HTTP (configurable per address in YAML)
- **Protocols:** All 4 MTProto transports (Abridged, Intermediate, Padded Intermediate, Full)
- **Config:** `etc/gnetway.yaml` — RSA key file, fingerprint, session client etcd config
- **Config struct:** `internal/config/config.go` — `GnetwayConfig` has `Server[]` with Proto/Addresses, `RSAKey[]` with KeyFile/KeyFingerprint
- **Handshake:** `internal/server/gnet/handshake.go` — full MTProto 2.0 handshake (see section 10)
- **Auth session manager:** `internal/server/gnet/auth_session_manager.go` — manage auth keys in memory

### 3.2 session
- **Path:** `app/interface/session/`
- Receives connections from gnetway, validates auth, routes MTProto RPC to bff
- Manages connection state and session lifecycle

### 3.3 bff (Backend-for-Frontend)
- **Path:** `app/bff/bff/`
- **Entry:** `cmd/bff/main.go`
- **Config:** `internal/config/config.go` — connects to: BizService, AuthSession, Media, Idgen, Msg, Sync (Kafka producer), Dfs, Status
- Has modules (each in `app/bff/<module>/`):
  - `account/` — account management
  - `authorization/` — auth flow
  - `chats/` — chat operations
  - `chatinvites/` — invite handling
  - `notification/` — push notifications, plugin-based
  - `privacy/` — privacy settings
  - `premium/` (enterprise stubs)
  - `sponsoredmessages/` — sponsored messaging
  - `updates/` — update delivery
  - `tos/` — terms of service
  - `qrcode/` — QR code login
  - `passkey/` — passkey auth
  - `passport/` — Telegram passport
  - `nsfw/` — NSFW content filtering
  - `savedmessagedialogs/` — saved messages
  - `autodownload/` — auto-download settings

### 3.4 biz (Business Logic)
- **Path:** `app/service/biz/`
- Core business: user, chat, dialog, message, updates gRPC
- Uses MySQL + Redis
- Called by bff via gRPC

### 3.5 msg (Message Service)
- **Path:** `app/messenger/msg/`
- Message storage, delivery, inbox management
- **Kafka topics:** consumes `Inbox-T`, produces/consumes `Sync-T`
- Uses MySQL + Redis

### 3.6 sync (Sync Service)
- **Path:** `app/messenger/sync/`
- Multi-device sync
- Consumes `Sync-T` from Kafka
- Pushes updates to connected sessions

### 3.7 authsession
- **Path:** `app/service/authsession/`
- Auth key management, session validation
- Stores auth keys in MySQL + Redis cache

### 3.8 dfs (Distributed File System)
- **Path:** `app/service/dfs/`
- Upload/download routing
- Stores files in MinIO buckets: `documents`, `encryptedfiles`, `photos`, `videos`
- Uses Redis for cache/SSDB

### 3.9 media
- **Path:** `app/service/media/`
- Image/document/video metadata and processing
- Uses FFmpeg for transcoding
- Calls dfs for file operations

### 3.10 idgen (ID Generator)
- **Path:** `app/service/idgen/`
- Distributed unique ID generation
- Uses [snowflake](https://github.com/bwmarrin/snowflake) algorithm via Redis

### 3.11 status
- **Path:** `app/service/status/`
- User online/offline status tracking
- Uses Redis

---

## 4. Infrastructure Dependencies

| Service | etcd | MySQL | Redis | Kafka | MinIO |
|---------|------|-------|-------|-------|-------|
| **idgen** | ✓ | — | ✓ (SeqIDGen) | — | — |
| **status** | ✓ | — | ✓ (Status) | — | — |
| **authsession** | ✓ | ✓ | ✓ (Cache, KV) | — | — |
| **dfs** | ✓ | — | ✓ (Cache, SSDB) | — | ✓ |
| **media** | ✓ | ✓ | ✓ (Cache) | — | — |
| **biz** | ✓ | ✓ | ✓ (Cache, KV) | — | — |
| **msg** | ✓ | ✓ | ✓ (Cache, KV) | ✓ (Inbox-T, Sync-T) | — |
| **sync** | ✓ | ✓ | ✓ (Cache, KV) | ✓ (Sync-T consumer) | — |
| **bff** | ✓ | — | ✓ (KV) | ✓ (Sync-T producer) | — |
| **session** | ✓ | — | ✓ (Cache) | — | — |
| **gnetway** | ✓ | — | — | — | — |

### Default Endpoints
| Component | Default Address |
|-----------|----------------|
| etcd | 127.0.0.1:2379 |
| MySQL | 127.0.0.1:3306 |
| Redis | 127.0.0.1:6379 |
| Kafka | 127.0.0.1:9092 |
| MinIO | localhost:9000 |

### Recommended Versions
| Component | Version |
|-----------|---------|
| MySQL | 8.0 |
| Redis | 7.x (alpine) |
| etcd | v3.5.11 |
| Kafka | 3.9.0 (KRaft mode, no ZK) |
| MinIO | latest |
| FFmpeg | latest stable |
| Go | 1.21+ (project uses 1.23) |

### Kafka Topics Created
- `Inbox-T` — partitions:1, replication:1 — message inbox delivery
- `Sync-T` — partitions:1, replication:1 — multi-device sync events
- `teamgram-log` — partitions:1, replication:1 — log pipeline

---

## 5. Request Flow

```
1. Client connects to gnetway via TCP/WebSocket MTProto
2. gnetway performs MTProto 2.0 handshake (PQ, DH, auth key creation)
3. gnetway discovers session service via etcd (interface.session)
4. gnetway forwards encrypted RPC payloads to session
5. Session validates auth key, decodes RPC, discovers bff via etcd (bff.bff)
6. Session routes RPC to bff
7. BFF translates MTProto RPC → internal gRPC calls:
   - Calls biz for user/chat/dialog operations
   - Calls authsession for auth validation
   - Calls media for media processing
   - Calls idgen for unique IDs
   - Calls msg for message storage/delivery
   - Calls dfs for file operations
   - Calls status for online status
   - Publishes to Kafka Sync-T for multi-device updates
8. Responses flow back: bff → session → gnetway → client
```

### Call Relationships
- **gnetway** → session (gRPC/streaming)
- **session** → authsession (gRPC), status (gRPC), bff (gRPC)
- **bff** → biz, authsession, media, idgen, msg, dfs, status (all gRPC); → Kafka Sync-T
- **biz** → media, idgen (gRPC)
- **msg** → idgen, biz (gRPC); Kafka Inbox-T consumer, Sync-T producer/consumer
- **sync** → idgen, status, session, biz (gRPC); Kafka Sync-T consumer
- **media** → dfs (gRPC)
- **dfs** → idgen (gRPC); MinIO for storage
- **authsession** → MySQL, Redis (no external gRPC calls)
- **idgen** → Redis (no external gRPC calls)
- **status** → Redis (no external gRPC calls)

---

## 6. Startup Order

Critical — services must start in this exact order:

| Order | Service | Config File | Depends On |
|-------|---------|-------------|------------|
| 1 | **idgen** | `teamgramd/etc/idgen.yaml` | etcd, Redis |
| 2 | **status** | `teamgramd/etc/status.yaml` | etcd, Redis |
| 3 | **authsession** | `teamgramd/etc/authsession.yaml` | etcd, MySQL, Redis |
| 4 | **dfs** | `teamgramd/etc/dfs.yaml` | etcd, Redis, MinIO |
| 5 | **media** | `teamgramd/etc/media.yaml` | etcd, MySQL, Redis |
| 6 | **biz** | `teamgramd/etc/biz.yaml` | etcd, MySQL, Redis |
| 7 | **msg** | `teamgramd/etc/msg.yaml` | etcd, MySQL, Redis, Kafka |
| 8 | **sync** | `teamgramd/etc/sync.yaml` | etcd, MySQL, Redis, Kafka |
| 9 | **bff** | `teamgramd/etc/bff.yaml` | etcd, Redis, Kafka (producer) |
| 10 | **session** | `teamgramd/etc/session.yaml` | etcd, Redis |
| 11 | **gnetway** | `teamgramd/etc/gnetway.yaml` | etcd |
| (opt) | **httpserver** | `teamgramd/etc/httpserver.yaml` | etcd |

---

## 7. Configuration

### Config Directory Structure
```
teamgramd/
├── etc/          # Manual deployment configs
├── etc2/         # Docker deployment configs
├── bin/          # Compiled binaries + run scripts
├── deploy/
│   ├── sql/      # Database init + migration scripts
│   ├── filebeat/conf/
│   ├── go-stash/etc/
│   ├── prometheus/server/
│   └── ...
```

### Config File Format (YAML)
Each service uses go-zero's config format:

```yaml
# Example: gnetway.yaml
Name: interface.gateway
ListenOn: 127.0.0.1:20110
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: interface.gateway
KeyFile: "./server_pkcs1.key"
KeyFingerprint: "12240908862933197005"
Server:
  Addrs:
    - 0.0.0.0:10443    # TCP MTProto
    - 0.0.0.0:5222     # TCP MTProto
    - 0.0.0.0:8801     # HTTP (optional)
Session:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: interface.session
```

```yaml
# Example: bff.yaml (internal config struct)
Config:
  zrpc.RpcServerConf      # standard go-zero RPC server
  KV kv.KvConf            # Redis KV store
  Code *SmsVerifyCodeConfig
  BizServiceClient zrpc.RpcClientConf
  AuthSessionClient zrpc.RpcClientConf
  MediaClient zrpc.RpcClientConf
  IdgenClient zrpc.RpcClientConf
  MsgClient zrpc.RpcClientConf
  SyncClient *KafkaProducerConf
  DfsClient zrpc.RpcClientConf
  StatusClient zrpc.RpcClientConf
```

### Critical Config Fields per Service
- **dfs.yaml:** `Minio.Endpoint`, `Minio.AccessKeyID`, `Minio.SecretAccessKey`, `SSDB` (Pika or Redis)
- **gnetway.yaml:** `KeyFile` (RSA private key), `KeyFingerprint`, `Server.Addrs`, `Session.Etcd`
- **All services:** `Etcd.Hosts`, MySQL DSN, Redis/Cache, Kafka brokers

---

## 8. MTProto Protocol Implementation

### Supported Transports (MTProto 2.0)
1. **Abridged** — first byte `0xef`, length encoded as 1 byte (÷4) or 0x7f + 3 bytes
2. **Intermediate** — 4-byte little-endian length prefix
3. **Padded Intermediate** — Intermediate with random padding
4. **Full** — 4-byte length + 4-byte seqno + 4-byte CRC32

### API Layer
- **Layer 222** (defined in `teamgram/proto` dependency)
- TL schema definitions in protobuf format (`.tl.proto` files)
- Code generation via `mtprotoc` tool

### Client Compatibility
- Default verification code: `12345` (change for production!)
- Patched clients available for Android, iOS, Desktop (TDesktop)
- Clients must point to self-hosted gnetway addresses

---

## 9. Codec Layer

Located at `app/interface/gnetway/internal/server/gnet/codec/`

### Files
| File | Purpose |
|------|---------|
| `mtproto_abridged_codec.go` | Abridged transport — single-byte length (÷4), no seqno/CRC |
| `mtproto_intermediate_codec.go` | 4-byte LE length, no seqno/CRC |
| `mtproto_padded_intermediate_codec.go` | Same as intermediate + random padding |
| `mtproto_full_codec.go` | Full transport — length + seqno + CRC32 |
| `mtproto_obfuscated_codec.go` | Obfuscation wrapper — random bytes header |
| `mtproto_transport_codec.go` | Interface definitions: `CodecReader`, `CodecWriter`, `Codec` |
| `aes_ctr128_crypto.go` | AES-CTR-128 encryption/decryption for transport |
| `inner_buffer.go` | Buffer management for packet assembly |
| `checker/` | Transport validation |

### Codec Interface
```go
type Codec interface {
    Encode(conn CodecWriter, msg interface{}) ([]byte, error)
    Decode(conn CodecReader) (needAck bool, buf []byte, err error)
    EncodeQuickAck(token uint32) []byte
}
```

### Decode State Machine
Each codec implements a state machine:
- `WAIT_PACKET_LENGTH_1` → reading first length byte(s)
- `WAIT_PACKET_LENGTH_3` → reading extended length (abridged only)
- `WAIT_PACKET_LENGTH_1_PACKET` / `WAIT_PACKET_LENGTH_3_PACKET` → reading the actual data

### Quick Ack
- Client sets MSB of length field to request quick acknowledgment
- Server responds with a 4-byte token (first 32 bits of SHA256 of encrypted payload | 0x80000000)
- Indicates receipt only, not processing completion

---

## 10. Handshake Process

Full MTProto 2.0 key exchange implemented in `app/interface/gnetway/internal/server/gnet/handshake.go`

### Constants (hardcoded in handshake.go)
- **PQ:** `{0x17, 0xED, 0x48, 0x94, 0x1A, 0x08, 0xF9, 0x81}`
- **P:** `{0x49, 0x4C, 0x55, 0x3B}` / **Q:** `{0x53, 0x91, 0x10, 0x73}`
- **DH-2048:** Standard Telegram good prime (256 bytes)
- **G:** `{0x03}`

### Handshake Steps

```
Step 1: req_pq / req_pq_multi
  Client → Server: nonce (128-bit)
  Server → Client: ResPQ { nonce, server_nonce, pq, fingerprints[] }

Step 2: req_DH_params
  Client → Server: { nonce, server_nonce, p, q, fingerprint, encrypted_data }
    encrypted_data = RSA_encrypt(PQ_inner_data)
    PQ_inner_data = { pq, p, q, nonce, server_nonce, new_nonce(256-bit) }
  Server (async):
    - RSA decrypt using matching key
    - Validate pq, p, q, nonce, server_nonce
    - Generate random A (256 bytes)
    - Compute gA = g^A mod p
    - Build Server_DHInnerData { nonce, server_nonce, g, gA, dh_prime, server_time }
    - Encrypt with AES-256-IGE key = SHA1(new_nonce+server_nonce)[:20] + SHA1(server_nonce+new_nonce)[:20] + SHA1(new_nonce+new_nonce)[:20] + new_nonce[:4]
  Server → Client: Server_DH_Params_ok { encrypted_answer }

Step 3: set_client_DH_params
  Client → Server: { nonce, server_nonce, encrypted_data }
    encrypted_data = AES-256-IGE(Client_DHInnerData { nonce, server_nonce, gB })
  Server:
    - Decrypt with same AES key derivation
    - Validate nonce/server_nonce
    - Compute auth_key = gB^A mod p (256 bytes)
    - Compute auth_key_id from new_nonce + SHA1(auth_key)
    - Store auth key in session service via gRPC
  Server → Client: dh_gen_ok { nonce, server_nonce, new_nonce_hash }
```

### Auth Key Storage
- gRPC call to session service: `SessionSetAuthKey(AuthKeyInfo, FutureSalt, ExpiresIn)`
- Also stored in local LRU cache for fast lookup
- Auth key ID derived from: `SHA1(new_nonce + 0x01 + SHA1(auth_key)[:8])`

---

## 11. Deployment

### 11.1 Docker Deployment (Recommended)

```bash
# 1. Clone
git clone https://github.com/teamgram/teamgram-server.git
cd teamgram-server

# 2. Configure
cp .env.example .env
# Edit .env for MySQL/MinIO passwords

# 3. Start dependency stack (MySQL, Redis, etcd, Kafka, MinIO + monitoring)
docker compose -f docker-compose-env.yaml up -d
# DB is auto-initialized from teamgramd/deploy/sql/
# MinIO buckets auto-created by minio-mc container

# 4. Start application services
docker compose up -d
# This builds and starts all 11 services using Dockerfile
```

### 11.2 Manual Linux Installation

```bash
# 1. Install dependencies
# MySQL 8.0, Redis 6.x/7.x, etcd v3.5+, Kafka 2.x/3.x, MinIO, FFmpeg

# 2. Clone & build
git clone https://github.com/teamgram/teamgram-server.git
cd teamgram-server
make
# Binaries → teamgramd/bin/

# 3. Initialize database
mysql -uroot -e "CREATE DATABASE IF NOT EXISTS teamgram;"
mysql -uroot teamgram < teamgramd/deploy/sql/1_teamgram.sql
for f in teamgramd/deploy/sql/migrate-*.sql; do mysql -uroot teamgram < "$f"; done
mysql -uroot teamgram < teamgramd/deploy/sql/z_init.sql

# 4. Configure
# Edit teamgramd/etc/*.yaml to match your environment

# 5. Start (all 11 services in order)
cd teamgramd/bin
./runall2.sh
```

### 11.3 Docker Compose Network
- Network: `teamgram_net` (bridge, 172.20.0.0/16)
- Data persisted under `./data/` (mysql, redis, etcd, kafka, minio, prometheus, grafana, elasticsearch)

---

## 12. Observability

### 12.1 Log Pipeline
```
App Logs (container) → Filebeat → Kafka (teamgram-log) → go-stash → Elasticsearch → Kibana
```
- **Filebeat:** `teamgramd/deploy/filebeat/conf/filebeat.yml`
- **go-stash:** `teamgramd/deploy/go-stash/etc/config.yaml` (filters, transforms)
- **Elasticsearch:** Index pattern `teamgram-{{yyyy-MM-dd}}`
- **Kibana:** http://localhost:5601

### 12.2 Service Monitoring (Prometheus + Grafana)
- Each service exposes `/metrics` endpoint via `Prometheus` config block
- `teamgramd/deploy/prometheus/server/prometheus.yml` defines scrape jobs
- Grafana dashboards: http://localhost:3000 (default admin/admin)

```yaml
# In each service's YAML (teamgramd/etc2/):
Prometheus:
  Host: 0.0.0.0
  Port: 20011    # Unique per service
  Path: /metrics
```

### 12.3 Distributed Tracing (Jaeger)
- go-zero built-in OpenTelemetry tracing
- `Telemetry` block in each service YAML:

```yaml
Telemetry:
  Name: bff.bff
  Endpoint: http://jaeger:14268/api/traces
  Sampler: 1.0
  Batcher: jaeger
```

- Jaeger UI: http://localhost:16686
- Jaeger included in `docker-compose-env.yaml` (with Elasticsearch storage)

---

## 13. Build System

### Makefile Targets
```makefile
make          # Builds all 11 services
make idgen    # Build individual service
make status
make dfs
make media
make authsession
make biz
make msg
make sync
make bff
make session
make gnetway
make httpserver   # optional
make clean
```

### Build Details
- **Version:** v0.96.0-teamgram-server (from VERSION variable)
- **Tags:** `jsoniter` (JSON serializer)
- **Output:** `teamgramd/bin/`
- **LDFlags:** injects gitTag, buildDate, gitCommit, gitTreeState, version, gitBranch
- **Go:** 1.23.0 (from go.mod)

### go.mod Key Dependencies
| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/teamgram/proto` | v0.223.1 | MTProto TL types & codec |
| `github.com/teamgram/marmota` | v0.2.0 | Shared framework utilities |
| `github.com/zeromicro/go-zero` | v1.10.0 | Microservice framework |
| `github.com/panjf2000/gnet/v2` | v2.9.1 | Event-loop networking |
| `github.com/minio/minio-go/v7` | v7.0.49 | S3-compatible storage client |
| `github.com/bwmarrin/snowflake` | v0.3.1 | Unique ID generation |
| `github.com/IBM/sarama` | v1.43.2 | Kafka client |
| `go.etcd.io/etcd/client/v3` | v3.5.15 | Service discovery |
| `github.com/redis/go-redis/v9` | v9.18.0 | Redis client |
| `github.com/disintegration/imaging` | v1.6.3 | Image processing |
| `github.com/chai2010/webp` | v1.4.0 | WebP encoding |
| `github.com/nyaruka/phonenumbers` | v1.6.11 | Phone number parsing |
| `github.com/oschwald/geoip2-golang` | v1.8.0 | GeoIP lookup |
| `google.golang.org/protobuf` | v1.36.11 | Protobuf runtime |

---

## 14. Project Structure

```
teamgram-server/
├── app/
│   ├── bff/                    # BFF service modules
│   │   ├── account/            # Account management
│   │   ├── authorization/      # Auth flow
│   │   ├── autodownload/       # Auto-download settings
│   │   ├── bff/                # BFF server (main entry, config)
│   │   ├── chats/              # Chat operations
│   │   ├── chatinvites/        # Invite handling
│   │   ├── notification/       # Push notifications (plugin-based)
│   │   ├── nsfw/               # NSFW filtering
│   │   ├── passkey/            # Passkey auth
│   │   ├── passport/           # Telegram passport
│   │   ├── premium/            # Premium features (enterprise)
│   │   ├── privacy/            # Privacy settings
│   │   ├── qrcode/             # QR code login
│   │   ├── savedmessagedialogs/ # Saved messages
│   │   ├── sponsoredmessages/  # Sponsored messages
│   │   ├── tos/                # Terms of service
│   │   └── updates/            # Update delivery
│   ├── interface/
│   │   ├── gnetway/            # MTProto gateway (TCP/WS/HTTP)
│   │   │   ├── cmd/gnetway/    # Entry point
│   │   │   ├── internal/
│   │   │   │   ├── config/     # Config structs
│   │   │   │   ├── svc/        # Service context
│   │   │   │   └── server/
│   │   │   │       ├── gnet/   # gnet-based TCP server
│   │   │   │       │   ├── codec/     # Transport codecs
│   │   │   │       │   ├── ws/        # WebSocket
│   │   │   │       │   ├── http/      # HTTP
│   │   │   │       │   ├── pp/        # Proxy protocol
│   │   │   │       │   ├── handshake.go
│   │   │   │       │   ├── server.go
│   │   │   │       │   ├── conn.go
│   │   │   │       │   └── ...
│   │   │   │       └── grpc/  # gRPC server
│   │   │   ├── etc/           # Config files
│   │   │   ├── client/        # Gateway client
│   │   │   └── gateway/       # Generated TL code
│   │   ├── session/           # Session service
│   │   └── httpserver/        # Optional HTTP server
│   ├── messenger/
│   │   ├── msg/               # Message service
│   │   └── sync/              # Sync service
│   └── service/
│       ├── biz/               # Business logic
│       ├── authsession/       # Auth session
│       ├── dfs/               # Distributed file storage
│       ├── media/             # Media processing
│       ├── idgen/             # ID generator
│       └── status/            # Online status
├── clients/                   # Client patch docs
├── docs/                      # Documentation (install guides, topology, monitoring)
├── pkg/                       # Shared packages
│   ├── phonenumber/           # Phone number parsing
│   ├── net2/                  # Network utilities
│   ├── env2/                  # Environment helpers
│   ├── code/                  # SMS verification code
│   ├── hashx/                 # Hash utilities
│   ├── goffmpeg/              # FFmpeg wraper
│   ├── deduplication/         # Message dedup
│   ├── conf/                  # Config helpers
│   ├── mention/               # @mention parsing
│   ├── pubsub/                # Pub/sub
│   └── httpx/                 # HTTP utilities
├── specs/                     # Specs (architecture, protocol, dependencies, etc.)
├── teamgramd/
│   ├── bin/                   # Compiled binaries + run scripts
│   ├── etc/                   # Manual deploy configs
│   ├── etc2/                  # Docker configs
│   └── deploy/                # SQL, monitoring, logging configs
├── docker-compose.yaml        # App docker-compose
├── docker-compose-env.yaml    # Infrastructure docker-compose
├── Dockerfile                 # App Dockerfile
├── Makefile                   # Build targets
├── go.mod / go.sum            # Go modules
└── README.md                  # Project README
```

---

## 15. Enterprise vs Community

### Community Edition (this repo)
- Private chat
- Basic groups
- Contacts
- Web support
- Sign-in with code `12345`

### Enterprise Edition (contact [@benqi](https://t.me/benqi))
- Stickers / themes / chat themes / wallpapers / reactions
- Secret chat / 2FA
- SMS / Push (APNS, Web, FCM) / Web
- Scheduled messages / auto-delete
- Channels / megagroups
- Audio / video / group / conference calls
- Bots
- Mini apps

### Community Roadmap
- Docs and specs maintenance
- CI (GitHub Actions) for build, test, lint
- Unit and integration tests
- Observability runbooks
- More deployment environments (Kubernetes, etc.)

---

## 16. Contributing & Release Process

### Branching
- Default branch: `master`
- Work in forks, create PRs
- One logical change per PR

### Code Style
- Go: `gofmt`, `golangci-lint` when available
- Conventional Commits: `feat:`, `fix:`, `docs:`, etc.

### Release Process
1. Update `VERSION` in Makefile
2. Update CHANGELOG.md (Keep a Changelog format)
3. Create Git tag (e.g., `v0.96.0`)
4. Build and publish binaries/Docker images

### Versioning
- Semantic versioning in spirit: `v0.96.0-teamgram-server`
- Major: incompatible API/architecture
- Minor: backward-compatible features
- Patch: backward-compatible fixes

### Security Reporting
- Report vulnerabilities privately to [@benqi](https://t.me/benqi)
- Do not disclose in public issues
- Security updates for currently maintained major version only
