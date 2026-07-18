import React from 'react'
import { NavLink } from 'react-router-dom'
import { LayoutGrid, Layers, Users, List, Receipt, Link2, X, ShoppingCart } from 'lucide-react'

export default function Sidebar({ isOpen, onClose }) {
  return (
    <aside className={`sidebar ${isOpen ? 'active' : ''}`}>
      <div className="sidebar-header">
        <img 
          src="/img/logo_markaz.png" 
          alt="Logo Markaz" 
          className="sidebar-logo"
          onError={(e) => {
            e.target.src = 'https://ui-avatars.com/api/?name=Markaz+Arabiyah&background=1E3A5F&color=fff&size=128'
          }}
        />
        <h2 className="sidebar-title">E-Rekap<br/><span>Markaz Arabiyah</span></h2>
        <X className="toggle-btn-mobile" size={24} onClick={onClose} />
      </div>

      <ul className="sidebar-menu">
        <li className="menu-label">MENU UTAMA</li>
        <li>
          <NavLink to="/dashboard" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <LayoutGrid size={20} />
            <span>Dashboard</span>
          </NavLink>
        </li>
        <li>
          <NavLink to="/categories" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <Layers size={20} />
            <span>Kategori Asatidz</span>
          </NavLink>
        </li>
        <li>
          <NavLink to="/employees" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <Users size={20} />
            <span>Asatidz</span>
          </NavLink>
        </li>
        <li>
          <NavLink to="/activities" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <List size={20} />
            <span>Aktivitas</span>
          </NavLink>
        </li>

        <li className="menu-label">KEUANGAN</li>
        <li>
          <NavLink to="/payroll" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <Receipt size={20} />
            <span>Transaksi Payroll</span>
          </NavLink>
        </li>
        <li>
          <NavLink to="/expenses" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <ShoppingCart size={20} />
            <span>Pengeluaran</span>
          </NavLink>
        </li>
        
        <li className="menu-label">INTEGRASI</li>
        <li>
          <NavLink to="/siakad" onClick={onClose} className={({ isActive }) => `menu-item ${isActive ? 'active' : ''}`}>
            <Link2 size={20} />
            <span>Data SIAKAD</span>
          </NavLink>
        </li>
      </ul>
    </aside>
  )
}
