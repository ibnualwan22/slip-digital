import React, { useEffect, useState } from 'react'
import { Users, UserCheck, Receipt, Wallet, LayoutGrid } from 'lucide-react'
import api from '../api'
import { getMonthName, formatRupiah } from '../utils/formatter'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend
} from 'recharts'

export default function Dashboard() {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState({
    totalEmployees: 0,
    activeEmployees: 0,
    totalTransactions: 0,
    globalExpense: 0,
    currentMonth: new Date().getMonth() + 1,
    currentYear: new Date().getFullYear(),
    monthlyStats: []
  })

  useEffect(() => {
    fetchDashboardData()
  }, [])

  const fetchDashboardData = async () => {
    try {
      const now = new Date()
      const currentMonth = now.getMonth() + 1
      const currentYear = now.getFullYear()

      const res = await api.get(`/dashboard/stats?year=${currentYear}`)
      const stats = res.data || {}

      const monthlyStats = stats.monthly_stats || []

      const currentMonthStat = monthlyStats.find(s => s.month === currentMonth) || {}

      setData({
        totalEmployees: stats.total_employees || 0,
        activeEmployees: stats.active_employees || 0,
        totalTransactions: currentMonthStat.payroll_count || 0,
        globalExpense: currentMonthStat.global_expense || 0,
        currentMonth,
        currentYear,
        monthlyStats: monthlyStats.map(s => ({
          name: getMonthName(s.month).substring(0, 3),
          Payroll: s.payroll_expense,
          'Pengeluaran Bulanan': s.monthly_expense,
          'Pengeluaran Global': s.global_expense
        }))
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
            <h3>{formatRupiah(data.globalExpense)}</h3>
            <p>Pengeluaran Global</p>
          </div>
        </div>
      </div>

      <div className="card" style={{ marginBottom: '24px' }}>
        <div className="card-header">
          <h3 className="card-title">Grafik Pengeluaran Tahun {data.currentYear}</h3>
          <p style={{ fontSize: '0.85rem', color: '#6b7280', margin: '4px 0 0 0' }}>
            *Setiap pencatatan pengeluaran dan gaji didasarkan pada **periode (bulan) kerja** aslinya.
          </p>
        </div>
        <div className="card-body" style={{ height: '400px', padding: '24px 24px 0 0' }}>
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={data.monthlyStats} margin={{ top: 10, right: 30, left: 20, bottom: 20 }}>
              <defs>
                <linearGradient id="colorGlobal" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="colorPayroll" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#10b981" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="colorExpense" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#f59e0b" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e7eb" />
              <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fill: '#6b7280', fontSize: 13 }} dy={10} />
              <YAxis
                axisLine={false}
                tickLine={false}
                tick={{ fill: '#6b7280', fontSize: 13 }}
                tickFormatter={(value) => value >= 1000000 ? `Rp ${value / 1000000}Jt` : value >= 1000 ? `Rp ${value / 1000}rb` : `Rp ${value}`}
                dx={-10}
              />
              <Tooltip
                formatter={(value) => formatRupiah(value)}
                contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)', fontWeight: 500 }}
              />
              <Legend verticalAlign="top" height={40} iconType="circle" wrapperStyle={{ paddingBottom: '16px' }} />
              <Area type="monotone" dataKey="Pengeluaran Global" stroke="#3b82f6" strokeWidth={3} fillOpacity={1} fill="url(#colorGlobal)" />
              <Area type="monotone" dataKey="Payroll" stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorPayroll)" />
              <Area type="monotone" dataKey="Pengeluaran Bulanan" stroke="#f59e0b" strokeWidth={2} fillOpacity={1} fill="url(#colorExpense)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Informasi Sistem</h3>
        </div>
        <div className="card-body">
          <p>Selamat datang di sistem E-Maliyah Markaz Arabiyah. Gunakan menu di sebelah kiri untuk mengelola data Asatidz, master aktivitas, dan transaksi payroll bulanan. Sekarang dengan <strong>Integrasi SIAKAD otomatis</strong>.</p>
        </div>
      </div>
    </div>
  )
}
