#!/bin/bash

# S3 Manager API Test Script
# This script demonstrates the new user management and audit logging features

BASE_URL="http://localhost:8081"

echo "🚀 S3 Manager API Test Script"
echo "=============================="

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# Register a test user
echo "2. Registering a test user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "email": "test@example.com"
  }')
echo "$REGISTER_RESPONSE" | jq '.'
echo ""

# Login as test user
echo "3. Logging in as test user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')
echo "$LOGIN_RESPONSE" | jq '.'

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
echo "Token: $TOKEN"
echo ""

# Test protected endpoint
echo "4. Testing protected endpoint (configs)..."
curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/configs" | jq '.'
echo ""

# Test change password
echo "5. Testing password change..."
curl -s -X POST "$BASE_URL/api/auth/change-password" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "testpass123",
    "new_password": "newpass123"
  }' | jq '.'
echo ""

echo "✅ Basic API tests completed!"
echo ""
echo "📝 Admin Features Test:"
echo "To test admin features, first create an admin user:"
echo "  ./create-admin -username admin -email admin@example.com"
echo ""
echo "Then login as admin and test these endpoints:"
echo "  GET    /api/admin/users                    - List all users"
echo "  POST   /api/admin/users                    - Create new user"
echo "  PUT    /api/admin/users/:username          - Update user"
echo "  DELETE /api/admin/users/:username          - Delete user"
echo "  GET    /api/admin/users/:username/config   - Get user config"
echo "  GET    /api/admin/audit-logs               - Get audit logs"
echo "  POST   /api/admin/audit-logs/filter        - Filter audit logs"
echo "  GET    /api/admin/audit-logs/incident/:id  - Get logs by incident"
