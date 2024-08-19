#!/bin/bash

# Script to extract tls.crt from a Kubernetes secret,
# create a port-forward, test the service over HTTPS,
# and then clean up.

# Exit script on any error
set -e

# Configuration variables
SECRET_NAME="todo-app"
NAMESPACE="todo"
LOCAL_PORT=8443
SERVICE_PORT=443
SERVICE_NAME="todo-app"

# Function to extract the TLS certificate from the K8s secret
extract_certificate() {
  echo "Extracting TLS certificate from secret '${SECRET_NAME}' in namespace '${NAMESPACE}'..."
  kubectl get secret -n ${NAMESPACE} ${SECRET_NAME} -o jsonpath='{.data.tls\.crt}' | base64 -d > tls.crt
}

# Function to clean up background port-forward process
cleanup() {
  echo "Cleaning up port-forward process"
  kill ${PID}
}

# Trap the EXIT signal to call the cleanup function
trap "cleanup" EXIT

# Extract the TLS certificate and save it to tls.crt
extract_certificate

# Create a port-forward for the todo-app service
echo "Creating a port-forward for the ${SERVICE_NAME} service on port ${LOCAL_PORT}..."
kubectl port-forward -n ${NAMESPACE} svc/${SERVICE_NAME} ${LOCAL_PORT}:${SERVICE_PORT} >/dev/null 2>&1 &

# PID of the last job running in the background
PID=$!

# Give the port-forward a second to establish
echo "Waiting for port-forward to establish..."
sleep 1

# Perform the curl test
echo "Testing if the service responds over HTTPS..."
curl -s --cacert tls.crt https://localhost:${LOCAL_PORT}/todo | jq .

# Cleanup will automatically be called when the script exits due to the trap.
