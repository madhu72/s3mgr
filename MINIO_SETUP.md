# MinIO Setup Guide

This guide will help you set up MinIO for use with the S3 Manager application.

## What is MinIO?

MinIO is a high-performance, S3-compatible object storage system. It's perfect for:
- Self-hosted storage solutions
- Development and testing
- Private cloud storage
- Edge computing scenarios

## Installation Options

### Option 1: Docker (Recommended)

#### Single Node Setup
```bash
# Create data directory
mkdir -p ~/minio/data

# Run MinIO server
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -v ~/minio/data:/data \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin123" \
  minio/minio server /data --console-address ":9001"
```

#### Docker Compose Setup
Create `docker-compose.minio.yml`:
```yaml
version: '3.8'

services:
  minio:
    image: minio/minio
    container_name: minio
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

volumes:
  minio_data:
```

Start with:
```bash
docker-compose -f docker-compose.minio.yml up -d
```

### Option 2: Binary Installation

#### Linux/macOS
```bash
# Download MinIO binary
wget https://dl.min.io/server/minio/release/linux-amd64/minio
chmod +x minio

# Create data directory
mkdir -p ~/minio-data

# Start MinIO server
./minio server ~/minio-data --console-address ":9001"
```

#### Windows
```powershell
# Download and run
Invoke-WebRequest -Uri "https://dl.min.io/server/minio/release/windows-amd64/minio.exe" -OutFile "minio.exe"
.\minio.exe server C:\minio-data --console-address ":9001"
```

### Option 3: Package Managers

#### Homebrew (macOS)
```bash
brew install minio/stable/minio
minio server ~/minio-data --console-address ":9001"
```

#### Chocolatey (Windows)
```powershell
choco install minio
minio server C:\minio-data --console-address ":9001"
```

## Initial Configuration

### 1. Access MinIO Console
- Open your browser and go to: http://localhost:9001
- Default credentials:
  - Username: `minioadmin`
  - Password: `minioadmin123` (or `minioadmin` if you used the simple setup)

### 2. Create Access Keys
1. In the MinIO Console, go to **Access Keys**
2. Click **Create Access Key**
3. Note down the Access Key and Secret Key
4. Optionally set an expiration date

### 3. Create a Bucket
1. Go to **Buckets** in the MinIO Console
2. Click **Create Bucket**
3. Enter a bucket name (e.g., `my-files`)
4. Click **Create Bucket**

### 4. Configure Bucket Policy (Optional)
For public read access:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::my-files/*"
    }
  ]
}
```

## S3 Manager Configuration

### In the S3 Manager Application:
1. Open the S3 Manager web interface
2. Go to **Settings** tab
3. Select **MinIO** as storage type
4. Enter the following configuration:
   - **Access Key ID**: Your MinIO access key
   - **Secret Access Key**: Your MinIO secret key
   - **Endpoint URL**: `http://localhost:9000` (or your MinIO server URL)
   - **Region**: `us-east-1` (or leave empty)
   - **Bucket Name**: Your bucket name
   - **Use SSL**: Unchecked for local development (checked for HTTPS)

## Production Deployment

### SSL/TLS Setup
For production, enable HTTPS:

1. **Generate certificates**:
   ```bash
   # Self-signed certificate (development)
   openssl req -new -x509 -days 365 -nodes -out server.crt -keyout server.key
   
   # Or use Let's Encrypt for production
   certbot certonly --standalone -d your-domain.com
   ```

2. **Configure MinIO with TLS**:
   ```bash
   # Create certs directory
   mkdir -p ~/.minio/certs
   
   # Copy certificates
   cp server.crt ~/.minio/certs/public.crt
   cp server.key ~/.minio/certs/private.key
   
   # Start MinIO (will automatically use HTTPS)
   minio server ~/minio-data --console-address ":9001"
   ```

### Reverse Proxy Setup (Nginx)
```nginx
server {
    listen 80;
    server_name minio.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name minio-console.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:9001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Environment Variables

Common MinIO environment variables:
```bash
# Authentication
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=password123

# Console
MINIO_BROWSER_REDIRECT_URL=http://localhost:9001

# Region
MINIO_REGION_NAME=us-east-1

# Logging
MINIO_LOG_LEVEL=INFO

# Prometheus metrics
MINIO_PROMETHEUS_AUTH_TYPE=public
```

## Monitoring and Maintenance

### Health Check
```bash
curl -f http://localhost:9000/minio/health/live
```

### Backup
```bash
# Using mc (MinIO Client)
mc mirror minio/my-bucket /backup/location
```

### Logs
```bash
# Docker logs
docker logs minio

# Binary logs are printed to stdout
```

## Troubleshooting

### Common Issues

1. **Port already in use**:
   ```bash
   # Check what's using the port
   lsof -i :9000
   lsof -i :9001
   
   # Use different ports
   docker run -p 9002:9000 -p 9003:9001 minio/minio ...
   ```

2. **Permission denied**:
   ```bash
   # Fix data directory permissions
   sudo chown -R $(whoami) ~/minio-data
   chmod 755 ~/minio-data
   ```

3. **Browser access issues**:
   - Check firewall settings
   - Ensure console address is set: `--console-address ":9001"`
   - Try accessing via IP instead of localhost

4. **S3 Manager connection issues**:
   - Verify endpoint URL format: `http://localhost:9000`
   - Check access keys are correct
   - Ensure bucket exists
   - Verify SSL settings match (HTTP vs HTTPS)

### Testing Connection
```bash
# Using curl
curl -X GET http://localhost:9000

# Using aws cli with MinIO
aws --endpoint-url http://localhost:9000 s3 ls
```

## MinIO Client (mc)

Install and configure the MinIO client:
```bash
# Install mc
curl https://dl.min.io/client/mc/release/linux-amd64/mc \
  --create-dirs \
  -o $HOME/minio-binaries/mc
chmod +x $HOME/minio-binaries/mc

# Add to PATH
export PATH=$PATH:$HOME/minio-binaries/

# Configure alias
mc alias set myminio http://localhost:9000 minioadmin minioadmin123

# Test
mc ls myminio
```

## Resources

- [MinIO Documentation](https://docs.min.io/)
- [MinIO Docker Hub](https://hub.docker.com/r/minio/minio)
- [MinIO Client Documentation](https://docs.min.io/docs/minio-client-complete-guide.html)
- [S3 Compatibility](https://docs.min.io/docs/minio-server-limits-per-tenant.html)
