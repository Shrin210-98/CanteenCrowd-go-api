#!/bin/bash

# Deployment script
set -e

APP_DIR="/home/ec2-user/app"
cd $APP_DIR

echo "🚀 Starting deployment..."

# Pull latest code
if [ -d ".git" ]; then
    echo "📥 Pulling latest changes..."
    git fetch origin main
    git reset --hard origin/main
else
    echo "📥 Cloning repository..."
    git clone https://github.com/Shrin210-98/CanteenCrowd-go-api.git .
fi

# Make scripts executable
chmod +x scripts/*.sh 2>/dev/null || true

# Create .env file
echo "🔐 Setting up environment variables..."
cat > .env << 'EOF'
DB_DATABASE=${DB_DATABASE}
DB_USERNAME=${DB_USERNAME}
DB_PASSWORD=${DB_PASSWORD}
DB_PORT=${DB_PORT}
DB_HOST=${DB_HOST}
JWT_SECRET=${JWT_SECRET}
GIN_MODE=release
EOF

# Deploy
echo "🐳 Deploying containers..."
docker compose -f docker-compose.prod.yml up -d --build

# Clean up
docker image prune -f

# Health check
echo "🔍 Checking deployment status..."
sleep 10
if docker compose -f docker-compose.prod.yml ps | grep -q "unhealthy\|exited"; then
    echo "❌ Deployment failed!"
    docker compose -f docker-compose.prod.yml logs --tail=50
    exit 1
fi

echo "✅ Deployment complete!"
docker compose -f docker-compose.prod.yml ps
