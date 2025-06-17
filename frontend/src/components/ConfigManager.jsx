import React, { useState, useEffect } from 'react'
import { s3API } from '../services/api'
import { Plus, Edit, Trash2, Star, StarOff, Settings, Database, Cloud, X, Copy, Eye, EyeOff } from 'lucide-react'
import S3Config from './S3Config'
import AutoMinIOConfig from './AutoMinIOConfig'
import { useAuth } from '../context/AuthContext'

function ConfigManager({ onConfigSelected, selectedConfigId, onConfigsUpdated }) {
  const [configs, setConfigs] = useState([])
  const [loading, setLoading] = useState(false)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [editingConfig, setEditingConfig] = useState(null)
  const [error, setError] = useState('')
  const [deleteModal, setDeleteModal] = useState({ show: false, configId: null, configName: '' })
  const [showCredentials, setShowCredentials] = useState({}) // Track which configs show credentials
  const [copySuccess, setCopySuccess] = useState({}) // Track copy success messages
  const [importExportFormat, setImportExportFormat] = useState('csv')
  const { isAdmin } = useAuth()

  useEffect(() => {
    loadConfigs()
  }, [])

  // Bulk Export Configs
  const handleExportConfigs = async (format = 'csv') => {
    setError('');
    try {
      // You may need to get token from context if required by s3API
      const token = localStorage.getItem('token');
      const response = await s3API.exportConfigs(token, format);
      const blob = new Blob([response.data], { type: format === 'json' ? 'application/json' : 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `configs.${format}`;
      document.body.appendChild(a);
      a.click();
      a.remove();
    } catch (err) {
      setError('Failed to export configs: ' + (err.response?.data?.error || err.message));
    }
  };

  // Bulk Import Configs
  const handleImportConfigs = async (e) => {
    setError('');
    const file = e.target.files[0];
    if (!file) return;
    try {
      const token = localStorage.getItem('token');
      await s3API.importConfigs(token, file, importExportFormat);
      setError('');
      loadConfigs();
    } catch (err) {
      setError('Failed to import configs: ' + (err.response?.data?.error || err.message));
    }
    e.target.value = '';
  };

  const loadConfigs = async () => {
    setLoading(true)
    try {
      const response = await s3API.getConfigs()
      // Backend returns {configurations: [...]} not direct array
      const configList = Array.isArray(response.data.configurations) ? response.data.configurations : []
      setConfigs(configList)
      setError('')
    } catch (error) {
      console.error('Failed to load configs:', error)
      setConfigs([]) // Ensure configs is always an array
      setError('Failed to load configurations')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateConfig = () => {
    setEditingConfig(null)
    setShowCreateForm(true)
  }

  const handleEditConfig = (config) => {
    setEditingConfig(config)
    setShowCreateForm(true)
  }

  const handleDeleteClick = (config) => {
    setDeleteModal({
      show: true,
      configId: config.id,
      configName: config.name
    })
  }

  const handleDeleteConfirm = async () => {
    try {
      await s3API.deleteConfig(deleteModal.configId)
      await loadConfigs()
      setError('')
      // Notify parent component that configs were updated
      if (onConfigsUpdated) {
        onConfigsUpdated()
      }
    } catch (error) {
      console.error('Failed to delete config:', error)
      setError('Failed to delete configuration')
    } finally {
      setDeleteModal({ show: false, configId: null, configName: '' })
    }
  }

  const handleDeleteCancel = () => {
    setDeleteModal({ show: false, configId: null, configName: '' })
  }

  const copyToClipboard = async (text, configId, type) => {
    if (type === 'secret_key' || type === 'access_key') {
      try {
        // Fetch full config from backend
        const response = await s3API.getConfigById(configId)
        const value = response?.data?.[type]
        if (!value) {
          setError(`${type.replace('_', ' ')} not available for this configuration.`)
          return
        }
        await navigator.clipboard.writeText(value)
        setCopySuccess({ ...copySuccess, [`${configId}-${type}`]: true })
        setTimeout(() => {
          setCopySuccess(prev => ({ ...prev, [`${configId}-${type}`]: false }))
        }, 2000)
      } catch (err) {
        setError(`Failed to fetch ${type.replace('_', ' ')} for copying.`)
        console.error(`Failed to fetch/copy ${type}:`, err)
      }
    } else {
      try {
        await navigator.clipboard.writeText(text)
        setCopySuccess({ ...copySuccess, [`${configId}-${type}`]: true })
        setTimeout(() => {
          setCopySuccess(prev => ({ ...prev, [`${configId}-${type}`]: false }))
        }, 2000)
      } catch (err) {
        console.error('Failed to copy: ', err)
      }
    }
  }

  const toggleCredentials = (configId) => {
    setShowCredentials(prev => ({
      ...prev,
      [configId]: !prev[configId]
    }))
  }

  const handleConfigSaved = async () => {
    setShowCreateForm(false)
    setEditingConfig(null)
    await loadConfigs()
    // Notify parent component that configs were updated
    if (onConfigsUpdated) {
      onConfigsUpdated()
    }
  }

  const handleConfigCancel = () => {
    setShowCreateForm(false)
    setEditingConfig(null)
  }

  const handleSetDefault = async (configId) => {
    try {
      await s3API.setDefaultConfig(configId)
      await loadConfigs()
      setError('')
      // Notify parent component that configs were updated
      if (onConfigsUpdated) {
        onConfigsUpdated()
      }
    } catch (error) {
      console.error('Failed to set default config:', error)
      setError('Failed to set default configuration')
    }
  }

  const getStorageIcon = (storageType) => {
    return storageType === 'minio' ? Database : Cloud
  }

  if (showCreateForm) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-medium text-gray-900">
            {editingConfig ? 'Edit Configuration' : 'Create New Configuration'}
          </h2>
          <button
            onClick={handleConfigCancel}
            className="text-gray-400 hover:text-gray-600"
          >
            Cancel
          </button>
        </div>
        <S3Config
          config={editingConfig}
          onConfigUpdated={handleConfigSaved}
          onCancel={handleConfigCancel}
          isCreating={!editingConfig}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium text-gray-900">Storage Configurations</h2>
        <button
          onClick={handleCreateConfig}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <Plus className="h-4 w-4 mr-2" />
          Add Configuration
        </button>

        {/* Bulk Import/Export Buttons (Admins only) */}
        {isAdmin() && (
          <div className="flex items-center space-x-2">
            <button
              onClick={() => handleExportConfigs('csv')}
              className="bg-green-600 text-white px-3 py-2 rounded-lg hover:bg-green-700"
              title="Export configs as CSV"
            >Export CSV</button>
            <button
              onClick={() => handleExportConfigs('json')}
              className="bg-green-600 text-white px-3 py-2 rounded-lg hover:bg-green-700"
              title="Export configs as JSON"
            >Export JSON</button>
            <label className="bg-blue-100 text-blue-700 px-3 py-2 rounded-lg hover:bg-blue-200 cursor-pointer ml-2">
              Import
              <input
                type="file"
                accept=".csv,.json"
                style={{ display: 'none' }}
                onChange={handleImportConfigs}
              />
            </label>
            <select value={importExportFormat} onChange={e => setImportExportFormat(e.target.value)} className="ml-2 border rounded px-2 py-1">
              <option value="csv">CSV</option>
              <option value="json">JSON</option>
            </select>
          </div>
        )}
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      {/* Show AutoMinIOConfig only when there are no configurations */}
      {!loading && (!Array.isArray(configs) || configs.length === 0) && (
        <AutoMinIOConfig onConfigCreated={handleConfigSaved} />
      )}

      {loading ? (
        <div className="text-center py-8">
          <Database className="mx-auto h-12 w-12 text-gray-400" />
          <p className="mt-2 text-sm text-gray-600">Loading configurations...</p>
        </div>
      ) : !Array.isArray(configs) || configs.length === 0 ? (
        <div className="text-center py-8">
          <Settings className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No configurations</h3>
          <p className="mt-1 text-sm text-gray-500">Get started by creating your first storage configuration.</p>
          <div className="mt-6">
            <button
              onClick={handleCreateConfig}
              className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Configuration
            </button>
          </div>
        </div>
      ) : (
        <div className="grid gap-4">
          {Array.isArray(configs) && configs.map((config) => {
            const StorageIcon = getStorageIcon(config.storage_type)
            const isSelected = selectedConfigId === config.id
            const isDefault = config.is_default

            return (
              <div
                key={config.id}
                className={`border rounded-lg p-4 cursor-pointer transition-colors ${
                  isSelected
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300 bg-white'
                }`}
                onClick={() => onConfigSelected(config.id)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <StorageIcon className="h-5 w-5 text-gray-400" />
                    <div>
                      <h3 className="text-sm font-medium text-gray-900 flex items-center">
                        {config.name}
                        {isDefault && (
                          <Star className="h-4 w-4 ml-2 text-yellow-400 fill-current" />
                        )}
                      </h3>
                      <p className="text-sm text-gray-500">
                        {config.storage_type === 'minio' ? 'MinIO' : 'AWS S3'} • {config.bucket_name}
                      </p>
                      {config.storage_type === 'minio' && config.endpoint_url && (
                        <p className="text-xs text-gray-400">{config.endpoint_url}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        toggleCredentials(config.id)
                      }}
                      className="p-1 text-gray-400 hover:text-blue-500"
                      title={showCredentials[config.id] ? "Hide credentials" : "Show credentials"}
                    >
                      {showCredentials[config.id] ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                    {!isDefault && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleSetDefault(config.id)
                        }}
                        className="p-1 text-gray-400 hover:text-yellow-500"
                        title="Set as default"
                      >
                        <StarOff className="h-4 w-4" />
                      </button>
                    )}
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleEditConfig(config)
                      }}
                      className="p-1 text-gray-400 hover:text-blue-500"
                      title="Edit configuration"
                    >
                      <Edit className="h-4 w-4" />
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeleteClick(config)
                      }}
                      className="p-1 text-gray-400 hover:text-red-500"
                      title="Delete configuration"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
                
                {/* Credentials section - shown when toggled */}
                {showCredentials[config.id] && (
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <h4 className="text-sm font-medium text-gray-700 mb-3">Credentials</h4>
                    <div className="space-y-3">
                      {/* Access Key */}
                      <div>
                        <label className="block text-xs font-medium text-gray-500 mb-1">Access Key</label>
                        <div className="flex items-center space-x-2">
                          <code className="flex-1 px-2 py-1 bg-gray-50 border rounded text-sm font-mono text-gray-800">
                            {config.access_key}
                          </code>
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              copyToClipboard(config.access_key, config.id, 'access_key')
                            }}
                            className="p-1 text-gray-400 hover:text-blue-500"
                            title="Copy access key"
                          >
                            <Copy className="h-4 w-4" />
                          </button>
                          {copySuccess[`${config.id}-access_key`] && (
                            <span className="text-xs text-green-600">Copied!</span>
                          )}
                        </div>
                      </div>
                      
                      {/* Secret Key */}
                      <div>
                        <label className="block text-xs font-medium text-gray-500 mb-1">Secret Key</label>
                        <div className="flex items-center space-x-2">
                          <code className="flex-1 px-2 py-1 bg-gray-50 border rounded text-sm font-mono text-gray-800">
                            {config.secret_key}
                          </code>
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              copyToClipboard(config.secret_key, config.id, 'secret_key')
                            }}
                            className="p-1 text-gray-400 hover:text-blue-500"
                            title="Copy secret key"
                          >
                            <Copy className="h-4 w-4" />
                          </button>
                          {copySuccess[`${config.id}-secret_key`] && (
                            <span className="text-xs text-green-600">Copied!</span>
                          )}
                        </div>
                      </div>
                      
                      {/* Endpoint URL for MinIO */}
                      {config.storage_type === 'minio' && config.endpoint_url && (
                        <div>
                          <label className="block text-xs font-medium text-gray-500 mb-1">Endpoint URL</label>
                          <div className="flex items-center space-x-2">
                            <code className="flex-1 px-2 py-1 bg-gray-50 border rounded text-sm font-mono text-gray-800">
                              {config.use_ssl ? 'https://' : 'http://'}{config.endpoint_url}
                            </code>
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                copyToClipboard(`${config.use_ssl ? 'https://' : 'http://'}${config.endpoint_url}`, config.id, 'endpoint')
                              }}
                              className="p-1 text-gray-400 hover:text-blue-500"
                              title="Copy endpoint URL"
                            >
                              <Copy className="h-4 w-4" />
                            </button>
                            {copySuccess[`${config.id}-endpoint`] && (
                              <span className="text-xs text-green-600">Copied!</span>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
      {deleteModal.show && (
        <div className="fixed inset-0 bg-gray-900 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl p-6 w-96 max-w-md mx-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Delete Configuration</h2>
              <button
                onClick={handleDeleteCancel}
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <p className="text-sm text-gray-600 mb-6">
              Are you sure you want to delete <span className="font-medium text-gray-900">"{deleteModal.configName}"</span>? This action cannot be undone.
            </p>
            <div className="flex justify-end space-x-3">
              <button
                onClick={handleDeleteCancel}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 rounded-md transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ConfigManager

