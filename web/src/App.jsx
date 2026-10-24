import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Categories from './pages/Categories'
import Employees from './pages/Employees'
import Activities from './pages/Activities'
import Payroll from './pages/Payroll'
import PayrollDetail from './pages/PayrollDetail'
import Siakad from './pages/Siakad'
import ExpenseReport from './pages/ExpenseReport'
import ExpenseReportDetail from './pages/ExpenseReportDetail'
import Login from './pages/Login'
import RekapSaya from './pages/RekapSaya'
import RincianBisyaroh from './pages/RincianBisyaroh'
import PeriodeDufah from './pages/PeriodeDufah'
import { getToken, getRole, clearSession } from './api'

function RequireAuth({ children }) {
  if (!getToken()) return <Navigate to="/login" replace />
  return children
}

// Halaman-halaman lama KHUSUS admin. Pengajar yang mencoba masuk
// diarahkan ke Rekap Saya (dan API-nya pun menolak di server).
function RequireAdmin({ children }) {
  if (!getToken()) return <Navigate to="/login" replace />
  const role = getRole()
  if (role === 'pengajar') return <Navigate to="/rincian-bisyaroh" replace />
  if (role !== 'admin') {
    // Sesi lama tanpa peran (atau peran tak dikenal): jangan berputar
    // di antara dua pagar — bersihkan sesi dan kembali ke login.
    clearSession()
    return <Navigate to="/login" replace />
  }
  return children
}

function RequirePengajar({ children }) {
  if (!getToken()) return <Navigate to="/login" replace />
  const role = getRole()
  if (role === 'admin') return <Navigate to="/dashboard" replace />
  if (role !== 'pengajar') {
    clearSession()
    return <Navigate to="/login" replace />
  }
  return children
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/rekap-saya" element={<RequirePengajar><RekapSaya /></RequirePengajar>} />
      <Route path="/rincian-bisyaroh" element={<RequirePengajar><RincianBisyaroh /></RequirePengajar>} />
      <Route path="/" element={<RequireAdmin><Layout /></RequireAdmin>}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="categories" element={<Categories />} />
        <Route path="employees" element={<Employees />} />
        <Route path="activities" element={<Activities />} />
        <Route path="payroll" element={<Payroll />} />
        <Route path="payroll/:id" element={<PayrollDetail />} />
        <Route path="siakad" element={<Siakad />} />
        <Route path="expenses" element={<ExpenseReport />} />
        <Route path="expenses/:id" element={<ExpenseReportDetail />} />
        <Route path="periods" element={<PeriodeDufah />} />
      </Route>
    </Routes>
  )
}

export default App
