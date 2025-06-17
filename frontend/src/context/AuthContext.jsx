import React, { createContext, useContext, useState, useEffect } from 'react'

const AuthContext = createContext()

export function useAuth() {
  return useContext(AuthContext)
}

// Helper function to decode JWT token
const decodeToken = (token) => {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
    }).join(''))
    return JSON.parse(jsonPayload)
  } catch (error) {
    console.error('Error decoding token:', error)
    return null
  }
}

export function AuthProvider({ children }) {
  const [token, setToken] = useState(null)
  const [username, setUsername] = useState(null)
  const [userInfo, setUserInfo] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Initialize auth state from localStorage
    const storedToken = localStorage.getItem('token')
    const storedUsername = localStorage.getItem('username')
    
    if (storedToken) {
      setToken(storedToken)
      // Decode token to get user info
      const decoded = decodeToken(storedToken)
      if (decoded) {
        setUserInfo(decoded)
      }
    }
    if (storedUsername) {
      setUsername(storedUsername)
    }
    
    setLoading(false)
  }, [])

  useEffect(() => {
    if (token) {
      localStorage.setItem('token', token)
      // Decode token to get user info
      const decoded = decodeToken(token)
      if (decoded) {
        setUserInfo(decoded)
      }
    } else {
      localStorage.removeItem('token')
      setUserInfo(null)
    }
  }, [token])

  useEffect(() => {
    if (username) {
      localStorage.setItem('username', username)
    } else {
      localStorage.removeItem('username')
    }
  }, [username])

  const login = (token, username) => {
    setToken(token)
    setUsername(username)
  }

  const logout = () => {
    setToken(null)
    setUsername(null)
    setUserInfo(null)
  }

  const isAdmin = () => {
    return userInfo && userInfo.is_admin === true
  }

  const value = {
    token,
    username,
    userInfo,
    isAdmin,
    login,
    logout,
    loading
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}
