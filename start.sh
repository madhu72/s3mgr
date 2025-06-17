#!/bin/bash

# S3 Manager Startup Script

echo "🚀 Starting S3 Manager Application"
echo "=================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 16 or higher."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Install frontend dependencies if node_modules doesn't exist
if [ ! -d "frontend/node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    cd frontend && npm install && cd ..
fi

# Check if port 8081 is available
if lsof -Pi :8081 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Port 8081 is already in use. Trying port 8082..."
    export PORT=8082
    BACKEND_PORT=8082
else
    export PORT=8081
    BACKEND_PORT=8081
fi

# Start backend in background
echo "🔧 Starting Go backend server on port $BACKEND_PORT..."
go run . &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 3

# Check if backend started successfully
if ! ps -p $BACKEND_PID > /dev/null; then
    echo "❌ Failed to start backend server"
    exit 1
fi

# Start frontend
echo "🎨 Starting React frontend..."
cd frontend && npm run dev &
FRONTEND_PID=$!

echo ""
echo "🎉 Application started successfully!"
echo "📱 Frontend: http://localhost:5173"
echo "🔧 Backend:  http://localhost:$BACKEND_PORT"
echo ""
echo "📋 Quick Start Guide:"
echo "1. Open http://localhost:5173 in your browser"
echo "2. Register a new account or login"
echo "3. Configure your AWS S3 credentials in Settings"
echo "4. Start uploading and managing files!"
echo ""
echo "Press Ctrl+C to stop both servers"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping servers..."
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    echo "✅ Servers stopped successfully"
    exit 0
}

# Trap Ctrl+C
trap cleanup INT

# Wait for processes
wait
