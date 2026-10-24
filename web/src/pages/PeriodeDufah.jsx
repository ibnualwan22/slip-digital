import React, { useEffect, useState } from 'react'
import api from '../api'

const BULAN = ['', 'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
const emptyForm = { dufah: '', month: new Date().getMonth() + 1, year: new Date().getFullYear(), start_date: '', end_date: '' }

// Rentang penuh satu dufah: tanggal mulai usbu paling awal sampai
// tanggal selesai usbu paling akhir (satu dufah berisi 3 usbu).
function dufahSpan(d) {
  if (!d) return { start: '', end: '' }
  const starts = d.usbu.filter((u) => u.mulai).map((u) => u.mulai).sort()
  const ends = d.usbu.filter((u) => u.selesai).map((u) => u.selesai).sort()
  return {
    start: starts.length ? starts[0] : '',
    end: ends.length ? ends[ends.length - 1] : '',
  }
}

export default function PeriodeDufah() {
  const [periods, setPeriods] = useState([])
  const [dufahList, setDufahList] = useState([])
  const [dufahError, setDufahError] = useState('')
  const [form, setForm] = useState(emptyForm)
  const [editingId, setEditingId] = useState(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  const load = async () => {
    try {
      const res = await api.get('/periods')
      setPeriods(res.data || [])
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    load()
    api.get('/siakad/dufah')
      .then((res) => setDufahList(res.data || []))
      .catch((err) => setDufahError(`Gagal memuat daftar dufah dari Siakad: ${err.message}`))
  }, [])

  const selectedDufah = dufahList.find((d) => d.nama === form.dufah)
  const span = dufahSpan(selectedDufah)
  const hasDates = !!(span.start && span.end)

  const pickDufah = (nama) => {
    const d = dufahList.find((x) => x.nama === nama)
    const s = dufahSpan(d)
    setForm({ ...form, dufah: nama, start_date: s.start, end_date: s.end })
  }

  const submit = async (e) => {
    e.preventDefault()
    setError('')
    if (!form.dufah || !hasDates) {
      setError('Pilih dufah yang tanggal usbunya sudah terisi di Siakad.')
      return
    }
    setSaving(true)
    try {
      const payload = {
        dufah_name: form.dufah,
        usbu: 0,
        month: Number(form.month),
        year: Number(form.year),
        start_date: form.start_date,
        end_date: form.end_date,
      }
      if (editingId) {
        await api.put(`/periods/${editingId}`, payload)
      } else {
        await api.post('/periods', payload)
      }
      setForm(emptyForm)
      setEditingId(null)
      load()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  const edit = (p) => {
    setEditingId(p.id)
    setForm({
      dufah: p.dufah_name,
      month: p.month,
      year: p.year,
      start_date: p.start_date,
      end_date: p.end_date,
    })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const remove = async (p) => {
    if (!window.confirm(`Hapus pemetaan periode ${p.dufah_name} (${BULAN[p.month]} ${p.year})?`)) return
    try {
      await api.delete(`/periods/${p.id}`)
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  const input = { padding: '9px 11px', border: '1px solid #D1D5DB', borderRadius: 8, fontSize: 14, width: '100%', boxSizing: 'border-box', background: '#fff' }
  const inputLocked = { ...input, background: '#F3F4F6', color: '#374151' }
  const label = { display: 'block', fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 5 }

  return (
    <div style={{ padding: 24, maxWidth: 900 }}>
      <h1 style={{ fontSize: 20, color: '#1E3A5F', margin: '0 0 4px' }}>Pemetaan Periode Dufah</h1>
      <p style={{ fontSize: 13, color: '#6B7280', margin: '0 0 20px' }}>
        Atur di awal bulan gaji: satu periode adalah satu dufah (di dalamnya ada 3 usbu). Tanggal periode mengikuti
        rentang penuh dufah di Siakad — dari awal usbu pertama sampai akhir usbu terakhir — dan terkunci.
        Admin hanya menentukan bulan &amp; tahun gajinya. Rekap pengajar mengikuti pemetaan ini, dan rentang antarperiode tidak boleh tumpang tindih.
      </p>

      {dufahError && (
        <div style={{ background: '#FFFBEB', color: '#92400E', border: '1px solid #FDE68A', borderRadius: 10, padding: '10px 14px', fontSize: 14, marginBottom: 16 }}>{dufahError}</div>
      )}
      {error && (
        <div style={{ background: '#FEF2F2', color: '#B91C1C', border: '1px solid #FECACA', borderRadius: 10, padding: '10px 14px', fontSize: 14, marginBottom: 16 }}>{error}</div>
      )}

      <form onSubmit={submit} style={{ background: '#fff', border: '1px solid #E5E7EB', borderRadius: 14, padding: 20, marginBottom: 20 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(160px,1fr))', gap: 14 }}>
          <div>
            <label style={label}>Dufah (dari Siakad)</label>
            <select style={input} value={form.dufah} onChange={(e) => pickDufah(e.target.value)} required>
              <option value="">— Pilih dufah —</option>
              {dufahList.map((d) => <option key={d.nama} value={d.nama}>{d.nama}</option>)}
            </select>
            {selectedDufah && (
              <div style={{ fontSize: 12, color: '#6B7280', marginTop: 5, lineHeight: 1.6 }}>
                {selectedDufah.usbu.map((u) => (
                  <div key={u.ke}>Usbu {u.ke}: {u.mulai && u.selesai ? `${u.mulai} s.d. ${u.selesai}${u.aktif ? ' • aktif' : ''}` : 'belum diatur di Siakad'}</div>
                ))}
                {!hasDates && <div style={{ color: '#B91C1C' }}>Dufah ini belum memiliki tanggal usbu di Siakad.</div>}
              </div>
            )}
          </div>
          <div>
            <label style={label}>Bulan Gaji</label>
            <select style={input} value={form.month} onChange={(e) => setForm({ ...form, month: e.target.value })}>
              {BULAN.slice(1).map((b, i) => <option key={i + 1} value={i + 1}>{b}</option>)}
            </select>
          </div>
          <div>
            <label style={label}>Tahun</label>
            <input style={input} type="number" value={form.year} onChange={(e) => setForm({ ...form, year: e.target.value })} required />
          </div>
          <div>
            <label style={label}>Tanggal Mulai (dari Siakad)</label>
            <input style={inputLocked} type="date" value={form.start_date} readOnly tabIndex={-1} />
          </div>
          <div>
            <label style={label}>Tanggal Selesai (dari Siakad)</label>
            <input style={inputLocked} type="date" value={form.end_date} readOnly tabIndex={-1} />
          </div>
        </div>
        <div style={{ marginTop: 16, display: 'flex', gap: 10 }}>
          <button type="submit" disabled={saving} style={{ background: '#1E3A5F', color: '#fff', border: 'none', borderRadius: 8, padding: '10px 18px', fontWeight: 700, cursor: 'pointer' }}>
            {saving ? 'Menyimpan…' : editingId ? 'Simpan Perubahan' : 'Tambah Periode'}
          </button>
          {editingId && (
            <button type="button" onClick={() => { setEditingId(null); setForm(emptyForm) }} style={{ background: '#F3F4F6', border: '1px solid #E5E7EB', borderRadius: 8, padding: '10px 18px', fontWeight: 600, cursor: 'pointer' }}>Batal</button>
          )}
        </div>
      </form>

      <div style={{ background: '#fff', border: '1px solid #E5E7EB', borderRadius: 14, overflow: 'hidden' }}>
        {periods.length === 0 && <div style={{ padding: '16px 20px', color: '#6B7280', fontSize: 14 }}>Belum ada periode yang dipetakan.</div>}
        {periods.map((p) => (
          <div key={p.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '13px 20px', borderBottom: '1px solid #F3F4F6' }}>
            <div>
              <div style={{ fontWeight: 700, color: '#111827' }}>{p.dufah_name} — {BULAN[p.month]} {p.year}</div>
              <div style={{ fontSize: 12, color: '#6B7280' }}>Sesi {p.start_date} s.d. {p.end_date}</div>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={() => edit(p)} style={{ background: '#EFF6FF', color: '#2563EB', border: '1px solid #BFDBFE', borderRadius: 8, padding: '7px 13px', fontWeight: 600, cursor: 'pointer' }}>Ubah</button>
              <button onClick={() => remove(p)} style={{ background: '#FEF2F2', color: '#B91C1C', border: '1px solid #FECACA', borderRadius: 8, padding: '7px 13px', fontWeight: 600, cursor: 'pointer' }}>Hapus</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
