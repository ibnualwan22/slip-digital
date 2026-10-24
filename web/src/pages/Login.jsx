import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { TOKEN_KEY } from '../api'

export default function Login() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      const result = await res.json()
      if (!result.success) throw new Error(result.message || 'Login gagal')
      localStorage.setItem(TOKEN_KEY, result.data.token)
      localStorage.setItem('emaliyah_role', result.data.role || 'admin')
      localStorage.setItem('emaliyah_name', result.data.name || '')
      navigate(result.data.role === 'pengajar' ? '/rekap-saya' : '/dashboard', { replace: true })
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const wrap = { minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(135deg,#1E3A5F,#2563EB)', padding: 16 }
  const card = { background: '#fff', borderRadius: 16, padding: '32px 28px', width: '100%', maxWidth: 380, boxShadow: '0 20px 50px rgba(0,0,0,.25)' }
  const label = { display: 'block', fontSize: 13, fontWeight: 600, color: '#374151', marginBottom: 6 }
  const input = { width: '100%', padding: '10px 12px', border: '1px solid #D1D5DB', borderRadius: 10, fontSize: 14, marginBottom: 16, boxSizing: 'border-box' }
  const button = { width: '100%', padding: '11px 12px', background: '#1E3A5F', color: '#fff', border: 'none', borderRadius: 10, fontSize: 15, fontWeight: 700, cursor: 'pointer' }

  return (
    <div style={wrap}>
      <form style={card} onSubmit={submit}>
        <h1 style={{ margin: '0 0 4px', fontSize: 22, color: '#1E3A5F' }}>E-Maliyah</h1>
        <p style={{ margin: '0 0 24px', fontSize: 13, color: '#6B7280' }}>Markaz Arabiyah — silakan masuk untuk melanjutkan.</p>
        {error && (
          <div style={{ background: '#FEF2F2', color: '#B91C1C', border: '1px solid #FECACA', borderRadius: 10, padding: '10px 12px', fontSize: 13, marginBottom: 16 }}>{error}</div>
        )}
        <label style={label}>Username</label>
        <input style={input} value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" required />
        <label style={label}>Password</label>
        <input style={input} type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" required />
        <button style={button} type="submit" disabled={loading}>{loading ? 'Memproses…' : 'Masuk'}</button>
      </form>
    </div>
  )
}
