import React, { useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Menu } from 'lucide-react'
import Sidebar from './Sidebar'

const routeTitles = {
  '/dashboard': 'Dashboard',
  '/categories': 'Master Kategori Asatidz',
  '/employees': 'Data Asatidz',
  '/activities': 'Master Aktivitas',
  '/payroll': 'Transaksi Payroll',
  '/siakad': 'Data SIAKAD Pengajar',
}

export default function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()

  let pageTitle = routeTitles[location.pathname] || 'Dashboard'
  if (location.pathname.startsWith('/payroll/')) {
    pageTitle = 'Detail Slip Gaji'
  }

  return (
    <div className="layout">
      <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      
      <div className="main-wrapper">
        <header className="header">
          <div className="header-left">
            <Menu 
              className="toggle-btn" 
              size={24} 
              onClick={() => setSidebarOpen(true)} 
            />
            <h1 className="page-title">{pageTitle}</h1>
          </div>
          <div className="header-right">
            <div className="user-profile">
              <div className="user-info">
                <span className="user-name">Administrator</span>
                <span className="user-role">Super Admin</span>
              </div>
              <img 
                src="https://ui-avatars.com/api/?name=Admin&background=EFF6FF&color=1E40AF" 
                alt="User" 
                className="user-avatar"
              />
            </div>
          </div>
        </header>

        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
