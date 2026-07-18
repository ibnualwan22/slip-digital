import React, { useState, useEffect } from 'react'
import { Plus, Edit } from 'lucide-react'
import api from '../api'
import { formatRupiah } from '../utils/formatter'
import Swal from 'sweetalert2'

export default function Activities() {
  const [activities, setActivities] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadActivities()
  }, [])

  const loadActivities = async () => {
    setLoading(true)
    try {
      const res = await api.get('/activities')
      setActivities(res.data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const showActivityForm = async (id = null) => {
    let act = { activity_name: '', code: '', type: 'ADDITION', default_rate: 0, description: '', is_active: true }
    let title = 'Tambah Master Aktivitas'

    if (id) {
      const existing = activities.find(a => a.id === id)
      if (existing) {
        act = { ...existing }
        title = 'Edit Master Aktivitas'
      }
    }

    const { value: formValues } = await Swal.fire({
      title: title,
      html: `
          <div class="form-group">
              <label class="form-label">Nama Aktivitas</label>
              <input id="swal-act-name" class="form-control" placeholder="Cth: Jam Mengajar" value="${act.activity_name}">
          </div>
          <div class="form-row">
              <div class="form-group">
                  <label class="form-label">Kode (Unik)</label>
                  <input id="swal-act-code" class="form-control" placeholder="Cth: JAM_AJAR" style="text-transform:uppercase" value="${act.code}" ${id ? 'readonly' : ''}>
              </div>
              <div class="form-group">
                  <label class="form-label">Tipe</label>
                  <select id="swal-act-type" class="form-control">
                      <option value="ADDITION" ${act.type === 'ADDITION' ? 'selected' : ''}>Penambahan (Income)</option>
                      <option value="DEDUCTION" ${act.type === 'DEDUCTION' ? 'selected' : ''}>Potongan (Deduction)</option>
                  </select>
              </div>
          </div>
          <div class="form-group">
              <label class="form-label">Rate Default (Rp)</label>
              <input type="number" id="swal-act-rate" class="form-control" value="${act.default_rate}">
          </div>
          <div class="form-group">
              <label class="form-label">Deskripsi</label>
              <input id="swal-act-desc" class="form-control" value="${act.description || ''}">
          </div>
          <div class="form-group">
              <label style="display:flex; align-items:center; gap:8px; justify-content:start;">
                  <input type="checkbox" id="swal-act-active" ${act.is_active ? 'checked' : ''}> Status Aktif
              </label>
          </div>
      `,
      focusConfirm: false,
      showCancelButton: true,
      confirmButtonText: 'Simpan',
      cancelButtonText: 'Batal',
      confirmButtonColor: 'var(--primary-color)',
      preConfirm: () => {
        return {
          activity_name: document.getElementById('swal-act-name').value,
          code: document.getElementById('swal-act-code').value.toUpperCase(),
          type: document.getElementById('swal-act-type').value,
          default_rate: parseFloat(document.getElementById('swal-act-rate').value) || 0,
          description: document.getElementById('swal-act-desc').value,
          is_active: document.getElementById('swal-act-active').checked
        }
      }
    })

    if (formValues) {
      try {
        if (id) {
          await api.put(`/activities/${id}`, formValues)
          Swal.fire('Tersimpan', 'Aktivitas berhasil diupdate', 'success')
        } else {
          await api.post('/activities', formValues)
          Swal.fire('Tersimpan', 'Aktivitas baru berhasil ditambahkan', 'success')
        }
        loadActivities()
      } catch (e) {
        // handled
      }
    }
  }

  return (
    <div className="card">
      <div className="card-header">
        <h3 className="card-title">Master Aktivitas</h3>
        <button className="btn btn-primary" onClick={() => showActivityForm()}>
          <Plus size={18} /> Tambah Aktivitas
        </button>
      </div>
      <div className="card-body p-0">
        <div className="table-responsive">
          <table className="table">
            <thead>
              <tr>
                <th>Nama Aktivitas</th>
                <th>Kode</th>
                <th>Tipe</th>
                <th>Rate Default</th>
                <th>Deskripsi</th>
                <th>Status</th>
                <th style={{ width: '80px' }}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan="7" style={{ textAlign: 'center' }}>Memuat data...</td></tr>
              ) : activities.length === 0 ? (
                <tr><td colSpan="7" style={{ textAlign: 'center' }}>Tidak ada aktivitas</td></tr>
              ) : (
                activities.map(a => (
                  <tr key={a.id}>
                    <td style={{ fontWeight: 600 }}>{a.activity_name}</td>
                    <td>
                      <span style={{ fontFamily: 'monospace', background: '#F1F5F9', padding: '2px 6px', borderRadius: '4px' }}>
                        {a.code}
                      </span>
                    </td>
                    <td>
                      {a.type === 'ADDITION' ? (
                        <span className="badge badge-success">Penambahan</span>
                      ) : (
                        <span className="badge badge-danger">Potongan</span>
                      )}
                    </td>
                    <td>{formatRupiah(a.default_rate)}</td>
                    <td>{a.description || '-'}</td>
                    <td>
                      {a.is_active ? 
                        <span className="badge badge-success">Aktif</span> : 
                        <span className="badge badge-danger">Nonaktif</span>
                      }
                    </td>
                    <td>
                      <button className="btn-icon" onClick={() => showActivityForm(a.id)} title="Edit">
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
