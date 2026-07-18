import React, { useState, useEffect } from 'react'
import { Plus, Edit, Trash2 } from 'lucide-react'
import api from '../api'
import { formatRupiah, getCategoryBadgeClass } from '../utils/formatter'
import Swal from 'sweetalert2'

export default function Employees() {
  const [employees, setEmployees] = useState([])
  const [categories, setCategories] = useState([])
  const [loading, setLoading] = useState(true)
  const [filterCat, setFilterCat] = useState('')
  const [filterActive, setFilterActive] = useState(true)

  useEffect(() => {
    loadCategories()
  }, [])

  useEffect(() => {
    loadEmployees()
  }, [filterCat, filterActive])

  const loadCategories = async () => {
    try {
      const res = await api.get('/categories')
      setCategories(res.data || [])
    } catch (e) {
      console.error(e)
    }
  }

  const loadEmployees = async () => {
    setLoading(true)
    try {
      let url = `/employees?active=${filterActive}`
      if (filterCat) url += `&category=${filterCat}`

      const res = await api.get(url)
      setEmployees(res.data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const showEmployeeForm = async (id = null) => {
    let emp = { name: '', category_id: '', role: '', is_active: true, phone_wa: '', structural_allowance: null, hourly_rate: null }
    let title = 'Tambah Asatidz Baru'

    if (id) {
      const existing = employees.find(e => e.id === id)
      if (existing) {
        emp = { ...existing }
        title = 'Edit Asatidz'
      }
    }

    // load siakad pengajar list
    let siakadOptions = '<option value="">-- Tidak Terhubung / Kosong --</option>'
    try {
      Swal.fire({ title: 'Memuat Form...', allowOutsideClick: false, didOpen: () => Swal.showLoading() })
      // Even though Phase 2 is backend, we can prepare the UI here. We will just comment out the fetch for now
      // or handle it gracefully if the endpoint doesn't exist yet.
      
      // Let's implement the siakad frontend part fully now in phase 1 as requested in phase 3, slightly out of order but more efficient for this file
      const res = await api.get('/siakad/pengajar').catch(() => ({ data: [] })) 
      const pengajarData = res.data || []
      
      siakadOptions += pengajarData.map(p => 
        `<option value="${p.id}" ${p.id === emp.siakad_id ? 'selected' : ''}>${p.nama}</option>`
      ).join('')
      Swal.close()
    } catch (e) { Swal.close() }

    const { value: formValues } = await Swal.fire({
      title: title,
      html: `
          <div class="form-group" style="background:#EFF6FF; padding:12px; border-radius:6px; border:1px solid #BFDBFE">
              <label class="form-label" style="color:#1E40AF">Tautkan ke SIAKAD (Opsional)</label>
              <select id="swal-siakad" class="form-control">
                  ${siakadOptions}
              </select>
          </div>
          <div class="form-group">
              <label class="form-label">Nama Asatidz</label>
              <input id="swal-name" class="form-control" value="${emp.name}">
          </div>
          <div class="form-row">
              <div class="form-group">
                  <label class="form-label">Kategori</label>
                  <select id="swal-cat" class="form-control">
                      <option value="">-- Pilih Kategori --</option>
                      ${categories.map(c => `<option value="${c.id}" ${c.id === emp.category_id ? 'selected' : ''}>${c.name}</option>`).join('')}
                  </select>
              </div>
              <div class="form-group">
                  <label class="form-label">Jabatan (Role)</label>
                  <input id="swal-role" class="form-control" value="${emp.role}">
              </div>
          </div>
          <div class="form-group">
              <label class="form-label">Nomor WhatsApp <span style="font-size:11px;color:var(--text-muted)">(Format: 0812... dsb)</span></label>
              <input type="text" id="swal-phone" class="form-control" value="${emp.phone_wa || ''}" placeholder="08123456789">
          </div>
          <div class="form-row">
              <div class="form-group">
                  <label class="form-label">Tunjangan Khusus <span style="font-size:11px;color:var(--text-muted)">(opsional)</span></label>
                  <input type="number" id="swal-allowance" class="form-control" value="${emp.structural_allowance !== null ? emp.structural_allowance : ''}" placeholder="Ikut kategori">
              </div>
              <div class="form-group">
                  <label class="form-label">Tarif Jam <span style="font-size:11px;color:var(--text-muted)">(opsional)</span></label>
                  <input type="number" id="swal-hourly" class="form-control" value="${emp.hourly_rate !== null ? emp.hourly_rate : ''}" placeholder="Ikut kategori">
              </div>
          </div>
          <div class="form-group" style="margin-top: 15px;">
              <label style="display:flex; align-items:center; gap:8px; justify-content:start;">
                  <input type="checkbox" id="swal-active" ${emp.is_active ? 'checked' : ''}> Status Aktif
              </label>
          </div>
      `,
      focusConfirm: false,
      showCancelButton: true,
      confirmButtonText: 'Simpan',
      cancelButtonText: 'Batal',
      confirmButtonColor: 'var(--primary-color)',
      preConfirm: () => {
        const catId = document.getElementById('swal-cat').value
        if (!catId) {
          Swal.showValidationMessage('Silakan pilih kategori')
          return false
        }
        
        const allowanceVal = document.getElementById('swal-allowance').value
        const hourlyVal = document.getElementById('swal-hourly').value
        
        return {
          name: document.getElementById('swal-name').value,
          category_id: catId,
          role: document.getElementById('swal-role').value,
          phone_wa: document.getElementById('swal-phone').value,
          structural_allowance: allowanceVal !== '' ? parseFloat(allowanceVal) : null,
          hourly_rate: hourlyVal !== '' ? parseFloat(hourlyVal) : null,
          is_active: document.getElementById('swal-active').checked,
          siakad_id: document.getElementById('swal-siakad').value || null
        }
      }
    })

    if (formValues) {
      try {
        if (id) {
          await api.put(`/employees/${id}`, formValues)
          Swal.fire('Tersimpan', 'Data Asatidz berhasil diupdate', 'success')
        } else {
          await api.post('/employees', formValues)
          Swal.fire('Tersimpan', 'Asatidz baru berhasil ditambahkan', 'success')
        }
        loadEmployees()
      } catch (e) {
        // error handled
      }
    }
  }

  const deleteEmployee = async (id) => {
    const result = await Swal.fire({
      title: 'Hapus Asatidz?',
      text: "Data yang dihapus tidak dapat dikembalikan!",
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: 'var(--danger)',
      cancelButtonColor: 'var(--text-muted)',
      confirmButtonText: 'Ya, Hapus!',
      cancelButtonText: 'Batal'
    })

    if (result.isConfirmed) {
      try {
        await api.delete(`/employees/${id}`)
        Swal.fire('Terhapus!', 'Data Asatidz berhasil dihapus.', 'success')
        loadEmployees()
      } catch (e) {
        // error handled
      }
    }
  }

  return (
    <div className="card">
      <div className="card-header">
        <div className="filter-bar" style={{ marginBottom: 0 }}>
          <select 
            className="form-control" 
            style={{ width: '200px' }}
            value={filterCat}
            onChange={(e) => setFilterCat(e.target.value)}
          >
            <option value="">Semua Kategori</option>
            {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
          <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
            <input 
              type="checkbox" 
              checked={filterActive}
              onChange={(e) => setFilterActive(e.target.checked)}
            /> 
            Aktif Saja
          </label>
        </div>
        <button className="btn btn-primary" onClick={() => showEmployeeForm()}>
          <Plus size={18} /> Tambah Asatidz
        </button>
      </div>
      <div className="card-body p-0">
        <div className="table-responsive">
          <table className="table">
            <thead>
              <tr>
                <th>Nama Asatidz & Jabatan</th>
                <th>Kategori</th>
                <th>Gaji Pokok</th>
                <th>Tunj. Khusus (Intensif)</th>
                <th>Target Insentif</th>
                <th>Rate Jam</th>
                <th>Status</th>
                <th style={{ width: '100px' }}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan="8" style={{ textAlign: 'center' }}>Memuat data...</td></tr>
              ) : employees.length === 0 ? (
                <tr><td colSpan="8" style={{ textAlign: 'center' }}>Tidak ada data Asatidz</td></tr>
              ) : (
                employees.map(emp => (
                  <tr key={emp.id}>
                    <td>
                      <div style={{ fontWeight: 600 }}>
                        {emp.name} {emp.siakad_id && <span title="Terhubung ke SIAKAD" style={{ color: '#1E40AF' }}>•</span>}
                      </div>
                      <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{emp.role || '-'}</div>
                    </td>
                    <td>
                      <span className={`badge ${getCategoryBadgeClass(emp.category?.code)}`}>
                        {emp.category?.name || '-'}
                      </span>
                    </td>
                    <td>{formatRupiah(emp.category?.fixed_salary)}</td>
                    <td>
                      {emp.structural_allowance !== null ? (
                        <span style={{ color: 'var(--primary-color)' }}>{formatRupiah(emp.structural_allowance)} (Override)</span>
                      ) : (
                        formatRupiah(emp.category?.structural_allowance)
                      )}
                    </td>
                    <td>{formatRupiah(emp.category?.target_incentive)}</td>
                    <td>
                      {emp.hourly_rate !== null ? (
                        <span style={{ color: 'var(--primary-color)' }}>{formatRupiah(emp.hourly_rate)} (Override)</span>
                      ) : (
                        formatRupiah(emp.category?.hourly_rate)
                      )}
                    </td>
                    <td>
                      {emp.is_active ? 
                        <span className="badge badge-success">Aktif</span> : 
                        <span className="badge badge-danger">Nonaktif</span>
                      }
                    </td>
                    <td>
                      <button className="btn-icon" onClick={() => showEmployeeForm(emp.id)} title="Edit">
                        <Edit size={18} />
                      </button>
                      <button className="btn-icon delete" onClick={() => deleteEmployee(emp.id)} title="Hapus">
                        <Trash2 size={18} />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
