import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2, FileText, Filter, Loader } from 'lucide-react'
import Swal from 'sweetalert2'
import api from '../api'

const MONTHS = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
]

const currentYear = new Date().getFullYear()
const YEARS = Array.from({ length: 6 }, (_, i) => currentYear - 2 + i)

function formatRp(num) {
  const n = parseFloat(num) || 0
  return 'Rp ' + n.toLocaleString('id-ID')
}

export default function ExpenseReport() {
  const navigate = useNavigate()
  const [reports, setReports] = useState([])
  const [loading, setLoading] = useState(true)
  const [filterMonth, setFilterMonth] = useState('')
  const [filterYear, setFilterYear] = useState(currentYear)
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ month: new Date().getMonth() + 1, year: currentYear, keterangan: '' })
  const [saving, setSaving] = useState(false)

  const fetchReports = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      if (filterMonth) params.set('month', filterMonth)
      if (filterYear) params.set('year', filterYear)
      const res = await api.get(`/expenses?${params.toString()}`)
      setReports(res.data || [])
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchReports() }, [filterMonth, filterYear])

  const handleCreate = async (e) => {
    e.preventDefault()
    setSaving(true)
    try {
      const res = await api.post('/expenses', form)
      setShowModal(false)
      navigate(`/expenses/${res.data.id}`)
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id, month, year) => {
    const result = await Swal.fire({
      title: 'Hapus Laporan?',
      text: `Laporan ${MONTHS[month - 1]} ${year} akan dihapus permanen beserta semua itemnya.`,
      icon: 'warning',
      showCancelButton: true,
      confirmButtonText: 'Ya, Hapus',
      cancelButtonText: 'Batal',
      confirmButtonColor: '#EF4444',
    })
    if (!result.isConfirmed) return
    try {
      await api.delete(`/expenses/${id}`)
      fetchReports()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    }
  }

  return (
    <div>
      <div className="page-header-row">
        <div>
          <h1 className="page-title">Laporan Pengeluaran</h1>
          <p className="page-subtitle">Manajemen laporan pengeluaran bulanan</p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          <Plus size={16} /> Buat Laporan Baru
        </button>
      </div>

      {/* Filter Bar */}
      <div className="card" style={{ marginBottom: 24 }}>
        <div className="card-body" style={{ padding: '16px 20px' }}>
          <div className="filter-bar">
            <Filter size={16} style={{ color: 'var(--text-muted)' }} />
            <div className="form-group" style={{ marginBottom: 0, minWidth: 160 }}>
              <select className="form-control" value={filterMonth} onChange={e => setFilterMonth(e.target.value)}>
                <option value="">Semua Bulan</option>
                {MONTHS.map((m, i) => <option key={i} value={i + 1}>{m}</option>)}
              </select>
            </div>
            <div className="form-group" style={{ marginBottom: 0, minWidth: 120 }}>
              <select className="form-control" value={filterYear} onChange={e => setFilterYear(e.target.value)}>
                <option value="">Semua Tahun</option>
                {YEARS.map(y => <option key={y} value={y}>{y}</option>)}
              </select>
            </div>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="card">
        <div className="card-header">
          <span className="card-title">Daftar Laporan</span>
          <span className="badge badge-primary">{reports.length} laporan</span>
        </div>
        <div className="table-responsive">
          {loading ? (
            <div className="loading-state">
              <Loader className="spinner" size={32} />
              <p>Memuat data...</p>
            </div>
          ) : reports.length === 0 ? (
            <div className="empty-state">
              <FileText size={48} style={{ color: 'var(--text-light)', marginBottom: 12 }} />
              <p>Belum ada laporan pengeluaran</p>
              <button className="btn btn-primary btn-sm" style={{ marginTop: 12 }} onClick={() => setShowModal(true)}>
                <Plus size={14} /> Buat Laporan Pertama
              </button>
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Periode</th>
                  <th>Keterangan</th>
                  <th>Total Harga Barang</th>
                  <th>Total Pengeluaran</th>
                  <th>Total Kembalian</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {reports.map(r => (
                  <tr key={r.id} style={{ cursor: 'pointer' }} onClick={() => navigate(`/expenses/${r.id}`)}>
                    <td>
                      <strong>{MONTHS[r.month - 1]} {r.year}</strong>
                    </td>
                    <td style={{ color: 'var(--text-muted)', fontSize: 13 }}>{r.keterangan || '-'}</td>
                    <td>{formatRp(r.total_harga_barang)}</td>
                    <td>{formatRp(r.total_pengeluaran)}</td>
                    <td>{formatRp(r.total_kembalian)}</td>
                    <td>
                      <span className={`badge ${r.status === 'CONFIRMED' ? 'badge-success' : 'badge-warning'}`}>
                        {r.status === 'CONFIRMED' ? 'Selesai' : 'Draft'}
                      </span>
                    </td>
                    <td style={{ textAlign: 'right' }} onClick={e => e.stopPropagation()}>
                      <button
                        className="btn-icon delete"
                        title="Hapus"
                        onClick={() => handleDelete(r.id, r.month, r.year)}
                      >
                        <Trash2 size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-box" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">Buat Laporan Pengeluaran Baru</h3>
            </div>
            <form onSubmit={handleCreate}>
              <div className="modal-body">
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">Bulan</label>
                    <select
                      className="form-control"
                      value={form.month}
                      onChange={e => setForm(f => ({ ...f, month: parseInt(e.target.value) }))}
                      required
                    >
                      {MONTHS.map((m, i) => <option key={i} value={i + 1}>{m}</option>)}
                    </select>
                  </div>
                  <div className="form-group">
                    <label className="form-label">Tahun</label>
                    <select
                      className="form-control"
                      value={form.year}
                      onChange={e => setForm(f => ({ ...f, year: parseInt(e.target.value) }))}
                      required
                    >
                      {YEARS.map(y => <option key={y} value={y}>{y}</option>)}
                    </select>
                  </div>
                </div>
                <div className="form-group">
                  <label className="form-label">Keterangan (Opsional)</label>
                  <input
                    className="form-control"
                    placeholder="Misal: Pengeluaran operasional bulan Juli..."
                    value={form.keterangan}
                    onChange={e => setForm(f => ({ ...f, keterangan: e.target.value }))}
                  />
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setShowModal(false)}>Batal</button>
                <button type="submit" className="btn btn-primary" disabled={saving}>
                  {saving ? <><Loader size={14} className="spinner" /> Menyimpan...</> : <><Plus size={14} /> Buat Laporan</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
