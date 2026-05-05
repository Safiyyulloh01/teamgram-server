#!/bin/bash
# Teamgram Server — Oracle 1GB VPS Lightweight Deployment
# Run: bash <(curl -sL https://raw.githubusercontent.com/Safiyyulloh01/teamgram-server/main/deploy_oracle.sh)

set -e

REPO_URL="https://github.com/Safiyyulloh01/teamgram-server.git"
RAW_URL="https://raw.githubusercontent.com/Safiyyulloh01/teamgram-server/main"

echo "=============================================="
echo "  Teamgram Lightweight Deployment (1GB VPS)"
echo "=============================================="

# ---- 1. Install Docker if missing ----
if ! command -v docker &>/dev/null; then
    echo "[1/8] Installing Docker..."
    curl -fsSL https://get.docker.com | sudo bash
    sudo usermod -aG docker $USER
    echo "  Docker installed. Log out and back in, then re-run."
    exit 0
fi

# ---- 2. Ensure SWAP (critical for 1GB) ----
echo "[2/8] Ensuring swap..."
if ! swapon --show | grep -q .; then
    sudo fallocate -l 2G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    echo "  Created 2GB swap."
else
    echo "  Swap already active."
fi

# ---- 3. Clone repo ----
echo "[3/8] Cloning Teamgram server..."
if [ ! -d ~/teamgram-server ]; then
    cd ~
    git clone "$REPO_URL"
fi
cd ~/teamgram-server

# ---- 4. Download light compose files ----
echo "[4/8] Ensuring lightweight compose files..."
for f in docker-compose-env-light.yaml docker-compose-light.yaml; do
    if [ ! -f "$f" ]; then
        wget -q "$RAW_URL/$f" -O "$f"
        echo "  Downloaded $f"
    fi
done

# ---- 5. Create data dirs & .env ----
echo "[5/8] Preparing config..."
mkdir -p data/mysql data/redis data/etcd data/minio
if [ ! -f .env ]; then
    cat > .env << 'ENVEOF'
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=teamgram
MYSQL_USER=teamgram
MYSQL_PASSWORD=teamgram
MINIO_ROOT_USER=minio
MINIO_ROOT_PASSWORD=miniostorage
ENVEOF
fi

# ---- 6. Start infrastructure ----
echo "[6/8] Starting infrastructure (MySQL, Redis, etcd, Kafka, MinIO)..."
sudo docker compose -f docker-compose-env-light.yaml up -d
echo "  Waiting for MySQL..."
for i in $(seq 1 30); do
    if sudo docker exec mysql mysqladmin ping -h localhost -uroot -proot --silent 2>/dev/null; then
        echo "  MySQL ready after ${i}s"
        break
    fi
    sleep 2
done

# ---- 7. SQL migrations ----
echo "[7/8] Running SQL migrations..."
sudo docker exec -i mysql mysql -uroot -proot teamgram < teamgramd/deploy/sql/1_teamgram.sql 2>/dev/null || true
for f in teamgramd/deploy/sql/migrate-*.sql; do
    [ -f "$f" ] && sudo docker exec -i mysql mysql -uroot -proot teamgram < "$f" 2>/dev/null || true
done
[ -f teamgramd/deploy/sql/migrate_channels.sql ] && sudo docker exec -i mysql mysql -uroot -proot teamgram < teamgramd/deploy/sql/migrate_channels.sql 2>/dev/null || true
echo "  SQL migrations done."

# ---- 8. Pull Teamgram image from GHCR ----
echo "[8/8] Pulling Teamgram image from GHCR..."
sudo docker pull ghcr.io/safiyyulloh01/teamgram-server:latest
sudo docker tag ghcr.io/safiyyulloh01/teamgram-server:latest teamgram-server
echo "  Starting Teamgram..."
sudo docker compose -f docker-compose-light.yaml up -d
sleep 5

# ---- Summary ----
echo ""
echo "=============================================="
echo "  DEPLOYMENT COMPLETE"
echo "=============================================="
echo ""
sudo docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "  Server IP:    $(hostname -I | awk '{print $1}')"
echo "  Server Port:  10443"
echo "  Login code:   12345"
echo ""
echo "Rebuild APK with this IP:"
echo "  https://github.com/Safiyyulloh01/telegram-teamgram-builder/actions"
echo ""
echo "Commands:"
echo "  Logs:   sudo docker compose -f docker-compose-light.yaml logs -f"
echo "  Stats:  sudo docker stats --no-stream"
echo "  Stop:   sudo docker compose -f docker-compose-env-light.yaml down"
echo "          sudo docker compose -f docker-compose-light.yaml down"
