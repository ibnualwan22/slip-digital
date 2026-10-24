import React from 'react'
import { Link } from 'react-router-dom'

// Bilah navigasi bersama halaman pengajar: dua tab + jalan kembali ke Siakad.
// Periode ikut terbawa lewat ?period_id= agar tidak perlu pilih ulang.
export default function TeacherNav({ active, periodId }) {
  const q = periodId ? `?period_id=${periodId}` : ''
  const tabStyle = (isActive) => ({
    flex: 1,
    textAlign: 'center',
    padding: '10px 8px',
    borderRadius: 10,
    fontWeight: 700,
    fontSize: 14,
    textDecoration: 'none',
    background: isActive ? '#1E3A5F' : 'transparent',
    color: isActive ? '#fff' : '#374151',
  })
  return (
    <div style={{ marginBottom: 16 }}>
      <div style={{ display: 'flex', gap: 6, background: '#fff', border: '1px solid #E5E7EB', borderRadius: 12, padding: 6 }}>
        <Link to={`/rincian-bisyaroh${q}`} style={tabStyle(active === 'rincian')}>Rincian Bisyaroh</Link>
        <Link to={`/rekap-saya${q}`} style={tabStyle(active === 'rekap')}>Rekap Saya</Link>
      </div>
      <div style={{ marginTop: 8, textAlign: 'right' }}>
        <a href="https://siakad.markazarabiyah.site/admin/dashboard" style={{ fontSize: 13, color: '#2563EB', fontWeight: 600, textDecoration: 'none' }}>← Kembali ke Siakad</a>
      </div>
    </div>
  )
}
