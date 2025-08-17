#!/bin/bash

# VPS Deployment Script for Commerce App
set -e


# Configuration
IMAGE_NAME=${IMAGE_NAME:-"ecommerce-app"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
ENVIRONMENT=${ENVIRONMENT:-"staging"}
NAMESPACE="ecommerce-app-${ENVIRONMENT}"
VPS_DOMAIN=${VPS_DOMAIN:-""}

echo "🔧 Environment: ${ENVIRONMENT}"
echo "🏷️  Image: ${IMAGE_NAME}:${IMAGE_TAG}"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

setup_nginx() {
    localhost=$1
    vps_domain=$2

    echo "🔧 Setting up NGINX..."

    # Install NGINX
    if ! command_exists nginx; then
        echo "[+] Installing NGINX..."
        sudo apt-get update
        sudo apt-get install -y nginx

   fi



    if [-e /etc/nginx/sites-available/ecommerce-app]
    then
        # Delete to create a new config for nginx
        sudo rm /etc/nginx/sites-available/ecommerce-app
        sudo systemctl reload nginx
    fi

    if [-e /etc/nginx/sites-enabled/ecommerce-app]
    then
        # Delete to create a new config for nginx
        sudo rm /etc/nginx/sites-available/ecommerce-app
        sudo systemctl reload nginx
    fi


    # check VPS_DOMAIN
    if [ -z "$vps_domain" ]; then
        echo "Error: VPS_DOMAIN is not set."
        exit 1
    fi

    # check localhost
    if [ -z "$localhost" ]; then
        echo "Error: localhost is not set."
        exit 1
    fi

    echo "[+] Creating NGINX configuration..."

    cat <<EOF | sudo tee /etc/nginx/sites-available/ecommerce-app
    server {
        listen 80;
        server_name $vps_domain;

        location / {
            proxy_pass $localhost;  # Your Minikube NodePort
            proxy_set_header Host \$http_host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
    }

    server {
        listen 443 ssl;
        server_name $vps_domain;

        ssl_certificate /etc/ssl/selfsigned/selfsigned.crt;
        ssl_certificate_key /etc/ssl/selfsigned/selfsigned.key;

        location / {
            proxy_pass $localhost;  # Your Minikube NodePort
            proxy_set_header Host \$http_host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
    }

    # Optional: redirect all HTTP (80) traffic to HTTPS
    server {
        listen 80;
        server_name $vps_domain;
        return 301 https://\$host\$request_uri;
    }

EOF

    if [ ! -e /etc/nginx/sites-enabled/ecommerce-app ]
    then
        # Delete to create a new config for nginx
        sudo ln -s /etc/nginx/sites-available/ecommerce-app /etc/nginx/sites-enabled/
    fi

    # Installing certbot
    sudo apt-get update
    sudo apt-get install -y certbot python3-certbot-nginx

    # Reload nginx
    sudo nginx -t
    echo "✅ Nginx configuration tested successfully!"
    sudo systemctl reload nginx
    echo "✅ Nginx restarted successfully!"

    echo "✅ Certbot setup successful!"

}

# Function to wait for service to be ready
wait_for_service() {
    local service=$1
    local namespace=$2
    local timeout=${3:-300}

    echo "⏳ Waiting for $service to be ready in namespace $namespace..."

    # First check if the deployment exists
    if ! kubectl get deployment $service -n $namespace >/dev/null 2>&1; then
        echo "❌ Deployment $service not found in namespace $namespace"
        echo "📋 Available deployments:"
        kubectl get deployments -n $namespace || echo "No deployments found"
        return 1
    fi

    # Check deployment status
    echo "📊 Current deployment status:"
    kubectl get deployment $service -n $namespace

    # Wait for deployment to be available
    if kubectl wait --for=condition=available --timeout=${timeout}s deployment/$service -n $namespace; then
        echo "✅ $service is ready!"
        return 0
    else
        echo "❌ Timeout waiting for $service"
        echo "📋 Deployment details:"
        kubectl describe deployment $service -n $namespace
        echo "📋 Pod status:"
        kubectl get pods -l app=$service -n $namespace
        echo "📋 Recent logs:"
        kubectl logs -l app=$service -n $namespace --tail=50 || echo "No logs available"
        return 1
    fi
}

check_namespace() {
    local namespace=$1
    echo "🔍 Checking namespace: $namespace"

    if kubectl get namespace $namespace >/dev/null 2>&1; then
        echo "✅ Namespace $namespace exists"
        return 0
    else
        echo "❌ Namespace $namespace does not exist"
        return 1
    fi
}


# Install dependencies if needed
install_dependencies() {
    echo "🔧 Checking dependencies..."

    # Install Docker
    if ! command_exists docker; then
        echo "Installing Docker..."
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker $USER && newgrp docker
        sudo systemctl start docker
        sudo systemctl enable docker
    fi

    # Install kubectl
    if ! command_exists kubectl; then
        echo "Installing kubectl..."
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    fi

    # Install minikube
    if ! command_exists minikube; then
        echo "Installing minikube..."
        curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
        chmod +x minikube-linux-amd64
        sudo mv minikube-linux-amd64 /usr/local/bin/minikube

    fi
}

# Configure_firewall
configure_firewall(){
    # Setup firewall
    echo "🔥 Configuring firewall..."
    sudo ufw allow ssh
    sudo ufw allow 80/tcp
    sudo ufw allow 443/tcp
    sudo ufw allow 8080/tcp
    sudo ufw allow 30000:32767/tcp  # NodePort range
    sudo ufw --force enable
}

# Setup minikube
setup_minikube() {
    echo "🎯 Setting up minikube..."

    # Check if minikube is running
    if ! minikube status | grep -q "Running"; then
        echo "Starting minikube..."
        minikube start --driver=docker --memory=2200mb --cpus=2
    else
        echo "Minikube is already running"
    fi

    # Configure kubectl
    minikube update-context

    # Set docker environment
    eval $(minikube docker-env)
}

# Build application image
build_image() {
    echo "🔨 Building Docker image..."

    # stop running process for app
    docker stop $(docker ps -q --filter "name=ecommerce-app") || true

    ecommerce_image_id=$(docker image ls | grep ecommerce-app | tail -n 1 | awk '{print $3}')

    #check previous images and delete if they exist
    if [ $ecommerce_image_id ]
    then
        echo "💡 Deleting previous image..."
        docker image rm -f $ecommerce_image_id | tail -n 1
    fi

    # Build the image
    docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .
    docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${NAMESPACE}:latest

    # Log Image Name and Image Tag
    echo "Image Name: ${IMAGE_NAME}:${IMAGE_TAG}"
    echo "Image Tag: ${IMAGE_TAG}"

    # Check built image
    docker image ls

    echo "✅ Image built successfully"
}

# Setup environment-specific configuration
setup_environment() {
    echo "⚙️ Setting up ${ENVIRONMENT} environment..."

    # Create environment-specific directory
    if [ -d deployments/${ENVIRONMENT}]
    then
        echo "💡 deployments/${ENVIRONMENT} directory exists"
    else
        mkdir -p deployments/${ENVIRONMENT}
    fi

    # Copy base deployments
    cp deployments/*.yaml deployments/${ENVIRONMENT}/

    # Update namespace in all files
    sed -i "s/ecommerce-app/${NAMESPACE}/g" deployments/${ENVIRONMENT}/*.yaml

    # Environment-specific configurations
    case $ENVIRONMENT in
        "production")
            # Production settings
            sed -i 's/replicas: 2/replicas: 3/g' deployments/${ENVIRONMENT}/app-deployment.yaml
            sed -i 's/memory: "256Mi"/memory: "512Mi"/g' deployments/${ENVIRONMENT}/app-deployment.yaml
            sed -i 's/cpu: "200m"/cpu: "500m"/g' deployments/${ENVIRONMENT}/app-deployment.yaml
            ;;
        "staging")
            # Staging settings
            sed -i 's/replicas: 2/replicas: 1/g' deployments/${ENVIRONMENT}/app-deployment.yaml
            ;;
    esac
}

# Deploy to Kubernetes
deploy_to_kubernetes() {
    echo "📦 Deploying to Kubernetes..."

       # Apply namespace first and verify
       echo "Creating namespace..."
       kubectl apply -f deployments/${ENVIRONMENT}/namespace.yaml

       # Wait a moment for namespace to be ready
       sleep 1

       # Verify namespace exists
       if ! check_namespace ${NAMESPACE}; then
           echo "❌ Failed to create namespace"
           return 1
       fi

       # set namesapce
       kubectl config set-context --current --namespace=${NAMESPACE}

       # Apply PostgreSQL configuration
       echo "Deploying PostgreSQL configuration..."
       kubectl apply -f deployments/${ENVIRONMENT}/postgres-configmap.yaml

       # Apply PostgreSQL deployment
       echo "Deploying PostgreSQL..."
       kubectl apply -f deployments/${ENVIRONMENT}/postgres-deployment.yaml

       # Wait a moment for deployment to be created
       sleep 1

       # Wait for PostgreSQL
       wait_for_service postgres ${NAMESPACE}

       # Apply application deployment
       echo "Deploying application..."
       kubectl apply -f deployments/${ENVIRONMENT}/app-deployment.yaml

       # Wait a moment for deployment to be created
       sleep 3

       # Wait for application
       wait_for_service ${NAMESPACE} ${NAMESPACE}

       echo "✅ Deployment completed!"
}

# Get service information
get_service_info() {
    echo "🌐 Getting service information..."

    # Get minikube IP
    export MINIKUBE_IP=$(minikube ip)

    # Get NodePort
    export NODE_PORT=$(kubectl get service ${NAMESPACE} -n ${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}')

    echo "🎉 Application deployed successfully!"
    echo "📍 Access your application at (local IP): http://${MINIKUBE_IP}:${NODE_PORT}/"
    echo "🔍 Health check: http://${MINIKUBE_IP}:${NODE_PORT}/health"

    # Save service info to file
    cat > service-info.txt << EOF
Environment: ${ENVIRONMENT}
Application URL: http://${MINIKUBE_IP}:${NODE_PORT}
Health Check: http://${MINIKUBE_IP}:${NODE_PORT}/health
Namespace: ${NAMESPACE}
Image: ${IMAGE_NAME}:${IMAGE_TAG}
Deployment Time: $(date)
EOF

    echo "📄 Service information saved to service-info.txt"
}

# Show deployment status
show_status() {
    echo "📊 Deployment Status:"
    kubectl get pods -n ${NAMESPACE}
    kubectl get services -n ${NAMESPACE}
    echo ""
    echo "📋 Recent Events:"
    kubectl get events -n ${NAMESPACE} --sort-by=.metadata.creationTimestamp
}

# Cleanup old deployments
cleanup_old_deployments() {
    echo "🧹 Cleaning up old resources..."

    # Clean up docker system
    docker system prune -f || true
}

# Main deployment flow
main() {
    echo "🚀 Starting VPS deployment..."
    echo "Repository: $(pwd)"
    echo "User: $(whoami)"

    # Install dependencies
    install_dependencies

    # Setup minikube
    setup_minikube

    # Build image
    build_image

    # Setup environment
    setup_environment

    # Deploy to Kubernetes
    deploy_to_kubernetes

    # Get service info
    get_service_info

    # Configure firewall
    configure_firewall

    # Show status
    show_status

    # Cleanup
    cleanup_old_deployments

    # Set Up nginx
    setup_nginx http://${MINIKUBE_IP}:${NODE_PORT} ${VPS_DOMAIN}

    echo "🎉 Deployment completed successfully!"
}

# Error handling
trap 'echo "❌ Deployment failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
