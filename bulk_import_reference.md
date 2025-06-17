# S3 Manager Bulk Import Reference

This document provides sample formats for bulk importing users and configurations using the admin import features. It covers both CSV and JSON, and explains how to handle passwords securely.

---

## User Import Samples

### CSV Format
- **Passwords are only required for new users.**
- Omit the `password` column for export or for updating users without changing their password.

**Example:**
```csv
id,username,password,email,is_admin,is_active,created_at,updated_at,last_login
4,newuser,supersecret,newuser@example.com,false,true,2025-06-01T09:00:00Z,2025-06-01T09:00:00Z,2025-06-01T09:00:00Z
2,jane,,jane@example.com,false,true,2024-02-01T11:00:00Z,2024-02-01T11:00:00Z,2025-06-01T09:00:00Z
```
- In this example, `newuser` will be created with password `supersecret`.
- `jane` will be updated, but her password will remain unchanged (empty password field).

### JSON Format
- Include a `"password"` field for new users, omit or leave blank for existing users.

**Example:**
```json
[
  {
    "id": "4",
    "username": "newuser",
    "password": "supersecret",
    "email": "newuser@example.com",
    "is_admin": false,
    "is_active": true,
    "created_at": "2025-06-01T09:00:00Z",
    "updated_at": "2025-06-01T09:00:00Z",
    "last_login": "2025-06-01T09:00:00Z"
  },
  {
    "id": "2",
    "username": "jane",
    "email": "jane@example.com",
    "is_admin": false,
    "is_active": true,
    "created_at": "2024-02-01T11:00:00Z",
    "updated_at": "2024-02-01T11:00:00Z",
    "last_login": "2025-06-01T09:00:00Z"
  }
]
```

---

## Configuration Import Samples

### CSV Format
```csv
id,name,storage_type,bucket_name,access_key,secret_key,endpoint_url,use_ssl,is_default,created_at,updated_at
1,Main MinIO,minio,mybucket,admin,admin123,minio.example.com,true,true,2024-01-01T10:00:00Z,2025-06-01T09:00:00Z
2,Backup S3,aws,backup-bucket,AKIAIOSFODNN7,SECRET123,,false,false,2024-02-01T11:00:00Z,2025-06-01T09:00:00Z
```

### JSON Format
```json
[
  {
    "id": "1",
    "name": "Main MinIO",
    "storage_type": "minio",
    "bucket_name": "mybucket",
    "access_key": "admin",
    "secret_key": "admin123",
    "endpoint_url": "minio.example.com",
    "use_ssl": true,
    "is_default": true,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2025-06-01T09:00:00Z"
  },
  {
    "id": "2",
    "name": "Backup S3",
    "storage_type": "aws",
    "bucket_name": "backup-bucket",
    "access_key": "AKIAIOSFODNN7",
    "secret_key": "SECRET123",
    "endpoint_url": "",
    "use_ssl": false,
    "is_default": false,
    "created_at": "2024-02-01T11:00:00Z",
    "updated_at": "2025-06-01T09:00:00Z"
  }
]
```

---

## Password Security Notes
- Passwords are **never exported**.
- Passwords are **required for new users** during import, and are securely hashed by the backend.
- For existing users, leave the password field blank or omit it to keep their password unchanged.

---

For further details or more sample templates, contact your S3 Manager admin or development team.
