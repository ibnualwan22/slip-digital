import React, { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft, Plus, Trash2, Camera, CheckSquare, Square,
  Loader, Upload, Eye, X, Check, AlertCircle, Save
} from 'lucide-react'
import Swal from 'sweetalert2'

const API_BASE = '/api/v1'

const MONTHS = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
]

function formatRp(num) {
  const n = parseFloat(num) || 0
  return 'Rp ' + n.toLocaleString('id-ID')
}

function fmtNum(num) {
  return parseFloat(num) || 0
}

const EMPTY_ITEM = {
  nama_barang: '',
  harga_satuan: '',
  jumlah: 1,
  kredit: '',
  debit: '',
}

// Compress image using Canvas before sending to AI — reduces payload 5-10x for speed
async function compressImage(file, maxWidth = 1024, quality = 0.75) {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => {
      const canvas = document.createElement('canvas')
      const scale = Math.min(1, maxWidth / img.width)
      canvas.width = img.width * scale
      canvas.height = img.height * scale
      canvas.getContext('2d').drawImage(img, 0, 0, canvas.width, canvas.height)
      canvas.toBlob(resolve, 'image/jpeg', quality)
    }
    img.src = URL.createObjectURL(file)
  })
}

export default function ExpenseReportDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const fileInputRef = useRef(null)
  const scanFileRef = useRef(null)

  const [report, setReport] = useState(null)
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  // Editing row state
  const [editing, setEditing] = useState(null) // { idx, field }

  // New item group form (manual entry)
  const [manualItems, setManualItems] = useState([ { ...EMPTY_ITEM } ])
  const [manualKredit, setManualKredit] = useState('')
  const [manualBuktiFile, setManualBuktiFile] = useState(null)
  const [manualBuktiPreview, setManualBuktiPreview] = useState(null)
  const [addingItem, setAddingItem] = useState(false)

  // AI Scan modal
  const [scanModal, setScanModal] = useState(false)
  const [scanning, setScanning] = useState(false)
  const [scanTimedOut, setScanTimedOut] = useState(false)
  const [scannedItems, setScannedItems] = useState([])
  const [scanPreview, setScanPreview] = useState(null)
  const [scanFileRaw, setScanFileRaw] = useState(null)
  const [scanKredit, setScanKredit] = useState(0)

  // Image preview modal
  const [previewImg, setPreviewImg] = useState(null)

  // Upload state per item
  const [uploadingIdx, setUploadingIdx] = useState(null)
  const [uploadingNew, setUploadingNew] = useState(false)

  const fetchReport = async () => {
    setLoading(true)
    try {
      const res = await fetch(`${API_BASE}/expenses/${id}`)
      const data = await res.json()
      if (!data.success) throw new Error(data.message)
      setReport(data.data)
      setItems(data.data.items || [])
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchReport() }, [id])

  // ----- Item actions -----

  const handleSaveManualGroup = async (e) => {
    e.preventDefault()
    // validasi
    const validItems = manualItems.filter(it => it.nama_barang.trim() !== '')
    if (validItems.length === 0) return

    setAddingItem(true)
    try {
      const groupId = Date.now().toString()
      let totalHargaBarang = validItems.reduce((acc, it) => acc + ((parseFloat(it.harga_satuan)||0) * (parseInt(it.jumlah)||1)), 0)
      let overallKredit = parseFloat(manualKredit) || 0
      let overallDebit = overallKredit > 0 ? (overallKredit - totalHargaBarang) : 0

      // Upload bukti first
      let uploadedUrl = ''
      if (manualBuktiFile) {
        const formData = new FormData()
        formData.append('file', manualBuktiFile)
        const upRes = await fetch(`${API_BASE}/expenses/upload`, { method: 'POST', body: formData })
        const upData = await upRes.json()
        if (upData.success && upData.data?.url) {
          uploadedUrl = upData.data.url
        }
      }

      // Save each item
      for (let i = 0; i < validItems.length; i++) {
        const item = validItems[i]
        const payload = {
          nama_barang: item.nama_barang,
          harga_satuan: parseFloat(item.harga_satuan) || 0,
          jumlah: parseInt(item.jumlah) || 1,
          kredit: i === 0 ? overallKredit : 0,
          debit: i === 0 ? overallDebit : 0,
          group_id: groupId,
          bukti_pembayaran: uploadedUrl
        }
        await fetch(`${API_BASE}/expenses/${id}/items`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        })
      }
      
      setManualItems([{...EMPTY_ITEM}])
      setManualKredit('')
      setManualBuktiFile(null)
      setManualBuktiPreview(null)
      fetchReport()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    } finally {
      setAddingItem(false)
    }
  }

  const handleManualItemChange = (idx, field, val) => {
    setManualItems(prev => {
      const nw = [...prev]
      nw[idx] = { ...nw[idx], [field]: val }
      return nw
    })
  }

  const handleUpdateItem = async (item) => {
    try {
      const res = await fetch(`${API_BASE}/expenses/items/${item.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nama_barang: item.nama_barang,
          harga_satuan: parseFloat(item.harga_satuan) || 0,
          jumlah: parseInt(item.jumlah) || 1,
          kredit: parseFloat(item.kredit) || 0,
          debit: parseFloat(item.debit) || 0,
          bukti_pembayaran: item.bukti_pembayaran || '',
          is_konfirmasi: item.is_konfirmasi,
          is_tashih: item.is_tashih,
        }),
      })
      const data = await res.json()
      if (!data.success) throw new Error(data.message)
      fetchReport()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    }
  }

  const handleToggle = async (item, field) => {
    const updated = { ...item, [field]: !item[field] }
    setItems(prev => prev.map(it => it.id === item.id ? updated : it))
    await handleUpdateItem(updated)
  }

  const handleToggleGroup = async (itemGroupId, field, currentVal) => {
    const newVal = !currentVal
    const groupItems = items.filter(it => (it.group_id || it.id) === itemGroupId)
    setItems(prev => prev.map(it => {
      if ((it.group_id || it.id) === itemGroupId) return { ...it, [field]: newVal }
      return it
    }))
    try {
      await Promise.all(groupItems.map(item => {
        const payload = { ...item, [field]: newVal }
        return fetch(`${API_BASE}/expenses/items/${item.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        })
      }))
      fetchReport()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    }
  }

  const formatInputAngka = (val) => {
    if (val === null || val === undefined || val === '') return ''
    const num = val.toString().replace(/\D/g, '')
    if (!num) return ''
    return parseInt(num, 10).toLocaleString('id-ID')
  }

  const parseInputAngka = (val) => {
    if (!val) return 0
    if (typeof val === 'number') return val
    const num = val.toString().replace(/\D/g, '')
    return parseInt(num, 10) || 0
  }

  const handleDelete = async (itemId) => {
    const result = await Swal.fire({
      title: 'Hapus Item?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonText: 'Hapus',
      cancelButtonText: 'Batal',
      confirmButtonColor: '#EF4444',
    })
    if (!result.isConfirmed) return
    try {
      const res = await fetch(`${API_BASE}/expenses/items/${itemId}`, { method: 'DELETE' })
      const data = await res.json()
      if (!data.success) throw new Error(data.message)
      fetchReport()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    }
  }

  // Inline edit: save on blur / enter
  const startEdit = (idx, field) => setEditing({ idx, field })

  const handleCellChange = (idx, field, value) => {
    setItems(prev => {
      const updated = [...prev]
      updated[idx] = { ...updated[idx], [field]: value }
      // Auto-calc total_harga
      if (field === 'harga_satuan' || field === 'jumlah') {
        const hs = parseFloat(field === 'harga_satuan' ? value : updated[idx].harga_satuan) || 0
        const jml = parseInt(field === 'jumlah' ? value : updated[idx].jumlah) || 1
        updated[idx].total_harga = hs * jml
      }
      return updated
    })
  }

  const saveEdit = async (idx) => {
    await handleUpdateItem(items[idx])
    setEditing(null)
  }

  // ----- Cloudinary Upload per item -----
  const handleUploadBukti = async (itemId, file) => {
    const formData = new FormData()
    formData.append('file', file)
    try {
      const res = await fetch(`${API_BASE}/expenses/upload`, { method: 'POST', body: formData })
      const data = await res.json()
      if (!data.success) throw new Error(data.message)
      // Update item with URL
      const item = items.find(it => it.id === itemId)
      if (item) {
        await handleUpdateItem({ ...item, bukti_pembayaran: data.data.url })
      }
    } catch (err) {
      Swal.fire('Error', 'Upload gagal: ' + err.message, 'error')
    }
  }

  // ----- AI Scan -----
  const handleScanFile = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    setScanning(true)
    setScanTimedOut(false)
    setScanPreview(URL.createObjectURL(file))
    setScanFileRaw(file)

    try {
      // 1. Compress image terlebih dahulu (max 1024px, quality 0.75) — jauh lebih kecil & cepat
      const compressed = await compressImage(file, 1024, 0.75)

      // 2. Convert ke base64
      const base64 = await new Promise((resolve) => {
        const reader = new FileReader()
        reader.onloadend = () => resolve(reader.result.split(',')[1])
        reader.readAsDataURL(compressed)
      })

      // 3. Kirim ke Gemini dengan timeout 30 detik
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 30000)
      let res
      try {
        res = await fetch(`${API_BASE}/expenses/scan-receipt`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ image_base64: base64, mime_type: 'image/jpeg' }),
          signal: controller.signal,
        })
      } finally {
        clearTimeout(timeout)
      }

      const data = await res.json()
      if (!data.success) throw new Error(data.message)

      const parsed = (data.data.items || []).map(it => ({
        nama_barang: it.nama_barang || '',
        harga_satuan: it.harga_satuan || 0,
        jumlah: it.jumlah || 1,
        // Remove individual kredit/debit mapping here, we use scanKredit
      }))

      // The AI prompt was told to put global kredit in item 0 (kredit).
      // Let's extract it and put it in scanKredit if it exists.
      const initialKredit = (data.data.items || [])[0]?.kredit || 0
      setScanKredit(initialKredit)

      // Jika AI tidak mendeteksi item apapun, langsung ke input manual
      if (parsed.length === 0) {
        setScanning(false)
        setScanTimedOut(true)
        setScannedItems([{ ...EMPTY_ITEM }])
        return
      }

      setScannedItems(parsed)
      setScanning(false)
    } catch (err) {
      setScanning(false)
      if (err.name === 'AbortError') {
        // Timeout — tampilkan baris kosong untuk isi manual
        setScanTimedOut(true)
        setScannedItems([{ ...EMPTY_ITEM }])
      } else {
        Swal.fire('Scan Gagal', err.message || 'AI tidak bisa membaca nota.', 'warning')
        setScanTimedOut(true)
        setScannedItems([{ ...EMPTY_ITEM }])
      }
    }
    e.target.value = ''
  }

  const handleConfirmScan = async () => {
    setSaving(true)
    try {
      // 1. Upload the scan photo automatically if it exists
      let uploadedUrl = ''
      if (scanFileRaw) {
        const formData = new FormData()
        formData.append('file', scanFileRaw)
        const upRes = await fetch(`${API_BASE}/expenses/upload`, { method: 'POST', body: formData })
        const upData = await upRes.json()
        if (upData.success && upData.data?.url) {
          uploadedUrl = upData.data.url
        }
      }

      // 2. Save items with the uploaded bukti and distribution of kredit/debit
      let totalSemuaScan = scannedItems.reduce((acc, it) => acc + ((it.harga_satuan || 0) * (it.jumlah || 1)), 0)
      let overallKredit = parseFloat(scanKredit) || 0
      let overallDebit = overallKredit > 0 ? (overallKredit - totalSemuaScan) : 0
      let groupId = Date.now().toString()

      for (let i = 0; i < scannedItems.length; i++) {
        const item = scannedItems[i]
        const payload = { ...item, group_id: groupId }
        if (uploadedUrl) {
          payload.bukti_pembayaran = uploadedUrl
        }
        // Only set kredit and debit on the first item to avoid double counting on the report totals
        if (i === 0) {
          payload.kredit = overallKredit
          payload.debit = overallDebit
        } else {
          payload.kredit = 0
          payload.debit = 0
        }

        await fetch(`${API_BASE}/expenses/${id}/items`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        })
      }
      setScanModal(false)
      setScannedItems([])
      setScanKredit(0)
      setScanPreview(null)
      setScanFileRaw(null)
      fetchReport()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  // ----- Tashih All -----
  const allTashih = items.length > 0 && items.every(it => it.is_tashih)

  const handleTashihAll = async () => {
    if (!allTashih) {
      Swal.fire('Belum Selesai', 'Harap tashih semua transaksi terlebih dahulu', 'warning')
      return
    }
    const result = await Swal.fire({
      title: 'Tashih Semua Transaksi?',
      text: 'Status laporan akan berubah menjadi Selesai / Confirmed.',
      icon: 'question',
      showCancelButton: true,
      confirmButtonText: 'Ya, Tashih Semua',
      cancelButtonText: 'Batal',
    })
    if (!result.isConfirmed) return
    setSaving(true)
    try {
      const res = await fetch(`${API_BASE}/expenses/${id}/tashih-all`, { method: 'POST' })
      const data = await res.json()
      if (!data.success) throw new Error(data.message)
      Swal.fire('Berhasil!', 'Semua transaksi telah di-tashih', 'success')
      fetchReport()
    } catch (err) {
      Swal.fire('Error', err.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  // ----- Totals -----
  const totalHarga = items.reduce((acc, it) => acc + fmtNum(it.total_harga), 0)
  const totalKredit = items.reduce((acc, it) => acc + fmtNum(it.kredit), 0)
  const totalDebit = items.reduce((acc, it) => acc + fmtNum(it.debit), 0)

  // Hitung jumlah baris setiap kelompok (berdasarkan group_id atau id asli)
  const groupSizes = {}
  items.forEach(it => {
    const g = it.group_id || it.id
    groupSizes[g] = (groupSizes[g] || 0) + 1
  })

  if (loading) return (
    <div className="loading-state">
      <Loader className="spinner" size={32} />
      <p>Memuat laporan...</p>
    </div>
  )

  if (!report) return (
    <div className="empty-state">
      <AlertCircle size={48} style={{ color: 'var(--danger)' }} />
      <p>Laporan tidak ditemukan</p>
      <button className="btn btn-outline btn-sm" onClick={() => navigate('/expenses')}>Kembali</button>
    </div>
  )

  return (
    <div>
      {/* Header */}
      <div className="page-header-row">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button className="btn-icon" onClick={() => navigate('/expenses')}>
            <ArrowLeft size={20} />
          </button>
          <div>
            <h1 className="page-title">
              Laporan Pengeluaran — {MONTHS[report.month - 1]} {report.year}
            </h1>
            <p className="page-subtitle">{report.keterangan || 'Tanpa keterangan'}</p>
          </div>
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <button
            className="btn btn-outline"
            onClick={() => { setScanModal(true); setScanPreview(null); setScannedItems([]) }}
          >
            <Camera size={16} /> Scan Nota (AI)
          </button>
          <button
            className={`btn ${allTashih ? 'btn-success' : 'btn-outline'}`}
            style={allTashih ? { background: 'var(--success)', color: 'white' } : {}}
            disabled={!allTashih || saving || report.status === 'CONFIRMED'}
            onClick={handleTashihAll}
            title={!allTashih ? 'Tashih semua item terlebih dahulu' : ''}
          >
            <CheckSquare size={16} />
            {report.status === 'CONFIRMED' ? 'Sudah Di-Tashih' : 'Tashih Semua Transaksi'}
          </button>
        </div>
      </div>

      {/* Status Bar */}
      <div className="info-grid" style={{ marginBottom: 24 }}>
        <div className="info-item">
          <div className="label">Status</div>
          <div className="value">
            <span className={`badge ${report.status === 'CONFIRMED' ? 'badge-success' : 'badge-warning'}`}>
              {report.status === 'CONFIRMED' ? '✓ Selesai' : 'Draft'}
            </span>
          </div>
        </div>
        <div className="info-item">
          <div className="label">Total Harga Barang</div>
          <div className="value">{formatRp(totalHarga)}</div>
        </div>
        <div className="info-item">
          <div className="label">Total Pengeluaran (Kredit)</div>
          <div className="value" style={{ color: 'var(--danger)' }}>{formatRp(totalKredit)}</div>
        </div>
        <div className="info-item">
          <div className="label">Total Kembalian (Debit)</div>
          <div className="value" style={{ color: 'var(--success)' }}>{formatRp(totalDebit)}</div>
        </div>
      </div>

      {/* Spreadsheet Table */}
      <div className="card">
        <div className="card-header">
          <span className="card-title">Detail Transaksi</span>
          <span className="badge badge-gray">{items.length} item</span>
        </div>
        <div className="table-responsive">
          <table className="table expense-table">
            <thead>
              <tr>
                <th style={{ width: 40 }}>No</th>
                <th>Nama Barang</th>
                <th style={{ width: 100 }}>Jumlah</th>
                <th style={{ width: 130 }}>Total Harga</th>
                <th style={{ width: 130 }}>Kredit</th>
                <th style={{ width: 130 }}>Debit</th>
                <th style={{ width: 90 }}>Bukti Nota</th>
                <th style={{ width: 100 }}>Konfirmasi</th>
                <th style={{ width: 80 }}>Tashih</th>
                <th style={{ width: 50 }}></th>
              </tr>
            </thead>
            <tbody>
              {(() => {
                let currentGroup = null
                let currentNo = 0
                return items.map((item, idx) => {
                  const itemGroupId = item.group_id || item.id
                  const isFirstInGroup = itemGroupId !== currentGroup

                  if (isFirstInGroup) {
                    currentGroup = itemGroupId
                    currentNo++
                  }
                  
                  const totalH = fmtNum(item.harga_satuan) * (parseInt(item.jumlah) || 1)
                  const gSize = groupSizes[itemGroupId] || 1
                  
                  return (
                    <tr key={item.id} className={item.is_tashih ? 'row-tashih' : ''}>
                      {isFirstInGroup && (
                        <td rowSpan={gSize} style={{ textAlign: 'center', color: 'var(--text-muted)' }}>
                          {currentNo}
                        </td>
                      )}

                    {/* Nama Barang */}
                    <td>
                      {editing?.idx === idx && editing?.field === 'nama_barang' ? (
                        <input
                          className="cell-input"
                          value={item.nama_barang}
                          onChange={e => handleCellChange(idx, 'nama_barang', e.target.value)}
                          onBlur={() => saveEdit(idx)}
                          onKeyDown={e => e.key === 'Enter' && saveEdit(idx)}
                          autoFocus
                        />
                      ) : (
                        <div className="cell-display" onClick={() => startEdit(idx, 'nama_barang')}>
                          <span className="barang-name">{item.nama_barang || '—'}</span>
                          <span className="barang-price">@{formatRp(item.harga_satuan)}</span>
                        </div>
                      )}
                    </td>

                    {/* Jumlah */}
                    <td>
                      {editing?.idx === idx && editing?.field === 'jumlah' ? (
                        <input
                          className="cell-input cell-number"
                          type="number"
                          min={1}
                          value={item.jumlah}
                          onChange={e => handleCellChange(idx, 'jumlah', e.target.value)}
                          onBlur={() => saveEdit(idx)}
                          onKeyDown={e => e.key === 'Enter' && saveEdit(idx)}
                          autoFocus
                        />
                      ) : (
                        <div className="cell-display cell-number" onClick={() => startEdit(idx, 'jumlah')}>
                          {item.jumlah}
                        </div>
                      )}
                    </td>

                    {/* Total Harga (computed) */}
                    <td>
                      <div className="cell-display cell-amount">{formatRp(totalH)}</div>
                    </td>

                    {/* Kredit */}
                    {isFirstInGroup && (
                      <td rowSpan={gSize} style={{ verticalAlign: 'middle', textAlign: 'right', paddingRight: 8 }}>
                        {editing?.idx === idx && editing?.field === 'kredit' ? (
                          <input
                            autoFocus
                            type="text"
                            className="cell-input cell-number"
                            value={formatInputAngka(item.kredit)}
                            onChange={e => handleCellChange(idx, 'kredit', parseInputAngka(e.target.value))}
                            onBlur={() => saveEdit(idx)}
                            onKeyDown={e => { if (e.key === 'Enter') saveEdit(idx) }}
                          />
                        ) : (
                          <div className="cell-text cell-number" onClick={() => (report.status !== 'CONFIRMED' && item.is_tashih === false) && setEditing({ idx, field: 'kredit' })}>
                            {formatRp(item.kredit)}
                          </div>
                        )}
                      </td>
                    )}

                    {/* Debit */}
                    {isFirstInGroup && (
                      <td rowSpan={gSize} style={{ verticalAlign: 'middle', textAlign: 'right', paddingRight: 8, background: 'var(--bg-secondary)', fontWeight: 600, fontSize: 13, color: 'var(--success)' }}>
                        {formatRp(item.debit)}
                      </td>
                    )}

                    {/* Bukti Nota */}
                    {isFirstInGroup && (
                      <td rowSpan={gSize} style={{ textAlign: 'center', verticalAlign: 'middle' }}>
                        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
                          {item.bukti_pembayaran ? (
                            <div style={{ position: 'relative' }}>
                              <img
                                src={item.bukti_pembayaran}
                                alt="Nota"
                                style={{ width: 36, height: 36, objectFit: 'cover', borderRadius: 4, cursor: 'pointer', border: '1px solid var(--border-color)' }}
                                onClick={() => setPreviewImg(item.bukti_pembayaran)}
                              />
                              {/* Tombol kecil untuk ganti/upload ulang jika belum confirmed */}
                              {report.status !== 'CONFIRMED' && (
                                <label
                                  title="Ganti Bukti"
                                  style={{
                                    position: 'absolute', bottom: -6, right: -6, background: 'var(--primary-color)',
                                    color: 'white', borderRadius: '50%', width: 20, height: 20, display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer',
                                    boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                                  }}
                                >
                                  {uploadingIdx === item.id ? <Loader size={10} className="spinner" /> : <Upload size={10} />}
                                  <input
                                    type="file"
                                    accept="image/*"
                                    style={{ display: 'none' }}
                                    onChange={async (e) => {
                                      const f = e.target.files[0]
                                      if (!f) return
                                      setUploadingIdx(item.id)
                                      await handleUploadBukti(item.id, f)
                                      setUploadingIdx(null)
                                      e.target.value = ''
                                    }}
                                  />
                                </label>
                              )}
                            </div>
                          ) : (
                            <label className="btn-icon" title="Upload Bukti" style={{ cursor: 'pointer', color: 'var(--text-muted)' }}>
                              {uploadingIdx === item.id ? <Loader size={14} className="spinner" /> : <Upload size={14} />}
                              <input
                                type="file"
                                accept="image/*"
                                style={{ display: 'none' }}
                                onChange={async (e) => {
                                  const f = e.target.files[0]
                                  if (!f) return
                                  setUploadingIdx(item.id)
                                  await handleUploadBukti(item.id, f)
                                  setUploadingIdx(null)
                                  e.target.value = ''
                                }}
                              />
                            </label>
                          )}
                        </div>
                      </td>
                    )}

                    {/* Konfirmasi */}
                    {isFirstInGroup && (
                      <td rowSpan={gSize} style={{ textAlign: 'center', verticalAlign: 'middle' }}>
                        <button
                          className={`toggle-check ${item.is_konfirmasi ? 'toggled' : ''}`}
                          onClick={() => handleToggleGroup(itemGroupId, 'is_konfirmasi', item.is_konfirmasi)}
                          title={item.is_konfirmasi ? 'Batalkan Konfirmasi' : 'Konfirmasi'}
                          disabled={report.status === 'CONFIRMED'}
                        >
                          {item.is_konfirmasi ? <Check size={16} /> : <Square size={16} />}
                        </button>
                      </td>
                    )}

                    {/* Tashih */}
                    {isFirstInGroup && (
                      <td rowSpan={gSize} style={{ textAlign: 'center', verticalAlign: 'middle' }}>
                        <button
                          className={`toggle-check ${item.is_tashih ? 'toggled tashih' : ''}`}
                          onClick={() => handleToggleGroup(itemGroupId, 'is_tashih', item.is_tashih)}
                          title={item.is_tashih ? 'Batalkan Tashih' : 'Tashih'}
                          disabled={report.status === 'CONFIRMED'}
                        >
                          {item.is_tashih ? <Check size={16} /> : <Square size={16} />}
                        </button>
                      </td>
                    )}

                    {/* Delete */}
                    <td style={{ textAlign: 'center' }}>
                      {report.status !== 'CONFIRMED' && (
                        <button className="btn-icon delete" onClick={() => handleDelete(item.id)}>
                          <Trash2 size={14} />
                        </button>
                      )}
                    </td>
                  </tr>
                )
              })})()}

            </tbody>

            {/* Manual input segment */}
            {report.status !== 'CONFIRMED' && (
              <tbody>
                {manualItems.map((mItem, mIdx) => (
                  <tr key={`manual-${mIdx}`} className="add-item-row" style={{ background: 'var(--bg-card)' }}>
                    <td style={{ textAlign: 'center', verticalAlign: 'top', paddingTop: 10 }}>
                      <button className="btn-icon" style={{ background: 'var(--border-color)', margin: '0 auto' }} onClick={() => setManualItems(p => [...p, { ...EMPTY_ITEM }])}>
                        <Plus size={14} />
                      </button>
                    </td>
                    <td style={{ verticalAlign: 'top', paddingTop: 6 }}>
                      <input
                        className="cell-input"
                        placeholder="Nama barang..."
                        value={mItem.nama_barang}
                        onChange={e => handleManualItemChange(mIdx, 'nama_barang', e.target.value)}
                      />
                      <input
                        type="text"
                        className="cell-input cell-number"
                        placeholder="Harga @Rp"
                        style={{ marginTop: 4 }}
                        value={formatInputAngka(mItem.harga_satuan)}
                        onChange={e => handleManualItemChange(mIdx, 'harga_satuan', parseInputAngka(e.target.value))}
                      />
                    </td>
                    <td style={{ verticalAlign: 'top', paddingTop: 6 }}>
                      <input
                        type="text"
                        className="cell-input cell-number"
                        placeholder="Qty"
                        value={formatInputAngka(mItem.jumlah)}
                        onChange={e => handleManualItemChange(mIdx, 'jumlah', parseInputAngka(e.target.value))}
                        style={{ textAlign: 'center' }}
                      />
                    </td>
                    <td style={{ textAlign: 'right', verticalAlign: 'top', paddingTop: 12, fontWeight: 600 }}>
                      {formatRp((parseFloat(mItem.harga_satuan) || 0) * (parseInt(mItem.jumlah) || 1))}
                    </td>

                    {mIdx === 0 && (
                      <>
                        <td rowSpan={manualItems.length} style={{ verticalAlign: 'top', paddingTop: 6 }}>
                          <input
                            className="cell-input cell-number"
                            type="text"
                            placeholder="Kredit"
                            value={formatInputAngka(manualKredit)}
                            onChange={e => setManualKredit(parseInputAngka(e.target.value))}
                          />
                        </td>
                        <td rowSpan={manualItems.length} style={{ verticalAlign: 'top', paddingTop: 6 }}>
                          <div style={{ padding: '6px', background: 'var(--bg-secondary)', borderRadius: 4, fontWeight: 600, fontSize: 13, color: 'var(--success)' }}>
                            {formatRp(
                              (() => {
                                const tot = manualItems.reduce((acc, it) => acc + ((parseFloat(it.harga_satuan)||0) * (parseInt(it.jumlah)||1)), 0)
                                const kr = parseFloat(manualKredit)||0
                                return kr > 0 ? (kr - tot) : 0
                              })()
                            )}
                          </div>
                        </td>
                        <td rowSpan={manualItems.length} style={{ verticalAlign: 'top', paddingTop: 6 }}>
                          <div className="bukti-cell-view">
                            {manualBuktiPreview ? (
                              <img src={manualBuktiPreview} alt="Nota" onClick={() => setPreviewImg(manualBuktiPreview)} style={{ width: 40, height: 40, objectFit: 'cover', borderRadius: 4, cursor: 'pointer' }} />
                            ) : (
                              <div className="bukti-placeholder" style={{ width: 40, height: 40, background: 'var(--bg-secondary)', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: 4 }}>
                                <Camera size={14} color="var(--text-muted)" />
                              </div>
                            )}
                            <label className="bukti-upload-btn-small" style={{ marginTop: 6, display: 'inline-block' }}>
                              <Upload size={12} />
                              <input
                                type="file"
                                accept="image/*"
                                style={{ display: 'none' }}
                                onChange={e => {
                                  if (e.target.files[0]) {
                                    setManualBuktiFile(e.target.files[0])
                                    setManualBuktiPreview(URL.createObjectURL(e.target.files[0]))
                                  }
                                }}
                              />
                            </label>
                          </div>
                        </td>
                        <td colSpan={3} rowSpan={manualItems.length} style={{ verticalAlign: 'middle', textAlign: 'center' }}>
                          <button className="btn btn-primary" onClick={handleSaveManualGroup} disabled={addingItem}>
                            {addingItem ? <Loader className="spinner" size={16} /> : <Save size={16} />}
                            Simpan Transaksi
                          </button>
                        </td>
                      </>
                    )}
                  </tr>
                ))}
              </tbody>
            )}

            {/* Footer totals */}
            <tfoot>
              <tr className="totals-row">
                <td colSpan={2} style={{ textAlign: 'right', fontWeight: 700 }}>TOTAL</td>
                <td></td>
                <td className="total-cell">{formatRp(totalHarga)}</td>
                <td className="total-cell kredit">{formatRp(totalKredit)}</td>
                <td className="total-cell debit">{formatRp(totalDebit)}</td>
                <td colSpan={4} style={{ textAlign: 'center', fontSize: 12, color: 'var(--text-muted)' }}>
                  {allTashih && items.length > 0
                    ? <span style={{ color: 'var(--success)', fontWeight: 600 }}>✓ Semua transaksi siap di-tashih</span>
                    : items.length > 0
                    ? `${items.filter(it => it.is_tashih).length}/${items.length} ditashih`
                    : ''}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>

      {/* AI Scan Modal */}
      {scanModal && (
        <div className="modal-overlay" onClick={() => setScanModal(false)}>
          <div className="modal-box modal-wide" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title"><Camera size={18} /> Scan Nota dengan AI</h3>
              <button className="btn-icon" onClick={() => setScanModal(false)}><X size={18} /></button>
            </div>
            <div className="modal-body">
              {!scannedItems.length && !scanning && (
                <div className="scan-upload-area">
                  <Camera size={40} style={{ color: 'var(--primary-color)', marginBottom: 12 }} />
                  <p style={{ fontWeight: 600, marginBottom: 4 }}>Upload Foto Nota / Kwitansi</p>
                  <p style={{ color: 'var(--text-muted)', fontSize: 13, marginBottom: 16 }}>
                    AI akan membaca nota dan mengisi kolom secara otomatis
                  </p>
                  <label className="btn btn-primary" style={{ cursor: 'pointer' }}>
                    <Upload size={16} /> Pilih Gambar
                    <input
                      type="file"
                      accept="image/*"
                      style={{ display: 'none' }}
                      ref={scanFileRef}
                      onChange={handleScanFile}
                    />
                  </label>
                </div>
              )}
              {scanning && (
                <div className="scan-upload-area">
                  <Loader size={36} className="spinner" style={{ color: 'var(--primary-color)', marginBottom: 12 }} />
                  <p style={{ fontWeight: 600 }}>AI sedang membaca nota...</p>
                  <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 4 }}>Proses biasanya 5-15 detik (maks. 30 detik)</p>
                  {scanPreview && <img src={scanPreview} alt="preview" className="scan-preview-img" />}
                </div>
              )}
              {!scanning && scannedItems.length > 0 && (
                <div>
                  <div style={{ display: 'flex', gap: 16, marginBottom: 16 }}>
                    {scanPreview && (
                      <img src={scanPreview} alt="nota" className="scan-preview-img-small" />
                    )}
                    <div style={{ flex: 1 }}>
                      <p style={{ fontWeight: 600, marginBottom: 8 }}>
                        {scanTimedOut
                          ? '⚠️ AI tidak berhasil membaca — isi manual di bawah'
                          : '✓ Hasil Scan — Periksa & Edit Jika Perlu'}
                      </p>
                      <table className="table scan-result-table">
                        <thead>
                          <tr>
                            <th>Nama Barang</th>
                            <th>Harga Satuan</th>
                            <th>Jumlah</th>
                            <th>Total (Rp)</th>
                            <th></th>
                          </tr>
                        </thead>
                        <tbody>
                          {scannedItems.map((item, idx) => (
                            <tr key={idx}>
                              <td>
                                <input className="cell-input" value={item.nama_barang}
                                  onChange={e => setScannedItems(prev => {
                                    const n = [...prev]; n[idx] = { ...n[idx], nama_barang: e.target.value }; return n
                                  })} />
                              </td>
                              <td>
                                <input className="cell-input cell-number" type="text" value={formatInputAngka(item.harga_satuan)}
                                  onChange={e => setScannedItems(prev => {
                                    const n = [...prev]; n[idx] = { ...n[idx], harga_satuan: parseInputAngka(e.target.value) }; return n
                                  })} />
                              </td>
                              <td>
                                <input className="cell-input cell-number" type="text" value={formatInputAngka(item.jumlah)}
                                  onChange={e => setScannedItems(prev => {
                                    const n = [...prev]; n[idx] = { ...n[idx], jumlah: parseInputAngka(e.target.value) }; return n
                                  })} />
                              </td>
                              <td style={{ textAlign: 'right', paddingRight: 8, fontWeight: 600, color: 'var(--text-color)' }}>
                                {formatRp((item.harga_satuan || 0) * (item.jumlah || 1))}
                              </td>
                              <td>
                                <button className="btn-icon delete"
                                  onClick={() => setScannedItems(prev => prev.filter((_, i) => i !== idx))}>
                                  <X size={14} />
                                </button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                        <tfoot>
                          {(() => {
                            const totScan = scannedItems.reduce((acc, it) => acc + ((it.harga_satuan || 0) * (it.jumlah || 1)), 0)
                            const valKredit = parseFloat(scanKredit) || 0
                            const valDebit = valKredit > 0 ? (valKredit - totScan) : 0
                            return (
                              <tr style={{ background: 'var(--bg-secondary)' }}>
                                <td colSpan={3} style={{ textAlign: 'right', fontWeight: 600 }}>Total Semua Barang:</td>
                                <td style={{ textAlign: 'right', paddingRight: 8, fontWeight: 700 }}>{formatRp(totScan)}</td>
                                <td></td>
                              </tr>
                            )
                          })()}
                        </tfoot>
                      </table>
                      <div style={{ display: 'flex', gap: 12, marginTop: 16, alignItems: 'center', background: 'var(--bg-secondary)', padding: 12, borderRadius: 8 }}>
                        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 4 }}>
                          <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-muted)' }}>KREDIT (Uang Dibayar)</span>
                          <input className="cell-input cell-number" type="text" 
                            style={{ background: 'white' }}
                            value={formatInputAngka(scanKredit)} 
                            placeholder="Isi uang yang dibayar keseluruhan"
                            onChange={e => setScanKredit(parseInputAngka(e.target.value))} />
                        </div>
                        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 4 }}>
                          <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-muted)' }}>DEBIT (Kembalian - Auto)</span>
                          <div style={{ padding: '8px 12px', background: 'white', borderRadius: 4, fontWeight: 600, color: 'var(--success)', border: '1px solid var(--border-color)' }}>
                            {formatRp((parseFloat(scanKredit)||0) > 0 ? ((parseFloat(scanKredit)||0) - scannedItems.reduce((acc, it) => acc + ((it.harga_satuan || 0) * (it.jumlah || 1)), 0)) : 0)}
                          </div>
                        </div>
                      </div>
                      <button className="btn btn-outline btn-sm" style={{ marginTop: 12 }}
                        onClick={() => setScannedItems(prev => [...prev, { ...EMPTY_ITEM }])}>
                        <Plus size={14} /> Tambah Baris
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
            {!scanning && scannedItems.length > 0 && (
              <div className="modal-footer">
                <button className="btn btn-outline"
                  onClick={() => { setScannedItems([]); setScanPreview(null); setScanFileRaw(null) }}>
                  ← Scan Ulang
                </button>
                <button className="btn btn-primary" onClick={handleConfirmScan} disabled={saving}>
                  {saving ? <Loader size={14} className="spinner" /> : <Save size={14} />}
                  Simpan Semua Item
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Image Preview Modal */}
      {previewImg && (
        <div className="modal-overlay" onClick={() => setPreviewImg(null)}>
          <div style={{ position: 'relative' }}>
            <button
              className="btn-icon"
              style={{ position: 'absolute', top: -40, right: 0, background: 'white', borderRadius: 8 }}
              onClick={() => setPreviewImg(null)}
            >
              <X size={20} />
            </button>
            <img
              src={previewImg}
              alt="Bukti Pembayaran"
              style={{ maxWidth: '90vw', maxHeight: '80vh', borderRadius: 12 }}
            />
          </div>
        </div>
      )}
    </div>
  )
}
