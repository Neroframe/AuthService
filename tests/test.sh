#!/usr/bin/env bash
set -euo pipefail

GRPC_ADDR="localhost:50051"
PROTO_PKG="auth.AuthService"

# Test credentials
ADMIN_EMAIL="26@admin.com"
ADMIN_PASS="Admin123!"
STUD_EMAIL="26@student.com"
STUD_PASS="Stud123!"

echo "Register Admin"
grpcurl -plaintext \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\",\"role\":1}" \
  $GRPC_ADDR $PROTO_PKG/Register | jq

echo "Login as Admin (Get admin token)"
ADMIN_TOKEN=$(grpcurl -plaintext \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}" \
  $GRPC_ADDR $PROTO_PKG/Login \
  | jq -r '.accessToken')
echo "     -> ADMIN_TOKEN=$ADMIN_TOKEN"

echo "Validate Admin Token"
ADMIN_ID=$(grpcurl -plaintext \
  -d "{\"jwt\":\"$ADMIN_TOKEN\"}" \
  $GRPC_ADDR $PROTO_PKG/ValidateToken \
  | jq -r '.userId')
echo "      -> ADMIN_ID=$ADMIN_ID"

echo "Register Student"
grpcurl -plaintext \
  -d "{\"email\":\"$STUD_EMAIL\",\"password\":\"$STUD_PASS\",\"role\":3}" \
  $GRPC_ADDR $PROTO_PKG/Register | jq

echo "Login as Student (Get student token)"
STUD_TOKEN=$(grpcurl -plaintext \
  -d "{\"email\":\"$STUD_EMAIL\",\"password\":\"$STUD_PASS\"}" \
  $GRPC_ADDR $PROTO_PKG/Login \
  | jq -r '.accessToken')
echo "     -> STUD_TOKEN=$STUD_TOKEN"

echo "Validate Student Token (Get StudentID)"
STUD_ID=$(grpcurl -plaintext \
  -d "{\"jwt\":\"$STUD_TOKEN\"}" \
  $GRPC_ADDR $PROTO_PKG/ValidateToken \
  | jq -r '.userId')
echo "     -> STUD_ID=$STUD_ID"

echo ""
echo "Admin fetching Student profile (should SUCCEED)"
grpcurl -plaintext \
  -H "authorization: Bearer $ADMIN_TOKEN" \
  -d "{\"user_id\":\"$STUD_ID\"}" \
  $GRPC_ADDR $PROTO_PKG/GetUserByID | jq

echo ""
echo "Student fetching Admin profile (should FAIL)"
grpcurl -plaintext \
  -H "authorization: Bearer $STUD_TOKEN" \
  -d "{\"user_id\":\"$ADMIN_ID\"}" \
  $GRPC_ADDR $PROTO_PKG/GetUserByID | jq


