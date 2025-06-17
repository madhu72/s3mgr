import React, { useState } from 'react'
import { Server, User, AlertCircle, CheckCircle } from 'lucide-react'
import { s3API } from '../services/api'

const AutoMinIOConfig = ({ onConfigCreated }) => {
  const [username, setUsername] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const handleAutoConfig = async (e) => {
    e.preventDefault()
    if (!username.trim()) {
      setError('Username is required')
      return
    }

    setLoading(true)
    setError('')
    setSuccess('')

    try {
      const response = await s3API.autoConfigureMinIO(username.trim())
      setSuccess('MinIO configuration created successfully!')
      setUsername('')
      
      // Call the callback to refresh configurations
      if (onConfigCreated) {
        onConfigCreated(response.data.config)
      }
    } catch (error) {
      console.error('Auto MinIO config error:', error)
      setError(error.response?.data?.error || 'Failed to create MinIO configuration')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
      <div className="flex items-center mb-4">
        <Server className="h-6 w-6 text-blue-600 mr-2" />
        <h3 className="text-lg font-semibold text-blue-900">Auto MinIO Setup</h3>
      </div>
      
      <p className="text-blue-700 mb-4">
        Automatically create a MinIO configuration with dedicated user credentials and bucket.
        This uses the admin credentials configured in your .env file.
      </p>

      <form onSubmit={handleAutoConfig} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Username
          </label>
          <div className="relative">
            <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              disabled={loading}
            />
          </div>
        </div>

        {error && (
          <div className="flex items-center p-3 bg-red-50 border border-red-200 rounded-md">
            <AlertCircle className="h-4 w-4 text-red-500 mr-2" />
            <span className="text-red-700 text-sm">{error}</span>
          </div>
        )}

        {success && (
          <div className="flex items-center p-3 bg-green-50 border border-green-200 rounded-md">
            <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
            <span className="text-green-700 text-sm">{success}</span>
          </div>
        )}

        <button
          type="submit"
          disabled={loading || !username.trim()}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? (
            <div className="flex items-center justify-center">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Creating Configuration...
            </div>
          ) : (
            'Create MinIO Configuration'
          )}
        </button>
      </form>

      <div className="mt-4 p-3 bg-gray-50 border border-gray-200 rounded-md">
        <h4 className="text-sm font-medium text-gray-700 mb-2">What this does:</h4>
        <ul className="text-xs text-gray-600 space-y-1">
          <li>• Creates a dedicated MinIO user with unique credentials</li>
          <li>• Creates a private bucket for your files</li>
          <li>• Sets up proper access policies</li>
          <li>• Saves the configuration to your S3 Manager</li>
        </ul>
      </div>
    </div>
  )
}

export default AutoMinIOConfig
