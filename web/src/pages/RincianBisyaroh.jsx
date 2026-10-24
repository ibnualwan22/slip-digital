import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api, { getName, clearSession } from '../api'
import TeacherNav from '../components/TeacherNav'

// Halaman ringkasan bisyaroh untuk pengajar: menampilkan HANYA komponen
// yang benar-benar ada pada payroll-nya (tidak ada baris template —
// mis. tanpa Gaji Pokok ya tidak tampil). Baris Jam Mengajar menampilkan
// total sesi + nominal per sesi, dengan tombol Detail ke Rekap Saya.
const TABAYUN_WA = '6285853248075'
const BULAN = ['', 'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
const JAM_CODES = ['JAM_MENGAJAR', 'JAM_AJAR', 'MENGAJAR']
const DENDA_CODES = ['TERLAMBAT', 'DENDA']

function formatTanggal(iso) {
  if (!iso) return '-'
  const [y, m, d] = String(iso).slice(0, 10).split('-').map(Number)
  if (!y || !m || !d) return iso
  return `${d} ${BULAN[m]} ${y}`
}

function formatRupiah(v) {
  const n = Math.round(Number(v || 0))
  return 'Rp' + n.toLocaleString('id-ID')
}

function statusLabel(status) {
  if (status === 'DRAFT') return 'Masih draft — angka dapat berubah mengikuti absensi terbaru'
  if (status === 'CONFIRMED') return 'Terkonfirmasi'
  if (status === 'PAID') return 'Sudah dibayar'
  return status || ''
}

export default function RincianBisyaroh() {
  const navigate = useNavigate()
  const [periods, setPeriods] = useState([])
  const [selected, setSelected] = useState(null)
  const [recap, setRecap] = useState(null)
  const [loading, setLoading] = useState(true)
  const [loadingRecap, setLoadingRecap] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.get('/me/periods')
      .then((res) => setPeriods(res.data || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  // Datang dari Rekap Saya: langsung buka periode tersebut.
  useEffect(() => {
    if (periods.length === 0 || selected) return
    const pid = new URLSearchParams(window.location.search).get('period_id')
    if (!pid) return
    const p = periods.find((x) => String(x.id) === String(pid))
    if (p) openPeriod(p)
  }, [periods])

  const openPeriod = async (p) => {
    setSelected(p)
    setRecap(null)
    setError('')
    setLoadingRecap(true)
    try {
      const res = await api.get(`/me/recap?period_id=${p.id}`)
      setRecap(res.data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoadingRecap(false)
    }
  }

  const logout = () => {
    clearSession()
    window.location.href = '/login'
  }

  const card = { background: '#fff', border: '1px solid #E5E7EB', borderRadius: 14, padding: '18px 20px' }
  const statLabel = { fontSize: 11, fontWeight: 700, color: '#6B7280', textTransform: 'uppercase', letterSpacing: '.04em' }
  const statValue = { fontSize: 24, fontWeight: 800, color: '#1E3A5F', marginTop: 4 }

  // Kelompokkan rincian: keluarga jam jadi satu baris, keluarga denda satu
  // baris, sisanya apa adanya — dan hanya yang benar-benar ada yang tampil.
  const buildRows = (details) => {
    const rows = []
    const jamRows = details.filter(d => JAM_CODES.includes(d.code))
    const dendaRows = details.filter(d => DENDA_CODES.includes(d.code))
    const rest = details.filter(d => !JAM_CODES.includes(d.code) && !DENDA_CODES.includes(d.code))
    const sum = (list, key) => list.reduce((s, d) => s + Number(d[key] || 0), 0)
    const uniformRate = (list) => {
      const rates = [...new Set(list.map(d => String(d.rate)))]
      return rates.length === 1 ? list[0].rate : null
    }
    let jamDone = false
    let dendaDone = false
    for (const d of details) {
      if (JAM_CODES.includes(d.code)) {
        if (!jamDone) {
          jamDone = true
          rows.push({ kind: 'jam', qty: sum(jamRows, 'quantity'), total: sum(jamRows, 'total_amount'), rate: uniformRate(jamRows) })
        }
      } else if (DENDA_CODES.includes(d.code)) {
        if (!dendaDone) {
          dendaDone = true
          rows.push({ kind: 'denda', qty: sum(dendaRows, 'quantity'), total: sum(dendaRows, 'total_amount'), rate: uniformRate(dendaRows) })
        }
      } else {
        rows.push({ kind: 'single', d })
      }
    }
    return rows
  }

  const verifikasiLink = () => {
    const msg =
      `Assalamu'alaikum, saya ${getName()} (pengajar). ` +
      `Saya ingin memverifikasi rincian bisyaroh saya:%0A%0A` +
      `Periode: ${selected.dufah_name} (Gaji ${BULAN[selected.month]} ${selected.year})%0A%0A` +
      `Mohon ditinjau kembali. Terima kasih.`
    return `https://wa.me/${TABAYUN_WA}?text=${msg}`
  }

  const detail = recap?.payroll
  const detailRows = detail ? buildRows(detail.details || []) : []

  return (
    <div style={{ minHeight: '100vh', background: '#F3F4F6', fontFamily: 'Inter, system-ui, sans-serif' }}>
      <header style={{ background: '#1E3A5F', color: '#fff', padding: '14px 20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <div style={{ fontWeight: 800, fontSize: 17 }}>E-Maliyah — Rincian Bisyaroh</div>
          <div style={{ fontSize: 12, opacity: .8 }}>{getName()}</div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <a href="https://bawwabah.markazarabiyah.site" style={{ background: '#fff', color: '#1E3A5F', border: 'none', borderRadius: 8, padding: '8px 14px', cursor: 'pointer', fontWeight: 700, textDecoration: 'none', textAlign: 'center' }}>Bawwabah</a>
          <button onClick={logout} style={{ background: 'rgba(255,255,255,.15)', color: '#fff', border: 'none', borderRadius: 8, padding: '8px 14px', cursor: 'pointer', fontWeight: 600 }}>Keluar</button>
        </div>
      </header>

      <main style={{ maxWidth: 900, margin: '0 auto', padding: 20 }}>
        <TeacherNav active="rincian" periodId={selected?.id} />
        {error && (
          <div style={{ background: '#FEF2F2', color: '#B91C1C', border: '1px solid #FECACA', borderRadius: 10, padding: '10px 14px', fontSize: 14, marginBottom: 16 }}>{error}</div>
        )}

        {!selected && (
          <>
            <h2 style={{ fontSize: 16, color: '#111827', margin: '4px 0 14px' }}>Pilih Periode (Dufah)</h2>
            {loading && <p style={{ color: '#6B7280' }}>Memuat…</p>}
            {!loading && periods.length === 0 && (
              <div style={card}>Belum ada periode yang dipetakan oleh admin. Silakan hubungi admin keuangan.</div>
            )}
            <div style={{ display: 'grid', gap: 12 }}>
              {periods.map((p) => (
                <button key={p.id} onClick={() => openPeriod(p)} style={{ ...card, textAlign: 'left', cursor: 'pointer' }}>
                  <div style={{ fontWeight: 700, fontSize: 16, color: '#1E3A5F' }}>{p.dufah_name}</div>
                  <div style={{ fontSize: 13, color: '#6B7280', marginTop: 2 }}>
                    Gaji {BULAN[p.month]} {p.year} • Periode sesi {formatTanggal(p.start_date)} s.d. {formatTanggal(p.end_date)}
                  </div>
                </button>
              ))}
            </div>
          </>
        )}

        {selected && (
          <>
            <button onClick={() => { setSelected(null); setRecap(null) }} style={{ background: 'none', border: 'none', color: '#2563EB', fontWeight: 600, cursor: 'pointer', padding: 0, marginBottom: 12 }}>← Kembali ke daftar periode</button>
            <h2 style={{ fontSize: 17, color: '#111827', margin: '0 0 4px' }}>{selected.dufah_name}</h2>
            <p style={{ fontSize: 13, color: '#6B7280', margin: '0 0 16px' }}>Gaji {BULAN[selected.month]} {selected.year} • Sesi {formatTanggal(selected.start_date)} s.d. {formatTanggal(selected.end_date)}</p>

            {loadingRecap && <p style={{ color: '#6B7280' }}>Memuat data…</p>}

            {recap && !detail && (
              <div style={card}>Payroll bulan ini belum dibuat. Silakan hubungi admin keuangan.</div>
            )}

            {recap && detail && (
              <>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(160px,1fr))', gap: 12, marginBottom: 16 }}>
                  <div style={card}>
                    <div style={statLabel}>Gaji Diterima (THP)</div>
                    <div style={{ ...statValue, fontSize: 20, lineHeight: 1.25, overflowWrap: 'anywhere' }}>{formatRupiah(detail.take_home_pay)}</div>
                    <div style={{ fontSize: 12, color: '#6B7280' }}>Payroll {BULAN[detail.month]} {detail.year} • {statusLabel(detail.status)}</div>
                  </div>
                </div>

                <div style={{ ...card, padding: 0, overflow: 'hidden', marginBottom: 16 }}>
                  <div style={{ padding: '14px 20px', fontWeight: 700, color: '#111827', borderBottom: '1px solid #E5E7EB' }}>Rincian Bisyaroh</div>
                  {detailRows.length === 0 && (
                    <div style={{ padding: '16px 20px', color: '#6B7280', fontSize: 14 }}>Belum ada komponen gaji pada payroll ini.</div>
                  )}
                  {detailRows.map((row, i) => {
                    if (row.kind === 'jam') {
                      return (
                        <div key={'jam'} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, padding: '12px 20px', borderBottom: '1px solid #F3F4F6' }}>
                          <div>
                            <div style={{ fontWeight: 600, fontSize: 14, color: '#111827' }}>Jam Mengajar</div>
                            <div style={{ fontSize: 12, color: '#6B7280' }}>{row.qty} sesi{row.rate !== null ? ` × ${formatRupiah(row.rate)} per sesi` : ' (tarif bervariasi)'}</div>
                          </div>
                          <div style={{ textAlign: 'right', flexShrink: 0 }}>
                            <div style={{ fontSize: 14, fontWeight: 700, color: '#111827' }}>{formatRupiah(row.total)}</div>
                            <button onClick={() => navigate(`/rekap-saya?period_id=${selected.id}`)} style={{ background: 'none', border: 'none', color: '#2563EB', fontWeight: 600, cursor: 'pointer', padding: 0, fontSize: 13 }}>Detail →</button>
                          </div>
                        </div>
                      )
                    }
                    if (row.kind === 'denda') {
                      return (
                        <div key={'denda'} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, padding: '12px 20px', borderBottom: '1px solid #F3F4F6' }}>
                          <div>
                            <div style={{ fontWeight: 600, fontSize: 14, color: '#111827' }}>Denda Terlambat</div>
                            <div style={{ fontSize: 12, color: '#6B7280' }}>{row.qty} menit{row.rate !== null ? ` × ${formatRupiah(row.rate)}` : ''}</div>
                          </div>
                          <div style={{ textAlign: 'right', flexShrink: 0 }}>
                            <div style={{ fontSize: 14, fontWeight: 700, color: '#B91C1C' }}>-{formatRupiah(row.total)}</div>
                            <button onClick={() => navigate(`/rekap-saya?period_id=${selected.id}`)} style={{ background: 'none', border: 'none', color: '#2563EB', fontWeight: 600, cursor: 'pointer', padding: 0, fontSize: 13 }}>Detail →</button>
                          </div>
                        </div>
                      )
                    }
                    const d = row.d
                    return (
                      <div key={d.activity_name + i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, padding: '12px 20px', borderBottom: '1px solid #F3F4F6' }}>
                        <div>
                          <div style={{ fontWeight: 600, fontSize: 14, color: '#111827' }}>{d.activity_name}</div>
                          <div style={{ fontSize: 12, color: '#6B7280' }}>{Number(d.quantity)} × {formatRupiah(d.rate)}</div>
                        </div>
                        <div style={{ fontSize: 14, fontWeight: 700, color: d.type === 'DEDUCTION' ? '#B91C1C' : '#111827' }}>{d.type === 'DEDUCTION' ? '-' : ''}{formatRupiah(d.total_amount)}</div>
                      </div>
                    )
                  })}
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '12px 20px', borderTop: '1px solid #E5E7EB', fontWeight: 700, fontSize: 14 }}>
                    <span>Total Diterima (THP)</span>
                    <span>{formatRupiah(detail.take_home_pay)}</span>
                  </div>
                </div>

                <div style={{ ...card, background: '#FFFBEB', borderColor: '#FDE68A' }}>
                  <div style={{ fontSize: 13, color: '#92400E', lineHeight: 1.6 }}>
                    Data absensi dan jam mengajar di halaman ini diambil otomatis dari Siakad.
                    Jika ada ketidaksesuaian, silakan verifikasi ke Ustadzah Karisma.
                  </div>
                  <a href={verifikasiLink()} target="_blank" rel="noreferrer" style={{ display: 'inline-block', marginTop: 10, background: '#059669', color: '#fff', borderRadius: 8, padding: '9px 16px', fontWeight: 700, fontSize: 13, textDecoration: 'none' }}>Verifikasi via WhatsApp</a>
                </div>
              </>
            )}
          </>
        )}
      </main>
    </div>
  )
}
