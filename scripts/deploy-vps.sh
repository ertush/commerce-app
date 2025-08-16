#!/bin/bash

# VPS Deployment Script for Commerce App
set -e

# Configuration
IMAGE_NAME=${IMAGE_NAME:-"ecommerce-app"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
ENVIRONMENT=${ENVIRONMENT:-"staging"}
NAMESPACE="ecommerce-app-${ENVIRONMENT}"

echo "🔧 Environment: ${ENVIRONMENT}"
echo "🏷️  Image: ${IMAGE_NAME}:${IMAGE_TAG}"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
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

    # Set docker environment to use minikube's docker daemon
    eval $(minikube docker-env)

    # Build the image
    docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .
    docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest

    echo "[+] Image Name: ${IMAGE_NAME}:${IMAGE_TAG}"
    echo "[+] Image Tag: ${IMAGE_TAG}"
    echo "[+] Image Repository: ${IMAGE_NAME}"

    sed -i "s|image: .*|image: ${IMAGE_NAME}:${IMAGE_TAG}|g" deployments/${ENVIRONMENT}/app-deployment.yaml

    cat deployments/${ENVIRONMENT}/app-deployment.yaml

    echo "✅ Image built successfully"
}

# Setup environment-specific configuration
setup_environment() {
    echo "⚙️  Setting up ${ENVIRONMENT} environment..."

    # Create environment-specific directory
    mkdir -p deployments/${ENVIRONMENT}

    # Copy base deployments
    cp deployments/*.yaml deployments/${ENVIRONMENT}/

    # Update namespace in all files
    # sed -i "s/ecommerce-app/${NAMESPACE}/g" deployments/${ENVIRONMENT}/*.yaml

    # Update image tag in app deployment
    # sed -i "s/${IMAGE_NAME}:latest/${IMAGE_NAME}:${IMAGE_TAG}/g" deployments/${ENVIRONMENT}/app-deployment.yaml

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
       sleep 2



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
       sleep 5

       # Wait for PostgreSQL
       wait_for_service postgres ${NAMESPACE}

       # Apply application deployment
       echo "Deploying application..."
       kubectl apply -f deployments/${ENVIRONMENT}/app-deployment.yaml

       # Wait a moment for deployment to be created
       sleep 5

       # Wait for application
       wait_for_service ecommerce-app ${NAMESPACE}

       echo "✅ Deployment completed!"
}

# Get service information
get_service_info() {
    echo "🌐 Getting service information..."

    # Get minikube IP
    MINIKUBE_IP=$(minikube ip)

    # Get NodePort
    NODE_PORT=$(kubectl get service ecommerce-app -n ${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}')

    echo "🎉 Application deployed successfully!"
    echo "📍 Access your application at: http://${MINIKUBE_IP}:${NODE_PORT}"
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
    kubectl get events -n ${NAMESPACE} --sort-by=.metadata.creationTimestamp --tail=10
}

# Cleanup old deployments
cleanup_old_deployments() {
    echo "🧹 Cleaning up old resources..."

    # Remove old unused images (keep last 3)
    docker images ${IMAGE_NAME} --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}" | tail -n +4 | awk '{print $3}' | head -n -3 | xargs -r docker rmi || true

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
    # build_image

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

    echo "🎉 Deployment completed successfully!"
}

# Error handling
trap 'echo "❌ Deployment failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
