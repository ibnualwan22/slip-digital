import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, CheckCircle, Share2, RefreshCw, Pencil } from 'lucide-react'
import api from '../api'
import { formatRupiah, getMonthName, getStatusConfig, formatRupiahInput, parseRupiahInput } from '../utils/formatter'
import Swal from 'sweetalert2'

export default function PayrollDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  
  const [tx, setTx] = useState(null)
  const [details, setDetails] = useState([])
  const [activities, setActivities] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [id])

  const loadData = async () => {
    setLoading(true)
    try {
      const [txRes, detRes, actRes] = await Promise.all([
        api.get(`/payroll/${id}`),
        api.get(`/payroll/${id}/details`),
        api.get('/activities?active=true')
      ])
      setTx(txRes.data)
      setDetails(detRes.data || [])
      setActivities(actRes.data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const showDetailForm = async (type, existingDetail = null) => {
    try {
      // Filter activities by type (ADDITION or DEDUCTION)
    const filteredActivities = activities.filter(a => a.type === type)
    const typeLabel = type === 'ADDITION' ? 'Pemasukan' : 'Potongan'
    
    let actOptions = filteredActivities.map(a => 
      `<option value="${a.id}" data-type="${a.type}" data-rate="${a.default_rate}" ${existingDetail && existingDetail.activity_id === a.id ? 'selected' : ''}>
        [${a.type === 'ADDITION' ? '+' : '-'}] ${a.activity_name} (Rp ${Number(a.default_rate).toLocaleString('id-ID')})
      </option>`
    ).join('')
    if (existingDetail && !existingDetail.activity_id) {
       actOptions = `<option value="" selected>-- Manual --</option>` + actOptions
    }

    const { value: formValues } = await Swal.fire({
      title: `${existingDetail ? 'Edit' : 'Tambah'} ${typeLabel}`,
      html: `
          <div class="form-group" style="text-align: left;">
              <label class="form-label">Aktivitas</label>
              <select id="swal-det-activity" class="form-control">
                  <option value="">-- Pilih Aktivitas --</option>
                  ${actOptions}
              </select>
          </div>
          <div class="form-row">
              <div class="form-group" style="text-align: left;">
                  <label class="form-label">Qty / Jumlah (Misal: 4 pertemuan)</label>
                  <input type="number" id="swal-det-qty" class="form-control" value="${existingDetail ? existingDetail.quantity : 1}" min="1">
              </div>
              <div class="form-group" style="text-align: left;">
                  <label class="form-label">Rate / Harga Satuan (Rp)</label>
                  <input type="text" id="swal-det-rate" class="form-control" value="${existingDetail ? formatRupiahInput(existingDetail.rate.toString()) : ''}" placeholder="0">
              </div>
          </div>
          <div class="form-group" style="text-align: left;">
              <label class="form-label">Keterangan Opsional</label>
              <input type="text" id="swal-det-desc" class="form-control" value="${existingDetail ? (existingDetail.description || '') : ''}" placeholder="Misal: untuk kelas A">
          </div>
      `,
      didOpen: () => {
        const actSelect = document.getElementById('swal-det-activity')
        const rateInput = document.getElementById('swal-det-rate')
        
        // Auto-fill rate based on selected activity + format local string
        actSelect.addEventListener('change', (e) => {
          const opt = e.target.selectedOptions[0]
          if (opt && opt.value) {
            rateInput.value = formatRupiahInput(opt.dataset.rate)
          } else {
            rateInput.value = ''
          }
        })
        
        // format on type
        rateInput.addEventListener('input', (e) => {
          const val = e.target.value.replace(/[^0-9]/g, '')
          e.target.value = formatRupiahInput(val)
        })
      },
      focusConfirm: false,
      showCancelButton: true,
      confirmButtonText: 'Simpan',
      cancelButtonText: 'Batal',
      confirmButtonColor: 'var(--primary-color)',
      preConfirm: () => {
        const actId = document.getElementById('swal-det-activity').value
        if (!actId) {
          Swal.showValidationMessage('Silakan pilih aktivitas')
          return false
        }
        return {
          activity_id: actId,
          type: type,
          quantity: parseFloat(document.getElementById('swal-det-qty').value) || 1,
          rate: parseRupiahInput(document.getElementById('swal-det-rate').value),
          description: document.getElementById('swal-det-desc').value
        }
      }
    })

    if (formValues) {
      Swal.fire({ title: 'Menyimpan...', didOpen: () => Swal.showLoading() })
      try {
        if (existingDetail) {
          await api.put(`/payroll/details/${existingDetail.id}`, formValues)
        } else {
          await api.post(`/payroll/${id}/details`, formValues)
        }
        Swal.fire({ toast: true, position: 'top-end', icon: 'success', title: 'Berhasil disimpan', showConfirmButton: false, timer: 1500 })
        loadData()
      } catch (e) {
        Swal.fire('Error', e.response?.data?.message || 'Gagal menyimpan detail', 'error')
      }
    }
  } catch (e) {
    console.error(e)
    Swal.fire('Error UI', 'Gagal memuat form: ' + e.message, 'error')
  }
}

  const deleteDetail = async (detailId) => {
    const result = await Swal.fire({
      title: 'Hapus Item?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: 'var(--danger)'
    })

    if (result.isConfirmed) {
      try {
        await api.delete(`/payroll/details/${detailId}`)
        loadData()
      } catch (e) { }
    }
  }

  const updateStatus = async (status) => {
    try {
      await api.post(`/payroll/${id}/status`, { status })
      loadData()
    } catch (e) { }
  }

  const previewWA = async () => {
    try {
      Swal.fire({ title: 'Loading preview...', didOpen: () => Swal.showLoading() })
      const res = await api.get(`/payroll/${id}/wa-preview`)
      const base64Str = res.image_base64 || res.data?.image_base64 || (res.data ? res.data.data?.image_base64 : '')
      
      Swal.fire({
        title: 'Preview WhatsApp',
        html: `<div style="text-align: center; background: #F8FAFC; padding: 16px; border-radius: 8px; border: 1px solid #E2E8F0">
                 <img src="data:image/png;base64,${base64Str}" style="max-width: 100%; border-radius: 4px; box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);" />
               </div>`,
        width: '600px',
        showCancelButton: true,
        confirmButtonText: 'Kirim via WA API',
        cancelButtonText: 'Tutup',
        confirmButtonColor: '#10B981'
      }).then(async (result) => {
        if (result.isConfirmed) {
          Swal.fire({ title: 'Mengirim...', didOpen: () => Swal.showLoading() })
          await api.post(`/payroll/${id}/send-wa`)
          Swal.fire('Sukses', 'Pesan WhatsApp berhasil diproses', 'success')
        }
      })
    } catch (e) {
      Swal.fire('Error', e.response?.data?.message || 'Gagal menyiapkan preview', 'error')
    }
  }

  if (loading && !tx) {
    return <div className="loading-state"><div className="spinner"><RefreshCw size={40} /></div><p>Memuat data slip gaji...</p></div>
  }

  const additions = details.filter(d => d.type === 'ADDITION')
  const deductions = details.filter(d => d.type === 'DEDUCTION')

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '24px' }}>
        <button className="btn btn-outline" style={{ display: 'flex', alignItems: 'center', gap: '8px' }} onClick={() => navigate('/payroll')}>
          <ArrowLeft size={18} /> Kembali
        </button>
        <h2 style={{ fontSize: '20px', margin: 0 }}>Detail Slip Gaji — {tx?.employee?.name}</h2>
      </div>

      <div className="info-grid">
        <div className="info-item">
          <div className="label">Periode</div>
          <div className="value">{getMonthName(tx?.month)} {tx?.year}</div>
        </div>
        <div className="info-item">
          <div className="label">Jabatan (Role)</div>
          <div className="value">{tx?.employee?.role || 'Asatidz'}</div>
        </div>
        <div className="info-item">
          <div className="label">Status Draft</div>
          <div className="value">
            <span className={`badge ${getStatusConfig(tx?.status).cls}`}>
              {getStatusConfig(tx?.status).text}
            </span>
          </div>
        </div>
        <div className="info-item" style={{ background: 'var(--primary-light)', padding: '12px', borderRadius: '6px' }}>
          <div className="label" style={{ color: 'var(--primary-color)' }}>TAKE HOME PAY</div>
          <div className="value" style={{ fontSize: '20px', color: 'var(--primary-color)' }}>{formatRupiah(tx?.take_home_pay)}</div>
        </div>
      </div>

      <div style={{ display: 'flex', gap: '12px', marginBottom: '24px' }}>
        <button className="btn btn-primary" style={{ background: 'var(--success)', borderColor: 'var(--success)' }} onClick={() => showDetailForm('ADDITION')} disabled={tx?.status !== 'DRAFT'}>
          <Plus size={18} /> Tambah Pemasukan
        </button>
        <button className="btn btn-primary" style={{ background: 'var(--danger)', borderColor: 'var(--danger)' }} onClick={() => showDetailForm('DEDUCTION')} disabled={tx?.status !== 'DRAFT'}>
          <Plus size={18} /> Tambah Potongan
        </button>
        <button className="btn btn-success" style={{ background: '#10B981', color: 'white', border: 'none' }} onClick={previewWA}>
          <Share2 size={18} /> Preview & Kirim WA
        </button>
        
        <div style={{ marginLeft: 'auto', display: 'flex', gap: '8px' }}>
          {tx?.status === 'DRAFT' && (
            <button className="btn btn-outline" style={{ color: 'var(--primary-color)', borderColor: 'var(--primary-color)' }} onClick={() => updateStatus('CONFIRMED')}>
              <CheckCircle size={18} /> Konfirmasi
            </button>
          )}
          {tx?.status === 'CONFIRMED' && (
            <>
              <button className="btn btn-outline" onClick={() => updateStatus('DRAFT')}>Revisi ke Draft</button>
              <button className="btn btn-outline" style={{ color: 'var(--success)', borderColor: 'var(--success)' }} onClick={() => updateStatus('PAID')}>Set Telah Dibayar</button>
            </>
          )}
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
        {/* Kolom Pemasukan */}
        <div className="card">
          <div className="card-header" style={{ background: '#F0FDF4', color: 'var(--success)' }}>
            <h3 className="card-title">PENAMBAHAN (INCOME)</h3>
          </div>
          <div className="card-body p-0">
            <table className="table">
              <thead>
                <tr>
                  <th>Deskripsi</th>
                  <th style={{ textAlign: 'right' }}>Qty</th>
                  <th style={{ textAlign: 'right' }}>Total</th>
                  <th style={{ width: '60px' }}></th>
                </tr>
              </thead>
              <tbody>
                {additions.length === 0 ? (
                  <tr><td colSpan="4" style={{ textAlign: 'center', padding: '20px' }}>Tidak ada data penambahan</td></tr>
                ) : (
                  additions.map(d => (
                    <tr key={d.id}>
                      <td>
                        <strong>{d.activity?.activity_name}</strong>
                        {d.description && <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{d.description}</div>}
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>@ {formatRupiah(d.rate)}</div>
                      </td>
                      <td style={{ textAlign: 'right' }}>{d.quantity}x</td>
                      <td style={{ textAlign: 'right', fontWeight: 600 }}>{formatRupiah(d.total_amount)}</td>
                      <td>
                         {tx?.status === 'DRAFT' && d.description !== 'Gaji Pokok' && d.description !== 'Tunjangan Struktural' && (
                           <>
                             <button className="btn-icon edit" style={{ marginRight: '8px', color: '#3b82f6', background: 'transparent', border: 'none', cursor: 'pointer' }} onClick={() => showDetailForm('ADDITION', d)}><Pencil size={16} /></button>
                             <button className="btn-icon delete" onClick={() => deleteDetail(d.id)}><Trash2 size={16} /></button>
                           </>
                         )}
                      </td>
                    </tr>
                  ))
                )}
                <tr style={{ background: '#F8FAFC' }}>
                  <td colSpan="2" style={{ fontWeight: 600, textAlign: 'right' }}>Total Pemasukan:</td>
                  <td style={{ fontWeight: 'bold', color: 'var(--success)', textAlign: 'right' }}>{formatRupiah(tx?.gross_income)}</td>
                  <td></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        {/* Kolom Potongan */}
        <div className="card">
          <div className="card-header" style={{ background: '#FEF2F2', color: 'var(--danger)' }}>
            <h3 className="card-title">POTONGAN (DEDUCTION)</h3>
          </div>
          <div className="card-body p-0">
            <table className="table">
              <thead>
                <tr>
                  <th>Deskripsi</th>
                  <th style={{ textAlign: 'right' }}>Qty</th>
                  <th style={{ textAlign: 'right' }}>Total</th>
                  <th style={{ width: '60px' }}></th>
                </tr>
              </thead>
              <tbody>
                {deductions.length === 0 ? (
                  <tr><td colSpan="4" style={{ textAlign: 'center', padding: '20px' }}>Tidak ada data potongan</td></tr>
                ) : (
                  deductions.map(d => (
                    <tr key={d.id}>
                      <td>
                        <strong>{d.activity?.activity_name}</strong>
                        {d.description && <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{d.description}</div>}
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>@ {formatRupiah(d.rate)}</div>
                      </td>
                      <td style={{ textAlign: 'right' }}>{d.quantity}x</td>
                      <td style={{ textAlign: 'right', fontWeight: 600 }}>{formatRupiah(d.total_amount)}</td>
                      <td>
                         {tx?.status === 'DRAFT' && (
                           <>
                             <button className="btn-icon edit" style={{ marginRight: '8px', color: '#3b82f6', background: 'transparent', border: 'none', cursor: 'pointer' }} onClick={() => showDetailForm('DEDUCTION', d)}><Pencil size={16} /></button>
                             <button className="btn-icon delete" onClick={() => deleteDetail(d.id)}><Trash2 size={16} /></button>
                           </>
                         )}
                      </td>
                    </tr>
                  ))
                )}
                <tr style={{ background: '#F8FAFC' }}>
                  <td colSpan="2" style={{ fontWeight: 600, textAlign: 'right' }}>Total Potongan:</td>
                  <td style={{ fontWeight: 'bold', color: 'var(--danger)', textAlign: 'right' }}>-{formatRupiah(tx?.total_deductions)}</td>
                  <td></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
