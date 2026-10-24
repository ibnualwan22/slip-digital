import React, { useEffect, useState } from 'react'
import api, { getName, clearSession } from '../api'
import TeacherNav from '../components/TeacherNav'

const TABAYUN_WA = '6285853248075'
const BULAN = ['', 'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
const HARI = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']

function formatTanggal(iso) {
  if (!iso) return '-'
  const [y, m, d] = String(iso).slice(0, 10).split('-').map(Number)
  if (!y || !m || !d) return iso
  return `${d} ${BULAN[m]} ${y}`
}

function formatTanggalLengkap(iso) {
  if (!iso) return '-'
  const [y, m, d] = String(iso).slice(0, 10).split('-').map(Number)
  if (!y || !m || !d) return iso
  const hari = HARI[new Date(y, m - 1, d).getDay()]
  return `${hari}, ${d} ${BULAN[m]} ${y}`
}

// "SESI_4" -> "Sesi 4" (ambil angkanya saja agar tidak dobel)
function labelSesi(sesi) {
  const angka = String(sesi || '').replace(/\D/g, '')
  return angka ? `Sesi ${angka}` : String(sesi || '-')
}

function nomorSesi(sesi) {
  const angka = parseInt(String(sesi || '').replace(/\D/g, ''), 10)
  return Number.isNaN(angka) ? 0 : angka
}

function formatRupiah(v) {
  const n = Math.round(Number(v || 0))
  return 'Rp' + n.toLocaleString('id-ID')
}

export default function RekapSaya() {
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

  // Datang dari halaman Rincian Bisyaroh: langsung buka periode tersebut.
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

  const sessionsSorted = recap
    ? [...recap.sessions].sort((a, b) => {
        const byDate = String(b.tanggal).localeCompare(String(a.tanggal))
        if (byDate !== 0) return byDate
        const bySesi = nomorSesi(b.sesi) - nomorSesi(a.sesi)
        if (bySesi !== 0) return bySesi
        return String(b.waktuMulai || '').localeCompare(String(a.waktuMulai || ''))
      })
    : []

  const tabayunLink = (s) => {
    const msg =
      `Assalamu'alaikum, saya ${getName()} (pengajar). ` +
      `Saya mengajukan tabayun untuk catatan keterlambatan saya:%0A%0A` +
      `Periode: ${selected.dufah_name} (${BULAN[selected.month]} ${selected.year})%0A` +
      `Tanggal: ${formatTanggalLengkap(s.tanggal)}%0A` +
      `${labelSesi(s.sesi)}${s.kelas ? ` (${s.kelas})` : ''}%0A` +
      `Tercatat terlambat: ${s.terlambatMenit} menit%0A%0A` +
      `Mohon ditinjau kembali. Terima kasih.`
    return `https://wa.me/${TABAYUN_WA}?text=${msg}`
  }

  const logout = () => {
    clearSession()
    window.location.href = '/login'
  }

  const card = { background: '#fff', border: '1px solid #E5E7EB', borderRadius: 14, padding: '18px 20px' }
  const statLabel = { fontSize: 11, fontWeight: 700, color: '#6B7280', textTransform: 'uppercase', letterSpacing: '.04em' }
  const statValue = { fontSize: 24, fontWeight: 800, color: '#1E3A5F', marginTop: 4 }

  return (
    <div style={{ minHeight: '100vh', background: '#F3F4F6', fontFamily: 'Inter, system-ui, sans-serif' }}>
      <header style={{ background: '#1E3A5F', color: '#fff', padding: '14px 20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <div style={{ fontWeight: 800, fontSize: 17 }}>E-Maliyah — Rekap Saya</div>
          <div style={{ fontSize: 12, opacity: .8 }}>{getName()}</div>
        </div>
        <button onClick={logout} style={{ background: 'rgba(255,255,255,.15)', color: '#fff', border: 'none', borderRadius: 8, padding: '8px 14px', cursor: 'pointer', fontWeight: 600 }}>Keluar</button>
      </header>

      <main style={{ maxWidth: 900, margin: '0 auto', padding: 20 }}>
        <TeacherNav active="rekap" periodId={selected?.id} />
        {error && (
          <div style={{ background: '#FEF2F2', color: '#B91C1C', border: '1px solid #FECACA', borderRadius: 10, padding: '10px 14px', fontSize: 14, marginBottom: 16 }}>{error}</div>
        )}

        {!selected && (
          <>
            <h2 style={{ fontSize: 16, color: '#111827', margin: '4px 0 14px' }}>Riwayat Periode (Dufah)</h2>
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

            {loadingRecap && <p style={{ color: '#6B7280' }}>Memuat data dari Siakad…</p>}

            {recap && (
              <>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(160px,1fr))', gap: 12, marginBottom: 16 }}>
                  <div style={card}><div style={statLabel}>Sesi Terverifikasi</div><div style={statValue}>{recap.summary.total_sesi}</div></div>
                  <div style={card}><div style={statLabel}>Total Terlambat</div><div style={statValue}>{recap.summary.total_terlambat_menit} <span style={{ fontSize: 13 }}>menit</span></div></div>
                  <div style={card}><div style={statLabel}>Denda Terlambat</div><div style={statValue}>{formatRupiah(recap.summary.total_denda)}</div></div>
                  <div style={card}>
                    <div style={statLabel}>Gaji Diterima (THP)</div>
                    <div style={{ ...statValue, fontSize: 20, lineHeight: 1.25, overflowWrap: 'anywhere' }}>{recap.payroll ? formatRupiah(recap.payroll.take_home_pay) : '—'}</div>
                    <div style={{ fontSize: 12, color: '#6B7280' }}>{recap.payroll ? `Payroll ${BULAN[recap.payroll.month]} ${recap.payroll.year} • ${recap.payroll.status}` : 'Payroll bulan ini belum dibuat'}</div>
                  </div>
                </div>

                <div style={{ ...card, padding: 0, overflow: 'hidden' }}>
                  <div style={{ padding: '14px 20px', fontWeight: 700, color: '#111827', borderBottom: '1px solid #E5E7EB' }}>Rincian Per Sesi</div>
                  {sessionsSorted.length === 0 && (
                    <div style={{ padding: '16px 20px', color: '#6B7280', fontSize: 14 }}>Tidak ada sesi terverifikasi pada periode ini.</div>
                  )}
                  {sessionsSorted.map((s, i) => (
                    <div key={s.id || i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, padding: '12px 20px', borderBottom: i < sessionsSorted.length - 1 ? '1px solid #F3F4F6' : 'none' }}>
                      <div>
                        <div style={{ fontWeight: 600, fontSize: 14, color: '#111827' }}>{formatTanggalLengkap(s.tanggal)} — {labelSesi(s.sesi)}{s.kelas ? ` • ${s.kelas}` : ''}</div>
                        <div style={{ fontSize: 12, color: '#6B7280' }}>
                          {(s.waktuMulai || '')}{s.waktuSelesai ? `–${s.waktuSelesai}` : ''}{s.materi ? ` • ${s.materi}` : ''}{s.isBadal ? ' • Badal' : ''}
                        </div>
                      </div>
                      <div style={{ textAlign: 'right', flexShrink: 0 }}>
                        {Number(s.terlambatMenit) > 0 ? (
                          <>
                            <div style={{ fontSize: 13, fontWeight: 700, color: '#B91C1C' }}>Terlambat {s.terlambatMenit} menit</div>
                            <a href={tabayunLink(s)} target="_blank" rel="noreferrer" style={{ fontSize: 13, color: '#2563EB', fontWeight: 600 }}>Ajukan Tabayun</a>
                          </>
                        ) : (
                          <div style={{ fontSize: 13, fontWeight: 600, color: '#059669' }}>Tepat waktu</div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}
          </>
        )}
      </main>
    </div>
  )
}
