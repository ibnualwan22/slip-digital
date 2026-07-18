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

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
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
      </Route>
    </Routes>
  )
}

export default App
