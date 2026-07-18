import React, { useEffect, useState } from 'react'
import { Users, UserCheck, Receipt, Wallet, LayoutGrid } from 'lucide-react'
import api from '../api'
import { getMonthName, formatRupiah } from '../utils/formatter'

export default function Dashboard() {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState({
    totalEmployees: 0,
    activeEmployees: 0,
    totalTransactions: 0,
    totalExpense: 0,
    currentMonth: new Date().getMonth() + 1,
    currentYear: new Date().getFullYear(),
  })

  useEffect(() => {
    fetchDashboardData()
  }, [])

  const fetchDashboardData = async () => {
    try {
      const empRes = await api.get('/employees')
      const employees = empRes.data || []
      
      const now = new Date()
      const currentMonth = now.getMonth() + 1
      const currentYear = now.getFullYear()

      const payRes = await api.get(`/payroll?month=${currentMonth}&year=${currentYear}`)
      const payrolls = payRes.data || []

      const totalActive = employees.filter(e => e.is_active).length
      
      let totalExpense = 0
      payrolls.forEach(p => {
        totalExpense += parseFloat(p.take_home_pay)
      })

      setData({
        totalEmployees: employees.length,
        activeEmployees: totalActive,
        totalTransactions: payrolls.length,
        totalExpense,
        currentMonth,
        currentYear,
      })
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="loading-state">
        <div className="spinner"><LayoutGrid size={40} /></div>
        <p>Memuat data dashboard...</p>
      </div>
    )
  }

  return (
    <div>
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon primary">
            <Users size={24} />
          </div>
          <div className="stat-info">
            <h3>{data.totalEmployees}</h3>
            <p>Total Asatidz</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon success">
            <UserCheck size={24} />
          </div>
          <div className="stat-info">
            <h3>{data.activeEmployees}</h3>
            <p>Asatidz Aktif</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon warning">
            <Receipt size={24} />
          </div>
          <div className="stat-info">
            <h3>{data.totalTransactions}</h3>
            <p>Slip Gaji ({getMonthName(data.currentMonth)} {data.currentYear})</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon danger">
            <Wallet size={24} />
          </div>
          <div className="stat-info">
            <h3>{formatRupiah(data.totalExpense)}</h3>
            <p>Estimasi Pengeluaran</p>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Informasi Sistem</h3>
        </div>
        <div className="card-body">
          <p>Selamat datang di sistem E-Rekap Markaz Arabiyah. Gunakan menu di sebelah kiri untuk mengelola data Asatidz, master aktivitas, dan transaksi payroll bulanan. Sekarang dengan <strong>Integrasi SIAKAD otomatis</strong>.</p>
        </div>
      </div>
    </div>
  )
}
