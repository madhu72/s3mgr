import React, { useState, useEffect } from 'react'
import { useAuth } from '../context/AuthContext'
import { s3API } from '../services/api'
import FileList from './FileList'
import FileUpload from './FileUpload'
import ConfigManager from './ConfigManager'
import UserManagement from './UserManagement'
import AuditLogs from './AuditLogs'
import { LogOut, Settings, Upload, Files, Database, Users, FileText, Shield } from 'lucide-react'

function Dashboard() {
  const { username, logout, isAdmin } = useAuth()
  const [activeTab, setActiveTab] = useState('files')
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(false)
  const [configs, setConfigs] = useState([])
  const [selectedConfigId, setSelectedConfigId] = useState(null)
  const [configsLoading, setConfigsLoading] = useState(true)
  // Pagination state for files
  const [filesPage, setFilesPage] = useState(1)
  const [filesPageSize, setFilesPageSize] = useState(10)
  const [filesTotal, setFilesTotal] = useState(0)

  useEffect(() => {
    loadConfigs()
  }, [])

  useEffect(() => {
    if (selectedConfigId) {
      loadFiles()
    }
  }, [selectedConfigId, filesPage, filesPageSize])

  const loadConfigs = async () => {
    try {
      setConfigsLoading(true)
      const response = await s3API.getConfigs()
      // Backend returns {configurations: [...]} not direct array
      const configList = Array.isArray(response.data.configurations) ? response.data.configurations : []
      setConfigs(configList)
      
      // Auto-select default config or first config
      const defaultConfig = configList.find(c => c.is_default)
      const configToSelect = defaultConfig || configList[0]
      if (configToSelect) {
        setSelectedConfigId(configToSelect.id)
      }
    } catch (error) {
      console.error('Failed to load configs:', error)
      setConfigs([]) // Ensure configs is always an array
    } finally {
      setConfigsLoading(false)
    }
  }

  const loadFiles = async () => {
    if (!selectedConfigId) {
      console.log('loadFiles: No selectedConfigId, skipping')
      return
    }
    
    console.log('loadFiles: Loading files for configId:', selectedConfigId)
    setLoading(true)
    try {
      // Pass pagination params to API
      const response = await s3API.getFiles(selectedConfigId, { page: filesPage, page_size: filesPageSize })
      console.log('loadFiles: API response:', response)
      console.log('loadFiles: Response data:', response.data)
      
      // Backend returns files under 'files' key and total count
      const fileList = Array.isArray(response.data.files) ? response.data.files : []
      setFiles(fileList)
      setFilesTotal(response.data.total || 0)
    } catch (error) {
      console.error('loadFiles: Failed to load files:', error)
      console.error('loadFiles: Error response:', error.response)
      setFiles([])
      setFilesTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const handleFileUploaded = () => {
    loadFiles()
  }

  const handleFileDeleted = () => {
    loadFiles()
  }

  const handleConfigSelected = (configId) => {
    setSelectedConfigId(configId)
  }

  const handleConfigsUpdated = () => {
    loadConfigs()
  }

  const selectedConfig = Array.isArray(configs) ? configs.find(c => c.id === selectedConfigId) : null

  const tabs = [
    { id: 'files', label: 'Files', icon: Files },
    { id: 'upload', label: 'Upload', icon: Upload },
    { id: 'config', label: 'Configurations', icon: Database },
    ...(isAdmin() ? [
      { id: 'users', label: 'Users', icon: Users },
      { id: 'audit', label: 'Audit Logs', icon: Shield },
    ] : []),
  ]

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">S3 Manager</h1>
              <div className="flex items-center space-x-2">
                <p className="text-sm text-gray-600">Welcome, {username}</p>
                {isAdmin() && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                    <Shield className="h-3 w-3 mr-1" />
                    Admin
                  </span>
                )}
              </div>
              {selectedConfig && (
                <p className="text-xs text-gray-500">
                  Active: {selectedConfig.name} ({selectedConfig.storage_type === 'minio' ? 'MinIO' : 'AWS S3'})
                </p>
              )}
            </div>
            <button
              onClick={logout}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
            >
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Configuration Selection Bar */}
      {Array.isArray(configs) && configs.length > 1 && activeTab !== 'config' && (
        <div className="bg-white border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="py-3">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Active Configuration:
              </label>
              <select
                value={selectedConfigId || ''}
                onChange={(e) => setSelectedConfigId(e.target.value)}
                className="block w-64 border border-gray-300 rounded-md px-3 py-2 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              >
                {configs.map((config) => (
                  <option key={config.id} value={config.id}>
                    {config.name} ({config.storage_type === 'minio' ? 'MinIO' : 'AWS S3'})
                    {config.is_default ? ' (Default)' : ''}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-8">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm inline-flex items-center ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <Icon className="h-4 w-4 mr-2" />
                  {tab.label}
                </button>
              )
            })}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* Loading state */}
          {configsLoading ? (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              <p className="mt-2 text-sm text-gray-600">Loading configurations...</p>
            </div>
          ) : (
            <>
              {/* No configurations message - only show when not on config tab and no configs exist */}
              {Array.isArray(configs) && configs.length === 0 && activeTab !== 'config' && (
                <div className="text-center py-12">
                  <Database className="mx-auto h-12 w-12 text-gray-400" />
                  <h3 className="mt-2 text-sm font-medium text-gray-900">No configurations found</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    You need to create a storage configuration before you can manage files.
                  </p>
                  <div className="mt-6">
                    <button
                      onClick={() => setActiveTab('config')}
                      className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                    >
                      <Database className="h-4 w-4 mr-2" />
                      Create Configuration
                    </button>
                  </div>
                </div>
              )}

              {/* Files tab content */}
              {activeTab === 'files' && (
                <>
                  {selectedConfigId ? (
                    <FileList 
                      files={files} 
                      loading={loading} 
                      onFileDeleted={handleFileDeleted}
                      onRefresh={loadFiles}
                      configId={selectedConfigId}
                      page={filesPage}
                      setPage={setFilesPage}
                      pageSize={filesPageSize}
                      setPageSize={setFilesPageSize}
                      total={filesTotal}
                    />
                  ) : Array.isArray(configs) && configs.length > 0 ? (
                    <div className="text-center py-12">
                      <Files className="mx-auto h-12 w-12 text-gray-400" />
                      <h3 className="mt-2 text-sm font-medium text-gray-900">Select a configuration</h3>
                      <p className="mt-1 text-sm text-gray-500">
                        Choose a storage configuration to view files.
                      </p>
                    </div>
                  ) : Array.isArray(configs) && configs.length === 0 ? (
                    <div className="text-center py-12">
                      <Files className="mx-auto h-12 w-12 text-gray-400" />
                      <h3 className="mt-2 text-sm font-medium text-gray-900">No configurations found</h3>
                      <p className="mt-1 text-sm text-gray-500">
                        Create a storage configuration first to view files.
                      </p>
                      <button
                        onClick={() => setActiveTab('config')}
                        className="mt-4 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                      >
                        <Settings className="mr-2 h-4 w-4" />
                        Create Configuration
                      </button>
                    </div>
                  ) : null}
                </>
              )}

              {/* Upload tab content */}
              {activeTab === 'upload' && (
                <>
                  {selectedConfigId ? (
                    <FileUpload 
                      onFileUploaded={handleFileUploaded} 
                      configId={selectedConfigId}
                    />
                  ) : Array.isArray(configs) && configs.length > 0 ? (
                    <div className="text-center py-12">
                      <Upload className="mx-auto h-12 w-12 text-gray-400" />
                      <h3 className="mt-2 text-sm font-medium text-gray-900">Select a configuration</h3>
                      <p className="mt-1 text-sm text-gray-500">
                        Choose a storage configuration to upload files.
                      </p>
                    </div>
                  ) : Array.isArray(configs) && configs.length === 0 ? (
                    <div className="text-center py-12">
                      <Upload className="mx-auto h-12 w-12 text-gray-400" />
                      <h3 className="mt-2 text-sm font-medium text-gray-900">No configurations found</h3>
                      <p className="mt-1 text-sm text-gray-500">
                        Create a storage configuration first to upload files.
                      </p>
                      <button
                        onClick={() => setActiveTab('config')}
                        className="mt-4 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                      >
                        <Settings className="mr-2 h-4 w-4" />
                        Create Configuration
                      </button>
                    </div>
                  ) : null}
                </>
              )}

              {/* Config tab content - always accessible */}
              {activeTab === 'config' && (
                <ConfigManager
                  onConfigSelected={handleConfigSelected}
                  selectedConfigId={selectedConfigId}
                  onConfigsUpdated={handleConfigsUpdated}
                />
              )}

              {/* Users tab content */}
              {activeTab === 'users' && (
                <UserManagement />
              )}

              {/* Audit tab content */}
              {activeTab === 'audit' && (
                <AuditLogs />
              )}
            </>
          )}
        </div>
      </main>
    </div>
  )
}

export default Dashboard
