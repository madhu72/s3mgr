import axios from 'axios'

const API_BASE_URL = 'http://localhost:8081/api'

const api = axios.create({
  baseURL: API_BASE_URL,
})

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  register: (userData) => api.post('/auth/register', userData),
  logout: () => api.post('/auth/logout'), // New: call backend logout endpoint
}

// S3 API
export const s3API = {
  // Bulk Config Export (GET, returns CSV or JSON)
  exportConfigs: (token, format = 'csv') => api.get(`/admin/configs/export?format=${format}`, {
    headers: { Authorization: `Bearer ${token}` },
    responseType: 'blob',
  }),
  // Bulk Config Import (POST, accepts CSV or JSON file)
  importConfigs: (token, file, format = 'csv') => {
    const formData = new FormData();
    formData.append('file', file);
    return api.post(`/admin/configs/import?format=${format}`, formData, {
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'multipart/form-data' },
    });
  },
  // File operations (with optional config_id)
  getFiles: (configId = null) => {
    const params = configId ? { config_id: configId } : {}
    return api.get('/files', { params })
  },
  uploadFile: (formData, configId = null) => {
    if (configId) {
      formData.append('config_id', configId)
    }
    return api.post('/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },
  downloadFile: (key, configId = null) => {
    const params = configId ? { config_id: configId } : {}
    return api.get(`/files/download/${encodeURIComponent(key)}`, {
      params,
      responseType: 'blob',
    })
  },
  deleteFile: async (key, configId) => {
    return await api.delete(`/files/${key}?config_id=${configId}`)
  },
  
  // Configuration management
  getConfigs: () => api.get('/configs'),
  createConfig: (config) => api.post('/configs', config),
  updateConfig: (configId, config) => api.put(`/configs/${configId}`, config),
  deleteConfig: (configId) => api.delete(`/configs/${configId}`),
  setDefaultConfig: (configId) => api.post(`/configs/${configId}/default`),

  // Fetch full config by ID (including secret_key)
  getConfigById: async (configId) => {
    return await api.get(`/configs/${configId}`)
  },

  // Auto configure MinIO
  autoConfigureMinIO: async (username) => {
    return await api.post('/configs/auto-minio', { username })
  }
}

// Admin API
export const adminAPI = {
  // Bulk User Export (GET, returns CSV or JSON)
  exportUsers: (token, format = 'csv') => api.get(`/admin/users/export?format=${format}`, {
    headers: { Authorization: `Bearer ${token}` },
    responseType: 'blob',
  }),
  // Bulk User Import (POST, accepts CSV or JSON file)
  importUsers: (token, file, format = 'csv') => {
    const formData = new FormData();
    formData.append('file', file);
    return api.post(`/admin/users/import?format=${format}`, formData, {
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'multipart/form-data' },
    });
  },
  // User Management
  getUsers: (token) => api.get('/admin/users', {
    headers: { Authorization: `Bearer ${token}` }
  }),
  createUser: (token, userData) => api.post('/admin/users', userData, {
    headers: { Authorization: `Bearer ${token}` }
  }),
  updateUser: (token, username, userData) => api.put(`/admin/users/${username}`, userData, {
    headers: { Authorization: `Bearer ${token}` }
  }),
  deleteUser: (token, username) => api.delete(`/admin/users/${username}`, {
    headers: { Authorization: `Bearer ${token}` }
  }),
  getUserConfig: (token, username) => api.get(`/admin/users/${username}/config`, {
    headers: { Authorization: `Bearer ${token}` }
  }),

  // Audit Logs
  getAuditLogs: (token, filters = {}) => {
    const params = new URLSearchParams()
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== '' && value !== null && value !== undefined) {
        params.append(key, value)
      }
    })
    return api.get(`/admin/audit-logs?${params.toString()}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  filterAuditLogs: (token, filterData) => api.post('/admin/audit-logs/filter', filterData, {
    headers: { Authorization: `Bearer ${token}` }
  }),
  getAuditLogsByIncident: (token, sessionId) => api.get(`/admin/audit-logs/incident/${sessionId}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
}

export default api
