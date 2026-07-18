import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Eye, Share2, Printer, CheckCircle, Search, RefreshCw } from 'lucide-react'
import api from '../api'
import { formatRupiah, getMonthName, getStatusConfig } from '../utils/formatter'
import Swal from 'sweetalert2'

export default function Payroll() {
  const [payrolls, setPayrolls] = useState([])
  const [loading, setLoading] = useState(true)

  const [filterMonth, setFilterMonth] = useState(new Date().getMonth() + 1)
  const [filterYear, setFilterYear] = useState(new Date().getFullYear())

  useEffect(() => {
    loadPayrolls()
  }, [filterMonth, filterYear])

  const loadPayrolls = async () => {
    setLoading(true)
    try {
      const res = await api.get(`/payroll?month=${filterMonth}&year=${filterYear}`)
      setPayrolls(res.data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const generatePayroll = async () => {
    try {
      setLoading(true)
      await api.post(`/payroll/generate?month=${filterMonth}&year=${filterYear}`)
      
      // Auto-sync Siakad
      try {
        await api.post('/siakad/sync-all', { bulan: parseInt(filterMonth), tahun: parseInt(filterYear) })
      } catch (e) {
        console.error('Failed to auto-sync siakad', e)
      }

      Swal.fire({
        toast: true,
        position: 'top-end',
        showConfirmButton: false,
        timer: 3000,
        icon: 'success',
        title: 'Draft slip gaji dan sinkronisasi SIAKAD berhasil'
      })
      loadPayrolls()
    } catch (e) {
      setLoading(false)
    }
  }

  const sendAllSlips = async () => {
    const result = await Swal.fire({
      title: 'Kirim Semua Slip?',
      html: `Sistem akan mengirimkan pesan WhatsApp ke <b>${payrolls.length}</b> asatidz.<br>Pastikan status WhatsApp Ready.`,
      icon: 'question',
      showCancelButton: true,
      confirmButtonText: 'Ya, Kirim Semua',
      cancelButtonText: 'Batal'
    })

    if (result.isConfirmed) {
      Swal.fire({
        title: 'Sedang Mengirim...',
        html: 'Mohon tunggu, proses ini mungkin memakan waktu beberapa saat.',
        allowOutsideClick: false,
        didOpen: () => Swal.showLoading()
      })

      try {
        const res = await api.post(`/payroll/bulk-send?month=${filterMonth}&year=${filterYear}`)
        Swal.fire('Berhasil', res.message, 'success')
      } catch (e) {
        // error handled
      }
    }
  }

  const syncSiakad = async () => {
    Swal.fire({ title: 'Menyinkronkan Siakad...', allowOutsideClick: false, didOpen: () => Swal.showLoading() })
    try {
      await api.post('/siakad/sync-all', { bulan: parseInt(filterMonth), tahun: parseInt(filterYear) })
      Swal.fire({ toast: true, position: 'top-end', icon: 'success', title: 'Data Siakad tersinkronisasi', showConfirmButton: false, timer: 1500 })
      loadPayrolls()
    } catch (e) {
      Swal.fire('Error', e.response?.data?.message || 'Gagal sinkronisasi Siakad', 'error')
    }
  }

  const confirmAll = async () => {
    const drafts = payrolls.filter(p => p.status === 'DRAFT')
    if (drafts.length === 0) {
      Swal.fire('Info', 'Tidak ada slip dengan status DRAFT', 'info')
      return
    }

    const result = await Swal.fire({
      title: 'Konfirmasi Semua?',
      text: `Ubah ${drafts.length} slip DRAFT menjadi TERKONFIRMASI?`,
      icon: 'question',
      showCancelButton: true,
      confirmButtonText: 'Ya, Konfirmasi'
    })

    if (result.isConfirmed) {
      Swal.fire({ title: 'Memproses...', allowOutsideClick: false, didOpen: () => Swal.showLoading() })
      try {
        for (let p of drafts) {
          await api.post(`/payroll/${p.id}/status`, { status: 'CONFIRMED' })
        }
        Swal.fire('Berhasil', 'Semua draft berhasil dikonfirmasi', 'success')
        loadPayrolls()
      } catch (e) {
        Swal.close()
      }
    }
  }

  return (
    <div className="card">
      <div className="card-header">
        <div className="filter-bar" style={{ marginBottom: 0 }}>
          <select 
            className="form-control" 
            style={{ width: '150px' }}
            value={filterMonth}
            onChange={(e) => setFilterMonth(e.target.value)}
          >
            {Array.from({ length: 12 }, (_, i) => i + 1).map(m => (
              <option key={m} value={m}>{getMonthName(m)}</option>
            ))}
          </select>
          <input 
            type="number" 
            className="form-control" 
            style={{ width: '100px' }}
            value={filterYear}
            onChange={(e) => setFilterYear(e.target.value)}
          />
          <button className="btn btn-primary" onClick={generatePayroll} disabled={loading}>
            <Plus size={18} /> Generate Draft Bulanan
          </button>
          <button className="btn btn-outline" style={{ display: 'flex', alignItems: 'center', gap: '8px' }} onClick={syncSiakad} disabled={loading}>
            <RefreshCw size={18} /> Tarik KBM Siakad
          </button>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button className="btn btn-outline" onClick={confirmAll} disabled={payrolls.length === 0}>
            <CheckCircle size={18} /> Konfirmasi Semua
          </button>
          <button className="btn btn-success" onClick={sendAllSlips} disabled={payrolls.length === 0} style={{ background: '#10B981', color: 'white', border: 'none' }}>
            <Share2 size={18} /> Blast WA
          </button>
        </div>
      </div>
      <div className="card-body p-0">
        <div className="table-responsive">
          <table className="table">
            <thead>
              <tr>
                <th>Nama Asatidz</th>
                <th>Pendapatan</th>
                <th>Potongan</th>
                <th>Take Home Pay</th>
                <th>Status WA</th>
                <th>Status Gaji</th>
                <th style={{ width: '120px' }}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan="7" style={{ textAlign: 'center' }}>Memuat data...</td></tr>
              ) : payrolls.length === 0 ? (
                <tr><td colSpan="7" style={{ textAlign: 'center' }}>Belum ada data payroll untuk bulan ini</td></tr>
              ) : (
                payrolls.map(p => {
                  const statusConf = getStatusConfig(p.status)
                  return (
                    <tr key={p.id}>
                      <td style={{ fontWeight: 600 }}>
                        {p.employee?.name}
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)', fontWeight: 400 }}>{p.employee?.category?.name}</div>
                      </td>
                      <td style={{ color: 'var(--success)' }}>{formatRupiah(p.gross_income)}</td>
                      <td style={{ color: 'var(--danger)' }}>{formatRupiah(p.total_deductions)}</td>
                      <td style={{ fontWeight: 'bold' }}>{formatRupiah(p.take_home_pay)}</td>
                      <td>
                        {p.wa_sent_at ? (
                          <span style={{ fontSize: '12px', color: 'var(--success)' }}>✅ Terkirim</span>
                        ) : (
                          <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Belum</span>
                        )}
                      </td>
                      <td>
                        <span className={`badge ${statusConf.cls}`}>
                          {statusConf.text}
                        </span>
                      </td>
                      <td>
                        <Link to={`/payroll/${p.id}`} className="btn-icon" title="Lihat Detail">
                          <Eye size={18} />
                        </Link>
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
