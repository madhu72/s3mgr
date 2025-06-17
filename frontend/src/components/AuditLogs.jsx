import React, { useState, useEffect } from 'react'
import { useAuth } from '../context/AuthContext'
import { adminAPI } from '../services/api'
import { 
  FileText, 
  Search, 
  Filter, 
  Calendar, 
  Clock, 
  User, 
  Activity, 
  AlertCircle, 
  CheckCircle, 
  XCircle,
  RefreshCw,
  Download,
  Eye,
  ChevronDown,
  ChevronRight
} from 'lucide-react'

function AuditLogs() {
  const { token } = useAuth()
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [filters, setFilters] = useState({
    user_id: '',
    action: '',
    resource: '',
    start_time: '',
    end_time: '',
    limit: 100
  })
  const [expandedLogs, setExpandedLogs] = useState(new Set())
  const [error, setError] = useState('')
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    loadAuditLogs()
  }, [])

  const loadAuditLogs = async () => {
    try {
      setLoading(true)
      setError('')
      const response = await adminAPI.getAuditLogs(token, filters)
      setLogs(response.data.audit_logs || [])
    } catch (error) {
      console.error('Failed to load audit logs:', error)
      setError('Failed to load audit logs: ' + (error.response?.data?.error || error.message))
    } finally {
      setLoading(false)
    }
  }

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value
    }))
  }

  const applyFilters = () => {
    loadAuditLogs()
  }

  const clearFilters = () => {
    setFilters({
      user_id: '',
      action: '',
      resource: '',
      start_time: '',
      end_time: '',
      limit: 100
    })
    setSearchTerm('')
  }

  const toggleLogExpansion = (logId) => {
    const newExpanded = new Set(expandedLogs)
    if (newExpanded.has(logId)) {
      newExpanded.delete(logId)
    } else {
      newExpanded.add(logId)
    }
    setExpandedLogs(newExpanded)
  }

  const exportLogs = () => {
    const csvContent = [
      ['Timestamp', 'User', 'Action', 'Resource', 'Success', 'IP Address', 'Error'].join(','),
      ...filteredLogs.map(log => [
        log.timestamp,
        log.username || log.user_id,
        log.action,
        log.resource || '',
        log.success ? 'Yes' : 'No',
        log.client_ip || '',
        log.error || ''
      ].map(field => `"${field}"`).join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-logs-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  const filteredLogs = logs.filter(log => {
    if (!searchTerm) return true
    const searchLower = searchTerm.toLowerCase()
    return (
      (log.username && log.username.toLowerCase().includes(searchLower)) ||
      (log.action && log.action.toLowerCase().includes(searchLower)) ||
      (log.resource && log.resource.toLowerCase().includes(searchLower)) ||
      (log.client_ip && log.client_ip.toLowerCase().includes(searchLower)) ||
      (log.error && log.error.toLowerCase().includes(searchLower))
    )
  })

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString()
  }

  const getActionIcon = (action) => {
    switch (action) {
      case 'login':
      case 'logout':
        return <User className="h-4 w-4" />
      case 'create_user':
      case 'update_user':
      case 'delete_user':
        return <User className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  const getStatusIcon = (success) => {
    return success ? (
      <CheckCircle className="h-4 w-4 text-green-500" />
    ) : (
      <XCircle className="h-4 w-4 text-red-500" />
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading audit logs...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <FileText className="h-8 w-8 text-blue-600" />
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Audit Logs</h2>
            <p className="text-sm text-gray-600">Monitor all user and system activities</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <button
            onClick={exportLogs}
            className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 flex items-center space-x-2"
          >
            <Download className="h-4 w-4" />
            <span>Export CSV</span>
          </button>
          <button
            onClick={loadAuditLogs}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 flex items-center space-x-2"
          >
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-center">
          <AlertCircle className="h-5 w-5 mr-2" />
          {error}
        </div>
      )}

      {/* Filters */}
      <div className="bg-white p-6 rounded-lg shadow-sm border">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900 flex items-center">
            <Filter className="h-5 w-5 mr-2" />
            Filters
          </h3>
          <button
            onClick={clearFilters}
            className="text-sm text-gray-500 hover:text-gray-700"
          >
            Clear all
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              User ID
            </label>
            <input
              type="text"
              value={filters.user_id}
              onChange={(e) => handleFilterChange('user_id', e.target.value)}
              placeholder="Filter by user ID"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Action
            </label>
            <select
              value={filters.action}
              onChange={(e) => handleFilterChange('action', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All actions</option>
              <option value="login">Login</option>
              <option value="logout">Logout</option>
              <option value="register">Register</option>
              <option value="create_user">Create User</option>
              <option value="update_user">Update User</option>
              <option value="delete_user">Delete User</option>
              <option value="change_password">Change Password</option>
              <option value="query_audit_logs">Query Audit Logs</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Resource
            </label>
            <input
              type="text"
              value={filters.resource}
              onChange={(e) => handleFilterChange('resource', e.target.value)}
              placeholder="Filter by resource"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Start Date
            </label>
            <input
              type="datetime-local"
              value={filters.start_time}
              onChange={(e) => handleFilterChange('start_time', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              End Date
            </label>
            <input
              type="datetime-local"
              value={filters.end_time}
              onChange={(e) => handleFilterChange('end_time', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Limit
            </label>
            <select
              value={filters.limit}
              onChange={(e) => handleFilterChange('limit', parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            >
              <option value={50}>50 records</option>
              <option value={100}>100 records</option>
              <option value={250}>250 records</option>
              <option value={500}>500 records</option>
            </select>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div className="flex-1 max-w-md">
            <div className="relative">
              <Search className="h-4 w-4 absolute left-3 top-3 text-gray-400" />
              <input
                type="text"
                placeholder="Search logs..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <button
            onClick={applyFilters}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 ml-4"
          >
            Apply Filters
          </button>
        </div>
      </div>

      {/* Logs List */}
      <div className="bg-white rounded-lg shadow-sm border">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">
              Audit Logs ({filteredLogs.length} records)
            </h3>
          </div>
        </div>

        <div className="divide-y divide-gray-200">
          {filteredLogs.map((log) => (
            <div key={log.id} className="p-6 hover:bg-gray-50">
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-4 flex-1">
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(log.success)}
                    {getActionIcon(log.action)}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-4 mb-2">
                      <span className="text-sm font-medium text-gray-900">
                        {log.action.replace(/_/g, ' ').toUpperCase()}
                      </span>
                      {log.resource && (
                        <span className="text-sm text-gray-500">
                          → {log.resource}
                        </span>
                      )}
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        log.success 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-red-100 text-red-800'
                      }`}>
                        {log.success ? 'Success' : 'Failed'}
                      </span>
                    </div>
                    
                    <div className="flex items-center space-x-6 text-sm text-gray-500">
                      <div className="flex items-center">
                        <User className="h-3 w-3 mr-1" />
                        {log.username || log.user_id}
                      </div>
                      <div className="flex items-center">
                        <Clock className="h-3 w-3 mr-1" />
                        {formatDate(log.timestamp)}
                      </div>
                      {log.client_ip && (
                        <div className="flex items-center">
                          <Activity className="h-3 w-3 mr-1" />
                          {log.client_ip}
                        </div>
                      )}
                    </div>
                    
                    {log.error && (
                      <div className="mt-2 text-sm text-red-600 bg-red-50 p-2 rounded">
                        <strong>Error:</strong> {log.error}
                      </div>
                    )}
                  </div>
                </div>
                
                <button
                  onClick={() => toggleLogExpansion(log.id)}
                  className="text-gray-400 hover:text-gray-600 p-1"
                >
                  {expandedLogs.has(log.id) ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                </button>
              </div>
              
              {expandedLogs.has(log.id) && (
                <div className="mt-4 pl-10 border-l-2 border-gray-200">
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Details</h4>
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="font-medium text-gray-700">Log ID:</span>
                        <span className="ml-2 text-gray-600">{log.id}</span>
                      </div>
                      <div>
                        <span className="font-medium text-gray-700">Session ID:</span>
                        <span className="ml-2 text-gray-600">{log.session_id || 'N/A'}</span>
                      </div>
                      <div>
                        <span className="font-medium text-gray-700">User Agent:</span>
                        <span className="ml-2 text-gray-600 break-all">{log.user_agent || 'N/A'}</span>
                      </div>
                      <div>
                        <span className="font-medium text-gray-700">Resource ID:</span>
                        <span className="ml-2 text-gray-600">{log.resource_id || 'N/A'}</span>
                      </div>
                    </div>
                    
                    {log.details && Object.keys(log.details).length > 0 && (
                      <div className="mt-4">
                        <span className="font-medium text-gray-700">Additional Details:</span>
                        <pre className="mt-2 text-xs bg-white p-3 rounded border overflow-x-auto">
                          {JSON.stringify(log.details, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
        
        {filteredLogs.length === 0 && (
          <div className="text-center py-12">
            <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-500">No audit logs found</p>
            <p className="text-sm text-gray-400 mt-1">Try adjusting your filters or date range</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default AuditLogs
