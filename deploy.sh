#!/bin/bash

# Exit on any error
set -e

echo "🚀 Starting deployment to minikube..."

# Check if minikube is running
if ! minikube status | grep -q "Running"; then
    echo "Starting minikube..."
    minikube start
fi

# Set docker environment to use minikube's docker daemon
eval $(minikube docker-env)

echo "🔨 Building Docker image..."
docker build -t ecommerce-app:latest .

echo "📦 Applying Kubernetes manifests..."

# Apply namespace
kubectl apply -f deployments/namespace.yaml

# Apply PostgreSQL
kubectl apply -f deployments/postgres-configmap.yaml
kubectl apply -f deployments/postgres-deployment.yaml

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/postgres -n ecommerce-app

# Apply application
kubectl apply -f deployments/app-deployment.yaml

# Wait for application to be ready
echo "⏳ Waiting for application to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/ecommerce-app -n ecommerce-app

echo "✅ Deployment completed!"

# Get the service URL
echo "🌐 Getting service URL..."
minikube service ecommerce-app -n ecommerce-app --url

echo "📊 Checking deployment status..."
kubectl get pods -n ecommerce-app

echo "🎉 Deployment successful! Your e-commerce app is now running on minikube."
echo "To access the application, run: minikube service ecommerce-app -n ecommerce-app"
