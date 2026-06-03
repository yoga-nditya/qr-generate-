#!/bin/bash
# ============================================================
#  QRIS Payment - Setup & Run Script
#  Linux Ubuntu 10.141.44.238
# ============================================================

set -e

APP_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$APP_DIR"

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║     QRIS Payment Setup & Runner              ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# 1. Cek Go
if ! command -v go &>/dev/null; then
  echo "❌ Go tidak ditemukan. Install dulu:"
  echo "   wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
  echo "   sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz"
  echo "   echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
  echo "   source ~/.bashrc"
  exit 1
fi

echo "✅ Go version: $(go version)"

# 2. Buat direktori data
mkdir -p "$APP_DIR/data"
echo "✅ Direktori data: $APP_DIR/data"

# 3. Cek .env
if [ ! -f "$APP_DIR/.env" ]; then
  echo "⚠️  File .env tidak ditemukan, buat dari template..."
  cp "$APP_DIR/.env.example" "$APP_DIR/.env" 2>/dev/null || true
  echo "   Edit .env dan isi PUBLIC_URL dengan URL ngrok!"
fi

# 4. Download dependencies
echo ""
echo "📦 Download dependencies..."
go mod tidy

echo ""
echo "✅ Dependencies siap"

# 5. Build
echo ""
echo "🔨 Build aplikasi..."
go build -o "$APP_DIR/qris-payment" "$APP_DIR/main.go"
echo "✅ Build berhasil: $APP_DIR/qris-payment"

# 6. Jalankan
echo ""
echo "🚀 Menjalankan server..."
echo "   Buka browser: http://10.141.44.238:3002"
echo "   Webhook URL : \$(PUBLIC_URL)/webhook/midtrans"
echo ""

"$APP_DIR/qris-payment"
