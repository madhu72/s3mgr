# S3 Manager

A full-stack web application for managing AWS S3 and MinIO file operations with user authentication.

## Features

- **Multi-Storage Support**: Works with both AWS S3 and MinIO
- **Auto MinIO Setup**: Automatically create MinIO configurations with dedicated user credentials
- **User Authentication**: Secure JWT-based login and registration
- **File Operations**: Upload, download, delete, and list files
- **Multiple Configurations**: Manage multiple storage configurations per user
- **Credential Management**: Secure storage and rotation of access keys
- **Modern UI**: Responsive React frontend with TailwindCSS
- **Real-time Updates**: Live file listing and upload progress
- **User Management**: Role-based user system with admin capabilities
- **Audit Logging**: Comprehensive logging of all user and admin actions

## Architecture

### Backend (Go)
- **Gin Framework**: Fast HTTP web framework
- **BadgerDB**: Embedded key-value database for user data and configuration
- **AWS SDK**: Official AWS SDK for S3 operations
- **MinIO SDK**: Official MinIO SDK for MinIO operations
- **JWT Authentication**: Secure token-based authentication
- **CORS Support**: Cross-origin resource sharing for frontend integration

### Frontend (React)
- **Vite**: Fast build tool and development server
- **React Router**: Client-side routing
- **TailwindCSS**: Utility-first CSS framework
- **Axios**: HTTP client for API communication
- **Lucide React**: Beautiful icon library

## Prerequisites

- Go 1.21 or higher
- Node.js 16 or higher
- AWS Account with S3 access (optional)
- MinIO server (optional)
- AWS Access Key and Secret Key with S3 permissions (optional)
- MinIO Access Key and Secret Key (optional)

## Installation

### Backend Setup

1. Navigate to the project directory:
   ```bash
   cd s3mgr
   ```

2. Initialize Go modules and install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the backend server:
   ```bash
   go run .
   ```

   The server will start on `http://localhost:8080`

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

   The frontend will be available at `http://localhost:5173`

## Configuration

### Environment Variables

Create a `.env` file in the project root with the following variables:

```bash
# MinIO Admin Configuration (for auto-setup)
MINIO_ADMIN_URL=http://localhost:9000
MINIO_ADMIN_ACCESS_KEY=minioadmin
MINIO_ADMIN_SECRET_KEY=minioadmin

# Default MinIO Configuration for Users
MINIO_DEFAULT_ENDPOINT=localhost:9000
MINIO_DEFAULT_BUCKET=s3manager-default
MINIO_DEFAULT_REGION=us-east-1
MINIO_DEFAULT_SSL=false

# Database Configuration
DB_PATH=./data/s3manager.db

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Server Configuration
PORT=8081
```

### Storage Configuration

#### Auto MinIO Setup (Recommended)

The easiest way to get started with MinIO:

1. **Setup MinIO Server**:
   ```bash
   # Run the setup script
   ./setup-minio.sh
   
   # Start MinIO server
   minio server ./minio-data --console-address :9001
   ```

2. **Auto-Configure in S3 Manager**:
   - Start the S3 Manager application
   - Go to the **Config** tab
   - Use the **Auto MinIO Setup** section
   - Enter your username and click "Create MinIO Configuration"
   - This automatically creates:
     - A dedicated MinIO user with unique credentials
     - A private bucket for your files
     - Proper access policies
     - Saves the configuration to your S3 Manager

#### AWS S3 Setup
1. Create an AWS account and S3 bucket
2. Generate access keys with S3 permissions
3. Configure in the Settings tab

#### MinIO Setup
1. Install and run MinIO server:
   ```bash
   # Using Docker
   docker run -p 9000:9000 -p 9001:9001 \
     -e "MINIO_ROOT_USER=minioadmin" \
     -e "MINIO_ROOT_PASSWORD=minioadmin" \
     minio/minio server /data --console-address ":9001"
   
   # Or download binary
   wget https://dl.min.io/server/minio/release/linux-amd64/minio
   chmod +x minio
   ./minio server /data
   ```

2. Access MinIO Console at http://localhost:9001
3. Create access keys and bucket
4. Configure in the Settings tab with:
   - Storage Type: MinIO
   - Endpoint URL: http://localhost:9000
   - Access Key and Secret Key from MinIO
   - Enable/disable SSL as needed

## Usage

### User Registration and Login

1. Open the application in your browser
2. Click "Sign up" to create a new account
3. Login with your credentials

### Storage Configuration

1. Go to the Settings tab
2. Enter your storage credentials and bucket information
3. Click "Save Configuration"

### File Operations

#### Upload Files
1. Navigate to the Upload tab
2. Select files or drag and drop
3. Click "Upload" to start the process

#### View and Manage Files
1. Go to the Files tab
2. View all files in your storage bucket
3. Download files by clicking the download icon
4. Delete files by clicking the trash icon

#### Key Rotation
1. Go to Settings tab
2. Click on "Rotate Keys"
3. Enter new storage credentials
4. Confirm the rotation

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login

### Storage Operations (Protected)
- `GET /api/files` - List all files
- `POST /api/upload` - Upload file
- `GET /api/download/:key` - Download file
- `DELETE /api/files/:key` - Delete file
- `GET /api/config` - Get storage configuration
- `PUT /api/config` - Update storage configuration
- `POST /api/rotate-keys` - Rotate storage keys

## User Management

### Admin User Creation

Use the command-line tool to create admin users:

```bash
# Interactive mode
go run cmd/create-admin.go -interactive

# Non-interactive mode
go run cmd/create-admin.go -username admin -email admin@example.com -db s3mgr.db
```

### Admin API Endpoints

Admin users have access to additional endpoints:

#### User Management
- `GET /api/admin/users` - List all users
- `POST /api/admin/users` - Create new user
- `PUT /api/admin/users/:username` - Update user details
- `DELETE /api/admin/users/:username` - Delete user
- `GET /api/admin/users/:username/config` - Get user's default configuration

#### Audit Logs
- `GET /api/admin/audit-logs` - Get audit logs with optional filters
- `POST /api/admin/audit-logs/filter` - Advanced filtering of audit logs
- `GET /api/admin/audit-logs/incident/:session_id` - Get logs by incident/session

### Query Parameters for Audit Logs

```
GET /api/admin/audit-logs?user_id=user123&action=login&start_time=2024-01-01T00:00:00Z&limit=50
```

## API Authentication

All API requests (except registration and login) require authentication:

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "your_username", "password": "your_password"}'

# Use token in subsequent requests
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/configs
```

## Security Features

- **Password Hashing**: Bcrypt for secure password storage
- **JWT Tokens**: Secure authentication with expiration
- **CORS Protection**: Configured for frontend domain only
- **Input Validation**: Server-side validation for all inputs
- **Credential Protection**: Sensitive data is never logged or exposed
- **Audit Trail**: Complete logging of all actions
- **Active User Control**: Ability to activate/deactivate accounts
- **Admin Protection**: Admins cannot delete their own accounts

## Development

### Backend Development
```bash
# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Run tests
go test ./...

# Build for production
go build -o s3mgr
```

### Frontend Development
```bash
# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Deployment

### Backend Deployment
1. Build the Go binary:
   ```bash
   go build -o s3mgr
   ```

2. Set environment variables:
   ```bash
   export PORT=8080
   export JWT_SECRET=your-production-secret
   ```

3. Run the binary:
   ```bash
   ./s3mgr
   ```

### Frontend Deployment
1. Build the frontend:
   ```bash
   npm run build
   ```

2. Serve the `dist` directory with any static file server

## Troubleshooting

### Common Issues

1. **CORS Errors**: Make sure the backend CORS configuration includes your frontend URL
2. **Storage Access Denied**: Verify your storage credentials have the necessary permissions
3. **Database Errors**: Ensure the `data` directory is writable
4. **Port Conflicts**: Change the PORT environment variable if 8080 is in use

### Required S3 Permissions

Your AWS user needs the following S3 permissions:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::your-bucket-name",
                "arn:aws:s3:::your-bucket-name/*"
            ]
        }
    ]
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
