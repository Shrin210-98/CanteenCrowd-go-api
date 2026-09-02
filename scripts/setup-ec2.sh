#!/bin/bash

# Run this once on your EC2 instance after creation
set -e

echo "🚀 Setting up EC2..."

# Update system
sudo yum update -y || sudo apt update -y

# Install Docker
if ! command -v docker &> /dev/null; then
    echo "🐳 Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    rm get-docker.sh
    sudo systemctl start docker
    sudo systemctl enable docker
fi

# Install Docker Compose
if ! command -v docker compose &> /dev/null; then
    echo "🔄 Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    sudo ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
fi

# Install Git
if ! command -v git &> /dev/null; then
    echo "📥 Installing Git..."
    sudo yum install -y git || sudo apt install -y git
fi

# Add user to docker group
sudo usermod -aG docker $USER

# Create app directory
mkdir -p /home/$USER/app

# Setup swap (important for t2.micro)
if [ ! -f /swapfile ]; then
    echo "💾 Setting up swap..."
    sudo fallocate -l 2G /swapfile || sudo dd if=/dev/zero of=/swapfile bs=1M count=2048
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
fi

echo "✅ Setup complete! Now add GitHub secrets and push to main."

