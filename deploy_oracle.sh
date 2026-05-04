#!/bin/bash
# Teamgram Server — Oracle 1GB VPS Lightweight Deployment
# Run this on your Oracle VPS via SSH or Cloud Shell
# Usage: bash deploy_oracle.sh

set -e

echo "==> Teamgram Lightweight Deployment (1GB VPS)"

# ---- 1. Install Docker if missing ----
if ! command -v docker &>/dev/null; then
    echo "==> Installing Docker..."
    curl -fsSL https://get.docker.com | bash
    sudo usermod -aG docker $USER
    echo "  Docker installed. Relog or use 'sudo' for the next steps."
    exit 0
fi

# ---- 2. Clone/fetch repo ----
if [ ! -d ~/teamgram-server ]; then
    echo "==> Cloning Teamgram server..."
    cd ~
    git clone https://github.com/Safiyyulloh01/teamgram-server.git
fi
cd ~/teamgram-server

if [ ! -f docker-compose-env-light.yaml ]; then
    echo "==> Downloading lightweight compose files..."
    # Download from raw GitHub URLs (adjust owner/repo as needed)
    # Or copy manually from the patches directory
    echo "Place docker-compose-env-light.yaml and docker-compose-light.yaml in this directory."
    echo "Download from:"
    echo "  https://raw.githubusercontent.com/Safiyyulloh01/teamgram-server/main/patches/docker-compose-env-light.yaml"
    echo "  https://raw.githubusercontent.com/Safiyyulloh01/teamgram-server/main/patches/docker-compose-light.yaml"
    echo ""
    echo "Or copy them now from your local machine:"
    echo "  scp docker-compose-env-light.yaml ubuntu@<IP>:~/teamgram-server/"
    echo "  scp docker-compose-light.yaml ubuntu@<IP>:~/teamgram-server/"
    echo ""
    read -p "Press Enter after placing the files, or Ctrl+C to abort..."
fi

# ---- 3. Create data dirs ----
mkdir -p data/mysql data/redis data/etcd data/minio data/kafka

# ---- 4. Create .env if missing ----
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

# ---- 5. Ensure SWAP is on (critical for 1GB) ----
if ! swapon --show | grep -q .; then
    echo "==> Creating 2GB swap file..."
    sudo fallocate -l 2G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
fi

# ---- 6. Start infrastructure ----
echo "==> Starting infrastructure (MySQL, Redis, etcd, Kafka, MinIO)..."
sudo docker compose -f docker-compose-env-light.yaml up -d
echo "==> Waiting for MySQL to be ready..."
sleep 40

# ---- 7. Run SQL migrations ----
echo "==> Running SQL migrations..."
# MySQL might take a while; wait for it
for i in $(seq 1 30); do
    if sudo docker exec mysql mysqladmin ping -h localhost -uroot -proot --silent 2>/dev/null; then
        echo "  MySQL ready after ${i}s"
        break
    fi
    sleep 2
done

sudo docker exec -i mysql mysql -uroot -proot teamgram < teamgramd/deploy/sql/1_teamgram.sql 2>/dev/null || true
for f in teamgramd/deploy/sql/migrate-*.sql; do
    [ -f "$f" ] && sudo docker exec -i mysql mysql -uroot -proot teamgram < "$f" 2>/dev/null || true
done
if [ -f teamgramd/deploy/sql/migrate_channels.sql ]; then
    sudo docker exec -i mysql mysql -uroot -proot teamgram < teamgramd/deploy/sql/migrate_channels.sql 2>/dev/null || true
fi

echo "  SQL migrations done."

# ---- 8. Build and start Teamgram ----
echo "==> Building Teamgram server (this may take 5-15 mins on 1GB VPS)..."
# Use limited parallelism to avoid OOM during build
sudo DOCKER_BUILDKIT=1 docker build --memory=512m --build-arg GOMEMLIMIT=384MiB -t teamgram-server .
echo "==> Build complete. Starting Teamgram..."
sudo docker compose -f docker-compose-light.yaml up -d
sleep 10

# ---- 9. Check status ----
echo ""
echo "==> Running containers:"
sudo docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "=== DEPLOYMENT COMPLETE ==="
echo "Server IP: $(hostname -I | awk '{print $1}')"
echo "Server Port: 10443"
echo "Verification code: 12345"
echo ""
echo "To rebuild APK with this IP:"
echo "  https://github.com/Safiyyulloh01/telegram-teamgram-builder/actions"
echo ""
echo "To check logs:  sudo docker logs teamgram-server --tail 50 -f"
echo "To stop:        sudo docker compose -f docker-compose-env-light.yaml down && sudo docker compose -f docker-compose-light.yaml down"
echo "To restart:     sudo docker compose -f docker-compose-light.yaml restart"
echo ""
echo "Memory tips:"
echo "  Check usage: sudo docker stats --no-stream"
echo "  If OOM:      sudo docker compose -f docker-compose-env-light.yaml restart mysql"
