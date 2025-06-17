#!/bin/bash

# MinIO Setup Script for S3 Manager
echo "🚀 Setting up MinIO for S3 Manager..."

# Check if MinIO is installed
if ! command -v minio &> /dev/null; then
    echo "❌ MinIO is not installed. Please install MinIO first:"
    echo "   macOS: brew install minio/stable/minio"
    echo "   Linux: wget https://dl.min.io/server/minio/release/linux-amd64/minio && chmod +x minio"
    exit 1
fi

# Create MinIO data directory
MINIO_DATA_DIR="./minio-data"
if [ ! -d "$MINIO_DATA_DIR" ]; then
    mkdir -p "$MINIO_DATA_DIR"
    echo "✅ Created MinIO data directory: $MINIO_DATA_DIR"
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found. Please make sure you have the .env file in the project root."
    exit 1
fi

echo "✅ MinIO setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Start MinIO server:"
echo "   minio server $MINIO_DATA_DIR --console-address :9001"
echo ""
echo "2. Access MinIO Console at: http://localhost:9001"
echo "   Username: minioadmin"
echo "   Password: minioadmin"
echo ""
echo "3. Start the S3 Manager application:"
echo "   ./start.sh"
echo ""
echo "4. Use the 'Auto MinIO Setup' feature in the Config tab to create your configuration"
echo ""
echo "🔧 Configuration:"
echo "   - MinIO Server: http://localhost:9000"
echo "   - MinIO Console: http://localhost:9001"
echo "   - Admin Credentials: minioadmin/minioadmin"
