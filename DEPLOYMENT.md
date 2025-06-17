# S3 Manager Deployment Guide

This guide covers different deployment options for the S3 Manager application.

## Quick Start (Development)

### Using the Start Script
```bash
./start.sh
```

### Manual Start
```bash
# Terminal 1 - Backend
PORT=8081 go run .

# Terminal 2 - Frontend
cd frontend && npm run dev
```

## Production Deployment

### Option 1: Docker Compose (Recommended)

1. **Build and start services:**
   ```bash
   docker-compose up -d
   ```

2. **View logs:**
   ```bash
   docker-compose logs -f
   ```

3. **Stop services:**
   ```bash
   docker-compose down
   ```

### Option 2: Manual Production Build

#### Backend
```bash
# Build the Go binary
go build -o s3mgr

# Set environment variables
export PORT=8081
export JWT_SECRET=your-production-secret
export GIN_MODE=release

# Run the binary
./s3mgr
```

#### Frontend
```bash
cd frontend

# Build for production
npm run build

# Serve with any static file server
# Example with serve:
npx serve -s dist -l 5173
```

### Option 3: Cloud Deployment

#### AWS EC2 Deployment

1. **Launch EC2 instance** (Ubuntu 20.04 LTS recommended)

2. **Install dependencies:**
   ```bash
   # Update system
   sudo apt update && sudo apt upgrade -y
   
   # Install Go
   wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   
   # Install Node.js
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
   sudo apt-get install -y nodejs
   
   # Install nginx
   sudo apt install nginx -y
   ```

3. **Deploy application:**
   ```bash
   # Clone/upload your code
   git clone <your-repo> s3mgr
   cd s3mgr
   
   # Build backend
   go build -o s3mgr
   
   # Build frontend
   cd frontend
   npm install
   npm run build
   ```

4. **Configure nginx:**
   ```bash
   sudo cp frontend/nginx.conf /etc/nginx/sites-available/s3mgr
   sudo ln -s /etc/nginx/sites-available/s3mgr /etc/nginx/sites-enabled/
   sudo rm /etc/nginx/sites-enabled/default
   sudo systemctl restart nginx
   ```

5. **Create systemd service:**
   ```bash
   sudo tee /etc/systemd/system/s3mgr.service > /dev/null <<EOF
   [Unit]
   Description=S3 Manager Backend
   After=network.target
   
   [Service]
   Type=simple
   User=ubuntu
   WorkingDirectory=/home/ubuntu/s3mgr
   ExecStart=/home/ubuntu/s3mgr/s3mgr
   Environment=PORT=8081
   Environment=JWT_SECRET=your-production-secret
   Environment=GIN_MODE=release
   Restart=always
   
   [Install]
   WantedBy=multi-user.target
   EOF
   
   sudo systemctl enable s3mgr
   sudo systemctl start s3mgr
   ```

#### Heroku Deployment

1. **Create Heroku apps:**
   ```bash
   # Backend
   heroku create your-s3mgr-backend
   
   # Frontend
   heroku create your-s3mgr-frontend
   ```

2. **Configure backend:**
   ```bash
   # Set environment variables
   heroku config:set JWT_SECRET=your-production-secret -a your-s3mgr-backend
   heroku config:set GIN_MODE=release -a your-s3mgr-backend
   
   # Deploy
   git subtree push --prefix=. heroku main
   ```

3. **Configure frontend:**
   ```bash
   # Update API URL in frontend/src/services/api.js
   # Set to your Heroku backend URL
   
   # Deploy
   cd frontend
   git init
   heroku git:remote -a your-s3mgr-frontend
   git add .
   git commit -m "Initial commit"
   git push heroku main
   ```

## Environment Variables

### Backend
- `PORT`: Server port (default: 8081)
- `JWT_SECRET`: JWT signing secret (required in production)
- `GIN_MODE`: Gin mode (debug/release)

### Frontend
- Update `API_BASE_URL` in `src/services/api.js` for production

## SSL/HTTPS Setup

### Using Let's Encrypt (Recommended)

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Using Cloudflare

1. Set up Cloudflare for your domain
2. Enable "Full (strict)" SSL/TLS encryption
3. Configure origin certificates if needed

## Monitoring and Logs

### Application Logs
```bash
# Backend logs
sudo journalctl -u s3mgr -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Health Checks
```bash
# Backend health
curl http://localhost:8081/api/files

# Frontend health
curl http://localhost/
```

## Security Considerations

1. **Use strong JWT secrets** in production
2. **Enable HTTPS** for all traffic
3. **Configure firewall** to only allow necessary ports
4. **Regular updates** of dependencies
5. **Monitor logs** for suspicious activity
6. **Backup database** regularly (BadgerDB data directory)

## Backup and Recovery

### Database Backup
```bash
# Create backup
tar -czf s3mgr-backup-$(date +%Y%m%d).tar.gz data/

# Restore backup
tar -xzf s3mgr-backup-YYYYMMDD.tar.gz
```

### Configuration Backup
```bash
# Backup environment and configs
cp .env .env.backup
cp -r frontend/dist frontend/dist.backup
```

## Troubleshooting

### Common Issues

1. **Port already in use:**
   ```bash
   sudo lsof -i :8081
   sudo kill -9 <PID>
   ```

2. **Permission denied:**
   ```bash
   sudo chown -R $USER:$USER data/
   chmod 755 s3mgr
   ```

3. **CORS errors:**
   - Check frontend URL in backend CORS config
   - Verify API_BASE_URL in frontend

4. **Database issues:**
   ```bash
   # Reset database
   rm -rf data/
   # Restart application
   ```

For more help, check the logs and ensure all prerequisites are met.
