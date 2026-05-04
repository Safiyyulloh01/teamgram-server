#!/bin/bash
# Teamgram Server - One-command Deployment
# Run this on your Linux PC (Ubuntu/Debian recommended)
# Usage: bash deploy_teamgram.sh

set -e

echo "==> Teamgram Server Deployment"
echo ""

# 1. Install Docker if not present
if ! command -v docker &>/dev/null; then
    echo "==> Installing Docker..."
    curl -fsSL https://get.docker.com | bash
    sudo usermod -aG docker $USER
    echo "  Docker installed. You may need to log out and back in."
fi

if ! command -v docker compose &>/dev/null; then
    echo "==> Installing Docker Compose..."
    sudo apt-get install -y docker-compose-plugin 2>/dev/null || \
    sudo apt-get install -y docker-compose 2>/dev/null
fi

# 2. Clone Teamgram server
echo "==> Cloning Teamgram server..."
cd ~
git clone https://github.com/Safiyyulloh01/teamgram-server.git
cd teamgram-server

# 3. Initialize database
echo "==> Setting up database..."
sudo docker compose -f docker-compose-env.yaml up -d mysql
sleep 10  # Wait for MySQL to be ready

# Run SQL migrations
mysql -h 127.0.0.1 -P 3306 -uroot -proot teamgram < teamgramd/deploy/sql/1_teamgram.sql 2>/dev/null || true
for f in teamgramd/deploy/sql/migrate-*.sql; do
    mysql -h 127.0.0.1 -P 3306 -uroot -proot teamgram < "$f" 2>/dev/null || true
done
mysql -h 127.0.0.1 -P 3306 -uroot -proot teamgram < teamgramd/deploy/sql/migrate_channels.sql 2>/dev/null || true

# 4. Start all infrastructure
echo "==> Starting infrastructure (MySQL, Redis, etcd, Kafka, MinIO)..."
sudo docker compose -f docker-compose-env.yaml up -d
sleep 15

# 5. Configure MinIO buckets via minio-mc
echo "==> Creating MinIO buckets..."
sudo docker run --rm --network teamgram_net \
  -e MINIO_ROOT_USER=minio \
  -e MINIO_ROOT_PASSWORD=miniostorage \
  minio/mc:latest \
  /bin/sh -c '
    sleep 5
    mc alias set myminio http://minio:9000 minio miniostorage
    mc mb myminio/documents --ignore-existing
    mc mb myminio/encryptedfiles --ignore-existing
    mc mb myminio/photos --ignore-existing
    mc mb myminio/videos --ignore-existing
  ' 2>/dev/null || true

# 6. Start Teamgram services
echo "==> Starting Teamgram services..."
sudo docker compose up -d
sleep 10

# 7. Check status
echo ""
echo "==> Checking services..."
sudo docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "=== DEPLOYMENT COMPLETE ==="
echo "Server IP: $(hostname -I | awk '{print $1}')"
echo "Server Port: 10443"
echo "Verification code: 12345"
echo ""
echo "To rebuild the Android APK with this server's IP:"
echo "  https://github.com/Safiyyulloh01/telegram-teamgram-builder/actions"
echo ""
echo "To check logs: sudo docker compose logs -f"
echo "To stop: sudo docker compose down"
