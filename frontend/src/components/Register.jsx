import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authAPI } from '../services/api'
import { UserPlus } from 'lucide-react'

function Register() {
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    confirmPassword: ''
  })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const navigate = useNavigate()

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match')
      setLoading(false)
      return
    }

    try {
      await authAPI.register({
        username: formData.username,
        password: formData.password
      })
      navigate('/login', { 
        state: { message: 'Registration successful! Please log in.' }
      })
    } catch (err) {
      setError(err.response?.data?.error || 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-green-50 via-white to-green-100 py-12 px-4 sm:px-6 lg:px-8 transition-all">
      <div className="max-w-md w-full">
        <div className="bg-white rounded-2xl shadow-xl p-8 space-y-8 animate-fade-in border border-green-100">
          <div className="flex flex-col items-center">
            <div className="mx-auto h-16 w-16 flex items-center justify-center rounded-full bg-green-100 shadow-md mb-2">
              <UserPlus className="h-8 w-8 text-green-600" />
            </div>
            <h2 className="text-center text-3xl font-extrabold text-gray-900 mb-2">
              Create your <span className="text-green-600">account</span>
            </h2>
            <p className="text-gray-500 text-center text-sm mb-2">Join S3 Manager to get started.</p>
          </div>
          <form className="space-y-6" onSubmit={handleSubmit} autoComplete="off">
            {error && (
              <div className="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded shadow text-center animate-pulse">
                {error}
              </div>
            )}
            <div className="space-y-4">
              <div>
                <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">Username</label>
                <input
                  id="username"
                  name="username"
                  type="text"
                  required
                  className="appearance-none block w-full px-4 py-2 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 text-gray-900 focus:outline-none focus:ring-2 focus:ring-green-400 focus:border-green-500 transition"
                  placeholder="Choose a username"
                  value={formData.username}
                  onChange={handleChange}
                  autoComplete="username"
                />
              </div>
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">Password</label>
                <div className="relative">
                  <input
                    id="password"
                    name="password"
                    type={showPassword ? 'text' : 'password'}
                    required
                    className="appearance-none block w-full px-4 py-2 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 text-gray-900 focus:outline-none focus:ring-2 focus:ring-green-400 focus:border-green-500 transition pr-10"
                    placeholder="Create a password"
                    value={formData.password}
                    onChange={handleChange}
                    autoComplete="new-password"
                  />
                  <button
                    type="button"
                    tabIndex={-1}
                    className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-green-600 focus:outline-none"
                    onClick={() => setShowPassword((v) => !v)}
                  >
                    {showPassword ? (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path d="M1 12s4-7 11-7 11 7 11 7-4 7-11 7S1 12 1 12z" strokeWidth="2" stroke="currentColor" fill="none"/><circle cx="12" cy="12" r="3" strokeWidth="2" stroke="currentColor" fill="none"/></svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path d="M17.94 17.94C16.11 19.23 14.13 20 12 20c-7 0-11-8-11-8a21.82 21.82 0 0 1 5.06-6.06M22.54 6.42A21.82 21.82 0 0 1 23 12s-4 8-11 8c-2.13 0-4.11-.77-5.94-2.06M1 1l22 22" strokeWidth="2" stroke="currentColor" fill="none"/><circle cx="12" cy="12" r="3" strokeWidth="2" stroke="currentColor" fill="none"/></svg>
                    )}
                  </button>
                </div>
              </div>
              <div>
                <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">Confirm Password</label>
                <div className="relative">
                  <input
                    id="confirmPassword"
                    name="confirmPassword"
                    type={showConfirmPassword ? 'text' : 'password'}
                    required
                    className="appearance-none block w-full px-4 py-2 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 text-gray-900 focus:outline-none focus:ring-2 focus:ring-green-400 focus:border-green-500 transition pr-10"
                    placeholder="Re-enter your password"
                    value={formData.confirmPassword}
                    onChange={handleChange}
                    autoComplete="new-password"
                  />
                  <button
                    type="button"
                    tabIndex={-1}
                    className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-green-600 focus:outline-none"
                    onClick={() => setShowConfirmPassword((v) => !v)}
                  >
                    {showConfirmPassword ? (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path d="M1 12s4-7 11-7 11 7 11 7-4 7-11 7S1 12 1 12z" strokeWidth="2" stroke="currentColor" fill="none"/><circle cx="12" cy="12" r="3" strokeWidth="2" stroke="currentColor" fill="none"/></svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path d="M17.94 17.94C16.11 19.23 14.13 20 12 20c-7 0-11-8-11-8a21.82 21.82 0 0 1 5.06-6.06M22.54 6.42A21.82 21.82 0 0 1 23 12s-4 8-11 8c-2.13 0-4.11-.77-5.94-2.06M1 1l22 22" strokeWidth="2" stroke="currentColor" fill="none"/><circle cx="12" cy="12" r="3" strokeWidth="2" stroke="currentColor" fill="none"/></svg>
                    )}
                  </button>
                </div>
              </div>
            </div>
            <button
              type="submit"
              disabled={loading}
              className="w-full py-2 px-4 bg-green-600 hover:bg-green-700 text-white font-semibold rounded-lg shadow-md focus:outline-none focus:ring-2 focus:ring-green-400 focus:ring-offset-2 transition disabled:opacity-50"
            >
              {loading ? 'Creating account...' : 'Sign up'}
            </button>
            <div className="text-center mt-4">
              <span className="text-sm text-gray-600">
                Already have an account?{' '}
                <Link to="/login" className="font-medium text-green-600 hover:text-green-500">
                  Sign in
                </Link>
              </span>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}

export default Register
