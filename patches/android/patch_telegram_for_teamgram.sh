#!/bin/bash
# =============================================================================
# Telegram Android → Teamgram Server Compatibility Patch
# =============================================================================
# This script patches the Telegram Android client to connect to a self-hosted
# Teamgram server. Run from the Telegram repo root.
#
# Usage: bash patch_telegram_for_teamgram.sh [SERVER_IP] [SERVER_PORT]
# Example: bash patch_telegram_for_teamgram.sh 192.168.1.100 10443
# =============================================================================

set -e

SERVER_IP="${1:-127.0.0.1}"
SERVER_PORT="${2:-10443}"

echo "==> Patching Telegram Android for Teamgram Server at ${SERVER_IP}:${SERVER_PORT}"

# ---------------------------------------------------------------------------
# 1. Patch BuildVars.java — set APP_ID and APP_HASH for Teamgram
# ---------------------------------------------------------------------------
BUILDVARS="TMessagesProj/src/main/java/org/telegram/messenger/BuildVars.java"

echo "==> Patching ${BUILDVARS}"

sed -i \
  -e 's/public static int APP_ID = [0-9]*;/public static int APP_ID = 4;/' \
  -e 's/public static String APP_HASH = ".*";/public static String APP_HASH = "014b35b6184100b085b0d0572f9b5103";/' \
  -e 's/public static boolean DEBUG_VERSION = .*/public static boolean DEBUG_VERSION = true;/' \
  -e 's/public static boolean LOGS_ENABLED = .*/public static boolean LOGS_ENABLED = true;/' \
  -e 's/public static boolean CHECK_UPDATES = .*/public static boolean CHECK_UPDATES = false;/' \
  -e 's/public static boolean USE_CLOUD_STRINGS = .*/public static boolean USE_CLOUD_STRINGS = false;/' \
  "$BUILDVARS"

echo "  APP_ID=4, APP_HASH=014b35b6184100b085b0d0572f9b5103 (Teamgram test credentials)"
echo "  DEBUG_VERSION=true, CHECK_UPDATES=false, USE_CLOUD_STRINGS=false"

# ---------------------------------------------------------------------------
# 2. Patch ConnectionsManager.cpp — replace Telegram DCs with Teamgram server
# ---------------------------------------------------------------------------
CPP_FILE="TMessagesProj/jni/tgnet/ConnectionsManager.cpp"

echo "==> Patching ${CPP_FILE}"

# Replace initDatacenters() to point all DCs to our Teamgram server
# We use a heredoc to inject the new function body
python3 << PYEOF
# Read the file
with open("${CPP_FILE}", "r") as f:
    content = f.read()

# Find the initDatacenters function and replace it
old_marker = "void ConnectionsManager::initDatacenters() {"
idx = content.find(old_marker)
if idx == -1:
    print("ERROR: Could not find initDatacenters()")
    exit(1)

# Find the end of the function (matching braces)
brace_count = 0
end_idx = idx + len(old_marker)
for i in range(end_idx, len(content)):
    if content[i] == '{':
        brace_count += 1
    elif content[i] == '}':
        brace_count -= 1
        if brace_count == 0:
            end_idx = i + 1
            break

new_function = """void ConnectionsManager::initDatacenters() {
    Datacenter *datacenter;
    // Custom Teamgram server datacenters
    // All DCs point to the same server for self-hosted deployment
    const char *serverIp = "${SERVER_IP}";
    int serverPort = ${SERVER_PORT};

    // DC 1 — Primary
    if (datacenters.find(1) == datacenters.end()) {
        datacenter = new Datacenter(instanceNum, 1);
        datacenter->addAddressAndPort(serverIp, serverPort, 0, "");
        datacenters[1] = datacenter;
    }

    // DC 2 — Secondary (same server for self-hosted)
    if (datacenters.find(2) == datacenters.end()) {
        datacenter = new Datacenter(instanceNum, 2);
        datacenter->addAddressAndPort(serverIp, serverPort, 0, "");
        datacenters[2] = datacenter;
    }

    // DC 3 — Media
    if (datacenters.find(3) == datacenters.end()) {
        datacenter = new Datacenter(instanceNum, 3);
        datacenter->addAddressAndPort(serverIp, serverPort, 0, "");
        datacenters[3] = datacenter;
    }

    // DC 4 — Files
    if (datacenters.find(4) == datacenters.end()) {
        datacenter = new Datacenter(instanceNum, 4);
        datacenter->addAddressAndPort(serverIp, serverPort, 0, "");
        datacenters[4] = datacenter;
    }

    // DC 5 — Push
    if (datacenters.find(5) == datacenters.end()) {
        datacenter = new Datacenter(instanceNum, 5);
        datacenter->addAddressAndPort(serverIp, serverPort, 0, "");
        datacenters[5] = datacenter;
    }

    // Test DC (same as production for self-hosted)
    if (datacenters.find(-1) == datacenters.end()) {
        datacenter = new Datacenter(instanceNum, -1);
        datacenter->addAddressAndPort(serverIp, serverPort, 0, "");
        datacenters[-1] = datacenter;
    }
}
"""

content = content[:idx] + new_function + content[end_idx:]
with open("${CPP_FILE}", "w") as f:
    f.write(content)

print("  initDatacenters() patched: all DCs → {}:{}".format("${SERVER_IP}", "${SERVER_PORT}"))
PYEOF

# ---------------------------------------------------------------------------
# 3. Patch TLRPC.java LAYER if needed (Teamgram supports layer 222/223)
# ---------------------------------------------------------------------------
TLRPC_FILE="TMessagesProj/src/main/java/org/telegram/tgnet/TLRPC.java"

echo "==> Checking API layer compatibility"
CURRENT_LAYER=$(grep -oP 'LAYER\s*=\s*\d+' "$TLRPC_FILE" | grep -oP '\d+')
echo "  Current API Layer: ${CURRENT_LAYER} (Telegram)"
echo "  Target API Layer: 222 (Teamgram)"
echo "  Note: If client layer > server layer, some features may not work."
echo "  Teamgram may need proto update to match."

# ---------------------------------------------------------------------------
# 4. Handle Teamgram's custom auth code (default: 12345)
# ---------------------------------------------------------------------------
echo ""
echo "==> IMPORTANT:"
echo "  Teamgram default verification code: 12345"
echo "  Change this in teamgram-server config for production."
echo ""
echo "==> Patch complete!"
echo "  Server: ${SERVER_IP}:${SERVER_PORT}"
echo "  Build with: ./gradlew :TMessagesProj_App:assembleAfatDebug"
echo ""
echo "  Note: Also set up your own APP_ID at https://core.telegram.org/api/obtaining_api_id"
echo "  for production use."
