#!/bin/bash
# Generates an RSA private key for signing JWTs, formatted for an env var.
#
# The base auth scaffold is session-based and does not need this. Use it if you
# add JWT/OAuth token signing — generate a key per environment and never commit
# production keys.
#
# Usage: ./scripts/generate_jwt_key.sh [key_size]   (default 2048)

set -e

KEY_SIZE=${1:-2048}

echo "Generating RSA private key for JWT signing (${KEY_SIZE} bits)..."
echo ""
echo "--- BASE64 FORMAT (recommended for deployment platforms) ---"
echo "JWT_PRIVATE_KEY=$(openssl genrsa "$KEY_SIZE" 2>/dev/null | base64 -w 0)"
echo ""
echo "--- PEM FORMAT ---"
openssl genrsa "$KEY_SIZE" 2>/dev/null
echo ""
echo "⚠️  Development use only — generate fresh keys for production, keep them"
echo "    out of version control, and back them up (losing one invalidates all"
echo "    tokens signed with it)."
