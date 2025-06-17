#!/bin/bash

echo "Testing MinIO Auto Configuration API..."

# First, register a test user
echo "1. Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass","email":"test@example.com"}')

echo "Register response: $REGISTER_RESPONSE"

# Login to get token
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}')

echo "Login response: $LOGIN_RESPONSE"

# Extract token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token: $TOKEN"

if [ -z "$TOKEN" ]; then
    echo "Failed to get token, exiting..."
    exit 1
fi

# Test auto-configure MinIO
echo "3. Testing auto-configure MinIO..."
CONFIG_RESPONSE=$(curl -s -X POST http://localhost:8081/api/auto-configure-minio \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"username":"testuser"}')

echo "Config response: $CONFIG_RESPONSE"
