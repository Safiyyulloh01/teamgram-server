#!/bin/bash
set -e
SERVER_IP="${1:-127.0.0.1}"
SERVER_PORT="${2:-10443}"

echo "==> Patching Telegram for Teamgram at ${SERVER_IP}:${SERVER_PORT}"

# Patch BuildVars.java
BUILDVARS="TMessagesProj/src/main/java/org/telegram/messenger/BuildVars.java"
sed -i \
  -e 's/public static int APP_ID = [0-9]*;/public static int APP_ID = 4;/' \
  -e 's|public static String APP_HASH = ".*";|public static String APP_HASH = "014b35b6184100b085b0d0572f9b5103";|' \
  -e 's/public static boolean DEBUG_VERSION = .*/public static boolean DEBUG_VERSION = true;/' \
  -e 's/public static boolean LOGS_ENABLED = .*/public static boolean LOGS_ENABLED = true;/' \
  -e 's/public static boolean CHECK_UPDATES = .*/public static boolean CHECK_UPDATES = false;/' \
  -e 's/public static boolean USE_CLOUD_STRINGS = .*/public static boolean USE_CLOUD_STRINGS = false;/' \
  "$BUILDVARS"

# Patch ConnectionsManager.cpp - robust approach using line-based replacement
CPP_FILE="TMessagesProj/jni/tgnet/ConnectionsManager.cpp"

cat > /tmp/fixdc.py << 'EOF'
import sys
f = sys.argv[1]
ip = sys.argv[2]
p = sys.argv[3]

with open(f) as fh:
    content = fh.read()

s = content.find("void ConnectionsManager::initDatacenters() {")
e = content.find("void ConnectionsManager::attachConnection", s)
if s < 0 or e < 0:
    print("ERROR: markers not found")
    sys.exit(1)

# Find last } before attachConnection
cut = content[s:e]
lb = cut.rfind("}")
if lb < 0:
    print("ERROR: no closing brace")
    sys.exit(1)

# Build replacement without any triple-quote issues
lines = []
lines.append('void ConnectionsManager::initDatacenters() {')
lines.append('    Datacenter *datacenter;')
lines.append('    for (int dcId = 1; dcId <= 5; dcId++) {')
lines.append('        if (datacenters.find(dcId) == datacenters.end()) {')
lines.append('            datacenter = new Datacenter(instanceNum, dcId);')
lines.append("            datacenter->addAddressAndPort(\"" + ip + "\", " + p + ", 0, \"\");")
lines.append('            datacenters[dcId] = datacenter;')
lines.append('        }')
lines.append('    }')
lines.append('    if (datacenters.find(-1) == datacenters.end()) {')
lines.append('        datacenter = new Datacenter(instanceNum, -1);')
lines.append("        datacenter->addAddressAndPort(\"" + ip + "\", " + p + ", 0, \"\");")
lines.append('        datacenters[-1] = datacenter;')
lines.append('    }')
lines.append('}')

new_func = '\n'.join(lines)
result = content[:s] + new_func + content[s+lb+1:]

with open(f, 'w') as fh:
    fh.write(result)
print("OK")
EOF

python3 /tmp/fixdc.py "$CPP_FILE" "$SERVER_IP" "$SERVER_PORT"
echo "==> Patch complete! Server: ${SERVER_IP}:${SERVER_PORT}"
