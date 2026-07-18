import React, { useState, useEffect } from 'react'
import { Plus, Edit } from 'lucide-react'
import api from '../api'
import { formatRupiah } from '../utils/formatter'
import Swal from 'sweetalert2'

export default function Categories() {
  const [categories, setCategories] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadCategories()
  }, [])

  const loadCategories = async () => {
    setLoading(true)
    try {
      const res = await api.get('/categories')
      setCategories(res.data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const showCategoryForm = async (id = null) => {
    let cat = { name: '', code: '', calc_method: 'FIXED', fixed_salary: 0, structural_allowance: 0, target_incentive: 0, hourly_rate: 0 }
    let title = 'Tambah Kategori Baru'

    if (id) {
      const existing = categories.find(c => c.id === id)
      if (existing) {
        cat = { ...existing }
        title = 'Edit Kategori'
      }
    }

    const { value: formValues } = await Swal.fire({
      title: title,
      width: '600px',
      html: `
        <div class="form-row">
            <div class="form-group">
                <label class="form-label">Nama Kategori</label>
                <input id="swal-name" class="form-control" value="${cat.name}" placeholder="Misal: S2 Gelombang 1">
            </div>
            <div class="form-group">
                <label class="form-label">Kode (Unix/Singkat)</label>
                <input id="swal-code" class="form-control" value="${cat.code}" placeholder="S2_GEL1" ${id ? 'readonly' : ''}>
            </div>
        </div>
        <div class="form-group">
            <label class="form-label">Metode Kalkulasi (Jam Mengajar)</label>
            <select id="swal-calc" class="form-control">
                <option value="FIXED" ${cat.calc_method === 'FIXED' ? 'selected' : ''}>FIXED (Tanpa Rate)</option>
                <option value="HOURLY" ${cat.calc_method === 'HOURLY' ? 'selected' : ''}>HOURLY (Jam x Tarif Lokal)</option>
                <option value="PROPORTIONAL" ${cat.calc_method === 'PROPORTIONAL' ? 'selected' : ''}>PROPORTIONAL (Jam / 70 x Target)</option>
            </select>
        </div>
        <hr style="margin:20px 0; border:1px solid #E2E8F0">
        <div class="form-row">
            <div class="form-group">
                <label class="form-label">Gaji Pokok</label>
                <input type="number" id="swal-salary" class="form-control" value="${cat.fixed_salary}">
            </div>
            <div class="form-group">
                <label class="form-label">Tunjangan Struktural</label>
                <input type="number" id="swal-allowance" class="form-control" value="${cat.structural_allowance}">
            </div>
        </div>
        <div class="form-row">
            <div class="form-group">
                <label class="form-label">Tarif Jam (Hourly Rate)</label>
                <input type="number" id="swal-hourly" class="form-control" value="${cat.hourly_rate}">
            </div>
            <div class="form-group">
                <label class="form-label">Target Insentif</label>
                <input type="number" id="swal-incentive" class="form-control" value="${cat.target_incentive}">
            </div>
        </div>
      `,
      focusConfirm: false,
      showCancelButton: true,
      confirmButtonText: 'Simpan',
      cancelButtonText: 'Batal',
      confirmButtonColor: 'var(--primary-color)',
      preConfirm: () => {
        const name = document.getElementById('swal-name').value
        const code = document.getElementById('swal-code').value
        if (!name || !code) {
          Swal.showValidationMessage('Nama dan Kode tidak boleh kosong')
          return false
        }
        return {
          name,
          code,
          calc_method: document.getElementById('swal-calc').value,
          fixed_salary: parseFloat(document.getElementById('swal-salary').value) || 0,
          structural_allowance: parseFloat(document.getElementById('swal-allowance').value) || 0,
          hourly_rate: parseFloat(document.getElementById('swal-hourly').value) || 0,
          target_incentive: parseFloat(document.getElementById('swal-incentive').value) || 0
        }
      }
    })

    if (formValues) {
      try {
        if (id) {
          await api.put(`/categories/${id}`, formValues)
          Swal.fire('Tersimpan', 'Kategori berhasil diupdate', 'success')
        } else {
          await api.post('/categories', formValues)
          Swal.fire('Tersimpan', 'Kategori baru berhasil ditambahkan', 'success')
        }
        loadCategories()
      } catch (e) {
        // error handled in api.js
      }
    }
  }

  return (
    <div className="card">
      <div className="card-header">
        <h3 className="card-title">Master Kategori Asatidz</h3>
        <button className="btn btn-primary" onClick={() => showCategoryForm()}>
          <Plus size={18} /> Tambah Kategori
        </button>
      </div>
      <div className="card-body p-0">
        <div className="table-responsive">
          <table className="table">
            <thead>
              <tr>
                <th>Nama Kategori</th>
                <th>Kode</th>
                <th>Gaji Pokok</th>
                <th>Struktural</th>
                <th>Insentif / Rate</th>
                <th>Kalkulasi</th>
                <th style={{ width: '100px' }}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan="7" style={{ textAlign: 'center' }}>Memuat data...</td></tr>
              ) : categories.length === 0 ? (
                <tr><td colSpan="7" style={{ textAlign: 'center' }}>Tidak ada data kategori</td></tr>
              ) : (
                categories.map(cat => (
                  <tr key={cat.id}>
                    <td style={{ fontWeight: 600 }}>{cat.name}</td>
                    <td><span className="badge" style={{ background: 'var(--gray-200)', color: 'var(--text-color)' }}>{cat.code}</span></td>
                    <td>{formatRupiah(cat.fixed_salary)}</td>
                    <td>{formatRupiah(cat.structural_allowance)}</td>
                    <td>
                      {cat.target_incentive > 0 && <div>Target: {formatRupiah(cat.target_incentive)}</div>}
                      {cat.hourly_rate > 0 && <div>Rate: {formatRupiah(cat.hourly_rate)}</div>}
                      {cat.target_incentive === 0 && cat.hourly_rate === 0 && '-'}
                    </td>
                    <td>{cat.calc_method}</td>
                    <td>
                      <button className="btn-icon" onClick={() => showCategoryForm(cat.id)} title="Edit">
                        <Edit size={18} />
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
