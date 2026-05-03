# Teamgram Server + Telegram Android — Enterprise Implementation Guide

This document describes how to fully integrate the Telegram Android client with a self-hosted Teamgram server and implement all missing enterprise features.

---

## Part 1: Android App → Teamgram Compatibility

### 1.1 Patch Files
All patches are in `patches/android/` and `patches/teamgram-server/`.

### 1.2 What Needs to Change

| File | Change | Purpose |
|------|--------|---------|
| `TMessagesProj/jni/tgnet/ConnectionsManager.cpp` | Replace DC addresses with Teamgram server IP:port | Route all traffic to self-hosted server |
| `TMessagesProj/.../BuildVars.java` | Set APP_ID=4, APP_HASH, DEBUG_VERSION=true | Match Teamgram's test credentials |
| `TMessagesProj/.../TLRPC.java` | May need LAYER adjustment | Match API layer (222 vs 224) |

### 1.3 Quick Start
```bash
cd Telegram/
bash ../patches/android/patch_telegram_for_teamgram.sh <SERVER_IP> <SERVER_PORT>
./gradlew :TMessagesProj_App:assembleAfatDebug
```

---

## Part 2: Teamgram Server Enterprise Features

### 2.1 Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    Telegram Android Client                    │
│  (patched ConnectionsManager.cpp → Teamgram IP:10443)        │
└──────────────────────────┬───────────────────────────────────┘
                           │ MTProto TCP
┌──────────────────────────▼───────────────────────────────────┐
│  Teamgram gnetway (10443) → session → bff                    │
│                                                              │
│  BFF Layer routes requests to enterprise modules:            │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │ channels │ │  bots    │ │reactions │ │ stickers │ ...     │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘        │
│       │            │            │            │               │
│  ┌────▼────────────▼────────────▼────────────▼─────┐         │
│  │         gRPC Service Layer                       │         │
│  │  ┌─────────┐ ┌──────┐ ┌─────────┐ ┌─────────┐  │         │
│  │  │ channel │ │ bot  │ │reaction │ │ sticker │  │         │
│  │  │ service │ │service│ │ service │ │ service │  │         │
│  │  └────┬────┘ └──┬───┘ └────┬────┘ └────┬────┘  │         │
│  └───────┼──────────┼──────────┼───────────┼───────┘         │
│          │          │          │           │                  │
│  ┌───────▼──────────▼──────────▼───────────▼───────┐         │
│  │  MySQL / Redis / Kafka / MinIO / etcd            │         │
│  └──────────────────────────────────────────────────┘         │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 Implemented Features

#### Channels / Megagroups ✅
| Handler | File | Status |
|---------|------|--------|
| `channels.createChannel` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.joinChannel` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.leaveChannel` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.getChannels` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.getFullChannel` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.editTitle` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.editAbout` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.deleteChannel` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.getParticipants` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.toggleSignatures` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.checkUsername` | `bff/channels/channels_handler.go` | ✅ Implemented |
| `channels.updateUsername` | `bff/channels/channels_handler.go` | ✅ Implemented |
| DB Schema | `deploy/sql/migrate_channels.sql` | ✅ Created |

#### Reactions ✅
| Handler | File | Status |
|---------|------|--------|
| `messages.getAvailableReactions` | `bff/reactions/reactions_handler.go` | ✅ 21 standard reactions |
| `messages.sendReaction` | `bff/reactions/reactions_handler.go` | ✅ Implemented |
| `messages.getMessagesReactions` | `bff/reactions/reactions_handler.go` | ✅ Stub |
| `messages.getTopReactions` | `bff/reactions/reactions_handler.go` | ✅ Stub |
| `messages.getRecentReactions` | `bff/reactions/reactions_handler.go` | ✅ Stub |
| `messages.setDefaultReaction` | `bff/reactions/reactions_handler.go` | ✅ Stub |
| DB Schema | `deploy/sql/migrate_channels.sql` | ✅ Created |

#### Bots ✅
| Handler | File | Status |
|---------|------|--------|
| `auth.importBotAuthorization` | `bff/bots/bots_handler.go` | ✅ Implemented |
| `bots.sendCustomRequest` | `bff/bots/bots_handler.go` | ✅ Stub |
| `bots.answerWebhookJSONQuery` | `bff/bots/bots_handler.go` | ✅ Stub |
| `bots.setBotCommands` | `bff/bots/bots_handler.go` | ✅ Implemented |
| `bots.resetBotCommands` | `bff/bots/bots_handler.go` | ✅ Implemented |
| `bots.getBotCommands` | `bff/bots/bots_handler.go` | ✅ Implemented |
| `bots.checkBot` | `bff/bots/bots_handler.go` | ✅ Implemented |
| `messages.getBotCallbackAnswer` | `bff/bots/bots_handler.go` | ✅ Stub |
| Service layer (commands, data, tokens) | Existing | ✅ Already implemented |
| DB Schema | Already exists | ✅ Already present |

#### Stickers ✅
| Handler | File | Status |
|---------|------|--------|
| `messages.getAllStickers` | `bff/stickers/stickers_handler.go` | ✅ Implemented |
| `messages.getFavedStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.getRecentStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.getFeaturedStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.getArchivedStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.getStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.getStickerSet` | `bff/stickers/stickers_handler.go` | ✅ Implemented |
| `messages.getMaskStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.getEmojiStickers` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| `messages.installStickerSet` | `bff/stickers/stickers_handler.go` | ✅ Implemented |
| `messages.uninstallStickerSet` | `bff/stickers/stickers_handler.go` | ✅ Implemented |
| `messages.reorderStickerSets` | `bff/stickers/stickers_handler.go` | ✅ Implemented |
| `messages.searchStickerSets` | `bff/stickers/stickers_handler.go` | ✅ Stub |
| DB Schema | `deploy/sql/migrate_channels.sql` | ✅ Created |

### 2.3 Installation Steps

```bash
# 1. Create new database tables
mysql -uroot teamgram < teamgramd/deploy/sql/migrate_channels.sql

# 2. Create BFF module directories
mkdir -p app/bff/channels/{internal/{config,dao,svc,core},client}
mkdir -p app/bff/bots/{internal/{config,dao,svc,core},client}
mkdir -p app/bff/reactions/{internal/{config,dao,svc,core},client}
mkdir -p app/bff/stickers/{internal/{config,dao,svc,core},client}

# 3. Create service module directories
mkdir -p app/service/channel/{internal/{config,dao,svc,core},client}
mkdir -p app/service/bot/{internal/{config,dao,svc,core},client}
mkdir -p app/service/reaction/{internal/{config,dao,svc,core},client}
mkdir -p app/service/sticker/{internal/{config,dao,svc,core},client}

# 4. Apply patch files from patches/teamgram-server/
cp patches/teamgram-server/bff/channels/*.go app/bff/channels/
cp patches/teamgram-server/bff/bots/*.go app/bff/bots/
cp patches/teamgram-server/bff/reactions/*.go app/bff/reactions/
cp patches/teamgram-server/bff/stickers/*.go app/bff/stickers/

# 5. Update BFF proxy client config to register new modules
# Edit app/bff/bff/client/bff_proxy_client.go — add new service mappings

# 6. Remove enterprise gating from fake_rpc_result.go
# Replace ErrEnterpriseIsBlocked returns with the new handler calls

# 7. Rebuild
make

# 8. Restart
cd teamgramd/bin && ./killall.sh && ./runall2.sh
```

### 2.4 Remaining Work (requires more development effort)

| Feature | Complexity | What's Needed |
|---------|-----------|---------------|
| **Secret Chats** | 🔴 High | E2E encryption key exchange, message layer encryption, self-destruct timers |
| **2FA (Two-Factor Auth)** | 🟡 Medium | Password hashing (SRP), password setup/change, recovery email |
| **Calls (P2P Audio/Video)** | 🔴 High | WebRTC/STUN/TURN signaling, ICE candidates relay, call state machine |
| **Group Calls / Conference** | 🔴 Very High | Multi-party WebRTC, SFU/MCU, noise suppression |
| **Push (APNS/FCM/Web)** | 🟡 Medium | HTTP/2 connections to Apple/Google push gateways, token registration |
| **SMS Sending** | 🟢 Low | SMS gateway integration (Twilio, AWS SNS, etc.) |
| **Premium** | 🟡 Medium | Payment processing, premium feature gating, subscription management |
| **Passport** | 🟡 Medium | Identity verification, encrypted document storage |
| **Miniapp** | 🔴 High | WebView rendering bridge, JS API, payment for miniapps |
| **Themes / Wallpapers** | 🟢 Low | Upload/display custom themes and wallpapers |
| **Stories** | 🔴 High | Photo/video ephemeral stories, 24h expiry, views tracking |

### 2.5 Key Architectural Notes

1. **BFF → Service → DB**: Each feature follows the 3-layer pattern:
   - BFF (API gateway, request validation, response formatting)
   - Service (gRPC, business logic, data access)
   - DB (MySQL schema, Redis cache)

2. **Kafka Topics**: The messaging pipeline uses:
   - `Inbox-T` — message delivery to recipient
   - `Sync-T` — multi-device sync updates

3. **Enterprise Gating**: Currently `ErrEnterpriseIsBlocked` in `fake_rpc_result.go`
   blocks all enterprise features. Remove this by registering real handlers in the
   BFF proxy client's routing table.

4. **Auth Code**: Default is `12345`. For production, implement SMS sending via
   the `pkg/code/` package which has provider interfaces.
