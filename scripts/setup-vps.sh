#!/bin/bash

# VPS Initial Setup Script
set -e

echo "🔧 Setting up VPS for Commerce App deployment..."

# Update system
echo "📦 Updating system packages..."
sudo apt-get update
sudo apt-get upgrade -y

# Install essential packages
echo "📦 Installing essential packages..."
sudo apt-get install -y \
    curl \
    wget \
    git \
    unzip \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release

# Install Docker
echo "🐳 Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    sudo systemctl start docker
    sudo systemctl enable docker
    rm get-docker.sh

    sudo usermod -aG docker $USER && newgrp docker
fi

# Install kubectl
echo "☸️  Installing kubectl..."
if ! command -v kubectl &> /dev/null; then
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
fi

# Install minikube
echo "🎯 Installing minikube..."
if ! command -v minikube &> /dev/null; then
    curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    chmod +x minikube-linux-amd64
    sudo mv minikube-linux-amd64 /usr/local/bin/minikube
fi

# Setup firewall
echo "🔥 Configuring firewall..."
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp
sudo ufw allow 30000:32767/tcp  # NodePort range
sudo ufw --force enable

# Create application directory
echo "📁 Creating application directory..."
mkdir -p /home/$USER/commerce-app
cd /home/$USER/commerce-app

# Setup environment variables
echo "⚙️  Setting up environment..."
cat > .env << EOF
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=$(openssl rand -base64 32)
DB_NAME=ecommerce

# Application Configuration
PORT=8080
ENVIRONMENT=staging

# OIDC Configuration (update with your values)
OIDC_PROVIDER_URL=https://accounts.google.com
OIDC_CLIENT_ID=your-google-client-id
OIDC_CLIENT_SECRET=your-google-client-secret
OIDC_REDIRECT_URL=http://$(curl -s ifconfig.me):8080/api/auth/callback
EOF

echo "🔑 Generated environment file with random database password"

# Start minikube
echo "🚀 Starting minikube..."
minikube start --driver=docker --memory=4096 --cpus=2

# Verify installation
echo "✅ Verifying installation..."
docker --version
kubectl version --client
minikube version
minikube status

echo "🎉 VPS setup completed!"
echo ""
echo "Next steps:"
echo "1. Update the .env file with your OIDC configuration"
echo "2. Configure GitHub repository secrets"
echo "3. Push code to trigger deployment"
echo ""
echo "Generated files:"
echo "- .env (update with your configuration)"
echo ""
echo "Important notes:"
echo "- Your application will be accessible on NodePort (30000-32767)"
echo "- Use 'minikube service list' to see all services"
echo "- Logs can be viewed with 'kubectl logs -f deployment/ecommerce-app -n ecommerce-app-production'"
