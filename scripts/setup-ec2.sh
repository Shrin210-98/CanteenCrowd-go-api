#!/bin/bash

# Run this ONCE on your EC2 instance after creation
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

# Install Docker Compose plugin (modern way)
if ! docker compose version &> /dev/null; then
    echo "🔄 Installing Docker Compose plugin..."
    sudo mkdir -p /usr/local/lib/docker/cli-plugins
    sudo curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/lib/docker/cli-plugins/docker-compose
    sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
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

# Setup swap (important for --build on small instances)
if [ ! -f /swapfile ]; then
    echo "💾 Setting up 2GB swap..."
    sudo fallocate -l 2G /swapfile || sudo dd if=/dev/zero of=/swapfile bs=1M count=2048
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
fi

echo "✅ Setup complete!"
echo "⚠️  Run 'exit' and SSH back in for docker group permissions to take effect"
echo "🚀 Then push to main to trigger your first deployment!"

# ssh -i ~/.ssh/go-api-key.pem ec2-user@<EC2_PUBLIC_IP>
# curl -fsSL https://raw.githubusercontent.com/Shrin210-98/CanteenCrowd-go-api/main/scripts/setup-ec2.sh | bash
# exit
# # SSH back in again
# ssh -i ~/.ssh/go-api-key.pem ec2-user@<EC2_PUBLIC_IP>
# docker --version  # verify ✅
