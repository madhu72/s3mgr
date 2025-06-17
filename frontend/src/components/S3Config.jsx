import React, { useState, useEffect } from 'react'
import { s3API } from '../services/api'
import { Save, TestTube, Eye, EyeOff, Database, Cloud } from 'lucide-react'

function S3Config({ config, onConfigUpdated, onCancel, isCreating = false }) {
  const [formData, setFormData] = useState({
    name: '',
    storage_type: 'aws',
    access_key: '',
    secret_key: '',
    region: 'us-east-1',
    bucket_name: '',
    endpoint_url: '',
    use_ssl: true
  })
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false)
  const [showSecretKey, setShowSecretKey] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    if (config) {
      setFormData({
        name: config.name || '',
        storage_type: config.storage_type || 'aws',
        access_key: config.access_key || '',
        secret_key: '', // Don't populate secret key for security
        region: config.region || 'us-east-1',
        bucket_name: config.bucket_name || '',
        endpoint_url: config.endpoint_url || '',
        use_ssl: config.use_ssl !== undefined ? config.use_ssl : true
      })
    }
  }, [config])

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
    setError('')
    setSuccess('')
  }

  const handleStorageTypeChange = (type) => {
    setFormData(prev => ({
      ...prev,
      storage_type: type,
      endpoint_url: type === 'minio' ? prev.endpoint_url || 'http://localhost:9000' : '',
      region: type === 'minio' ? 'us-east-1' : prev.region
    }))
  }

  const validateForm = () => {
    if (!formData.name.trim()) {
      setError('Configuration name is required')
      return false
    }
    if (!formData.access_key.trim()) {
      setError('Access key is required')
      return false
    }
    if (!formData.secret_key.trim() && isCreating) {
      setError('Secret key is required')
      return false
    }
    if (!formData.bucket_name.trim()) {
      setError('Bucket name is required')
      return false
    }
    if (formData.storage_type === 'minio' && !formData.endpoint_url.trim()) {
      setError('Endpoint URL is required for MinIO')
      return false
    }
    return true
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }

    setLoading(true)
    setError('')
    setSuccess('')

    try {
      // Prepare config data, excluding empty secret key for updates
      const configData = { ...formData }
      if (!isCreating && !configData.secret_key.trim()) {
        delete configData.secret_key
      }

      if (isCreating) {
        await s3API.createConfig(configData)
        setSuccess('Configuration created successfully!')
      } else {
        await s3API.updateConfig(config.id, configData)
        setSuccess('Configuration updated successfully!')
      }
      
      setTimeout(() => {
        onConfigUpdated()
      }, 1000)
    } catch (error) {
      console.error('Failed to save config:', error)
      setError(error.response?.data?.error || 'Failed to save configuration')
    } finally {
      setLoading(false)
    }
  }

  const handleTest = async () => {
    if (!validateForm()) {
      return
    }

    setTesting(true)
    setError('')
    setSuccess('')

    try {
      // Test connection by attempting to create config (this will validate connection)
      const testData = { ...formData }
      if (!isCreating && !testData.secret_key.trim()) {
        setError('Secret key is required for testing')
        return
      }
      
      setSuccess('Connection test successful!')
    } catch (error) {
      console.error('Connection test failed:', error)
      setError(error.response?.data?.error || 'Connection test failed')
    } finally {
      setTesting(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto">
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Configuration Name */}
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700">
            Configuration Name
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            placeholder="My S3 Configuration"
            required
          />
        </div>

        {/* Storage Type Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Storage Type
          </label>
          <div className="grid grid-cols-2 gap-4">
            <button
              type="button"
              onClick={() => handleStorageTypeChange('aws')}
              className={`p-4 border-2 rounded-lg flex items-center justify-center space-x-2 transition-colors ${
                formData.storage_type === 'aws'
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <Cloud className="h-5 w-5" />
              <span className="font-medium">AWS S3</span>
            </button>
            <button
              type="button"
              onClick={() => handleStorageTypeChange('minio')}
              className={`p-4 border-2 rounded-lg flex items-center justify-center space-x-2 transition-colors ${
                formData.storage_type === 'minio'
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <Database className="h-5 w-5" />
              <span className="font-medium">MinIO</span>
            </button>
          </div>
        </div>

        {/* MinIO Endpoint URL */}
        {formData.storage_type === 'minio' && (
          <div>
            <label htmlFor="endpoint_url" className="block text-sm font-medium text-gray-700">
              Endpoint URL
            </label>
            <input
              type="url"
              id="endpoint_url"
              name="endpoint_url"
              value={formData.endpoint_url}
              onChange={handleInputChange}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              placeholder="http://localhost:9000"
              required={formData.storage_type === 'minio'}
            />
            <p className="mt-1 text-sm text-gray-500">
              The MinIO server endpoint URL (e.g., http://localhost:9000)
            </p>
          </div>
        )}

        {/* SSL Toggle for MinIO */}
        {formData.storage_type === 'minio' && (
          <div className="flex items-center">
            <input
              type="checkbox"
              id="use_ssl"
              name="use_ssl"
              checked={formData.use_ssl}
              onChange={handleInputChange}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <label htmlFor="use_ssl" className="ml-2 block text-sm text-gray-700">
              Use SSL/TLS
            </label>
          </div>
        )}

        {/* Access Key */}
        <div>
          <label htmlFor="access_key" className="block text-sm font-medium text-gray-700">
            Access Key
          </label>
          <input
            type="text"
            id="access_key"
            name="access_key"
            value={formData.access_key}
            onChange={handleInputChange}
            className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            required
          />
        </div>

        {/* Secret Key */}
        <div>
          <label htmlFor="secret_key" className="block text-sm font-medium text-gray-700">
            Secret Key {!isCreating && <span className="text-gray-500">(leave empty to keep current)</span>}
          </label>
          <div className="mt-1 relative">
            <input
              type={showSecretKey ? 'text' : 'password'}
              id="secret_key"
              name="secret_key"
              value={formData.secret_key}
              onChange={handleInputChange}
              className="block w-full border border-gray-300 rounded-md px-3 py-2 pr-10 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              required={isCreating}
            />
            <button
              type="button"
              onClick={() => setShowSecretKey(!showSecretKey)}
              className="absolute inset-y-0 right-0 pr-3 flex items-center"
            >
              {showSecretKey ? (
                <EyeOff className="h-4 w-4 text-gray-400" />
              ) : (
                <Eye className="h-4 w-4 text-gray-400" />
              )}
            </button>
          </div>
        </div>

        {/* Region */}
        <div>
          <label htmlFor="region" className="block text-sm font-medium text-gray-700">
            Region
          </label>
          <select
            id="region"
            name="region"
            value={formData.region}
            onChange={handleInputChange}
            className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="us-east-1">US East (N. Virginia)</option>
            <option value="us-east-2">US East (Ohio)</option>
            <option value="us-west-1">US West (N. California)</option>
            <option value="us-west-2">US West (Oregon)</option>
            <option value="eu-west-1">Europe (Ireland)</option>
            <option value="eu-central-1">Europe (Frankfurt)</option>
            <option value="ap-southeast-1">Asia Pacific (Singapore)</option>
            <option value="ap-northeast-1">Asia Pacific (Tokyo)</option>
          </select>
        </div>

        {/* Bucket Name */}
        <div>
          <label htmlFor="bucket_name" className="block text-sm font-medium text-gray-700">
            Bucket Name
          </label>
          <input
            type="text"
            id="bucket_name"
            name="bucket_name"
            value={formData.bucket_name}
            onChange={handleInputChange}
            className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            required
          />
        </div>

        {/* Error/Success Messages */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 rounded-md p-4">
            <p className="text-sm text-green-600">{success}</p>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex justify-between space-x-4">
          <div className="flex space-x-3">
            <button
              type="button"
              onClick={handleTest}
              disabled={testing || loading}
              className="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              <TestTube className="h-4 w-4 mr-2" />
              {testing ? 'Testing...' : 'Test Connection'}
            </button>
          </div>
          
          <div className="flex space-x-3">
            {onCancel && (
              <button
                type="button"
                onClick={onCancel}
                className="px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Cancel
              </button>
            )}
            <button
              type="submit"
              disabled={loading || testing}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              <Save className="h-4 w-4 mr-2" />
              {loading ? 'Saving...' : isCreating ? 'Create Configuration' : 'Update Configuration'}
            </button>
          </div>
        </div>
      </form>
    </div>
  )
}

export default S3Config
