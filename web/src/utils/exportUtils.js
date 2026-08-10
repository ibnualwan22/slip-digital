import * as xlsx from 'xlsx'
import { jsPDF } from 'jspdf'
import autoTable from 'jspdf-autotable'

// Helper to format currency
const formatRp = (num) => {
  const n = parseFloat(num) || 0
  return 'Rp ' + n.toLocaleString('id-ID')
}

/**
 * Loads an image from a URL and converts it to a base64 Data URL.
 */
const getBase64Image = (url) => {
  return new Promise((resolve) => {
    if (!url) {
      resolve(null)
      return
    }
    const img = new Image()
    img.crossOrigin = 'Anonymous'
    img.onload = () => {
      const canvas = document.createElement('canvas')
      // Scale down image to fit PDF nicely (max width 400px)
      const scale = Math.min(1, 400 / img.width)
      canvas.width = img.width * scale
      canvas.height = img.height * scale
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
      try {
        const dataUrl = canvas.toDataURL('image/jpeg', 0.7)
        resolve(dataUrl)
      } catch (err) {
        console.error('Canvas export error (CORS might block it):', err)
        resolve(null)
      }
    }
    img.onerror = () => {
      console.error('Failed to load image for PDF:', url)
      resolve(null)
    }
    img.src = url
  })
}

export const exportToExcel = (report, items) => {
  if (!items || items.length === 0) return

  // Format data for Excel
  const excelData = items.map((item, idx) => ({
    'No': idx + 1,
    'Tanggal': item.tanggal ? new Date(item.tanggal).toLocaleDateString('id-ID') : '-',
    'Nama Barang': item.nama_barang || '-',
    'Harga Satuan': parseFloat(item.harga_satuan) || 0,
    'Jumlah': parseInt(item.jumlah) || 1,
    'Total Harga': (parseFloat(item.harga_satuan) || 0) * (parseInt(item.jumlah) || 1),
    'Kredit': parseFloat(item.kredit) || 0,
    'Debit': parseFloat(item.debit) || 0,
    'Keterangan': item.keterangan || '-',
    'Status': item.is_tashih ? 'Tervalidasi' : 'Belum'
  }))

  const ws = xlsx.utils.json_to_sheet(excelData)

  // Auto-width columns
  const colWidths = [
    { wch: 5 }, { wch: 15 }, { wch: 40 }, { wch: 15 }, { wch: 10 },
    { wch: 15 }, { wch: 15 }, { wch: 15 }, { wch: 20 }, { wch: 15 }
  ]
  ws['!cols'] = colWidths

  const wb = xlsx.utils.book_new()
  xlsx.utils.book_append_sheet(wb, ws, 'Pengeluaran')

  const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  const mName = monthNames[report.month - 1]
  const fileName = `Laporan_Pengeluaran_${mName}_${report.year}.xlsx`

  xlsx.writeFile(wb, fileName)
}

export const exportToPDF = async (report, items) => {
  if (!items || items.length === 0) return

  const doc = new jsPDF('p', 'pt', 'a4')

  const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  const mName = monthNames[report.month - 1]
  const reportTitle = `Laporan Pengeluaran - ${mName} ${report.year}`

  // Title
  doc.setFontSize(16)
  doc.text(reportTitle, 40, 40)

  if (report.keterangan) {
    doc.setFontSize(12)
    doc.setTextColor(100)
    doc.text(report.keterangan, 40, 58)
  }

  // Pre-load images
  // We only show one image per group (the first one) to avoid duplicates, 
  // or we map by item.id if each has one. 
  // The UI groups by group_id || id.
  const groupSizes = {}
  items.forEach(it => {
    const g = it.group_id || it.id
    groupSizes[g] = (groupSizes[g] || 0) + 1
  })

  // Group item images
  const loadedImages = {} // Keyed by item.id

  // Prepare table data
  let tableData = []
  let currentGroup = null
  let currentNo = 0

  for (let idx = 0; idx < items.length; idx++) {
    const item = items[idx]
    const itemGroupId = item.group_id || item.id
    const isFirstInGroup = itemGroupId !== currentGroup

    if (isFirstInGroup) {
      currentGroup = itemGroupId
      currentNo++
      if (item.bukti_pembayaran) {
        const imgBase64 = await getBase64Image(item.bukti_pembayaran)
        if (imgBase64) {
          loadedImages[itemGroupId] = imgBase64
        }
      }
    }

    const hs = parseFloat(item.harga_satuan) || 0
    const jml = parseInt(item.jumlah) || 1
    const total = hs * jml

    tableData.push([
      isFirstInGroup ? currentNo : '',
      item.tanggal ? new Date(item.tanggal).toLocaleDateString('id-ID') : '-',
      item.nama_barang || '-',
      formatRp(hs),
      jml,
      formatRp(total),
      isFirstInGroup ? formatRp(item.kredit) : '',
      isFirstInGroup ? formatRp(item.debit) : '',
      isFirstInGroup ? (loadedImages[itemGroupId] ? 'IMAGE_PLACEHOLDER' : (item.bukti_pembayaran ? 'Gagal Muat' : '-')) : ''
    ])
  }

  const totalsData = [
    [
      '', '', 'TOTAL KESELURUHAN', '', '',
      formatRp(items.reduce((acc, it) => acc + ((parseFloat(it.harga_satuan) || 0) * (parseInt(it.jumlah) || 1)), 0)),
      formatRp(items.reduce((acc, it) => acc + (parseFloat(it.kredit) || 0), 0)),
      formatRp(items.reduce((acc, it) => acc + (parseFloat(it.debit) || 0), 0)),
      ''
    ]
  ]

  // Render Table
  autoTable(doc, {
    startY: 80,
    head: [['No', 'Tanggal', 'Nama Barang', 'Harga/Pcs', 'Jml', 'Total', 'Kredit', 'Debit', 'Bukti']],
    body: [...tableData, ...totalsData],
    styles: { fontSize: 9, cellPadding: 4, valign: 'middle' },
    headStyles: { fillColor: [41, 128, 185], textColor: 255 },
    columnStyles: {
      0: { halign: 'center', cellWidth: 20 },
      1: { cellWidth: 50 },
      2: { cellWidth: 'auto' },
      3: { halign: 'right', cellWidth: 60 },
      4: { halign: 'center', cellWidth: 25 },
      5: { halign: 'right', cellWidth: 60 },
      6: { halign: 'right', cellWidth: 60 },
      7: { halign: 'right', cellWidth: 60 },
      8: { halign: 'center', cellWidth: 60 } // Bukti Image column
    },
    didParseCell: function (data) {
      // Find if this cell is our image placeholder
      if (data.section === 'body' && data.column.index === 8) {
        if (data.cell.raw === 'IMAGE_PLACEHOLDER') {
          // Minimum row height for images
          data.cell.styles.minCellHeight = 60;
        }
      }
      // Total row style
      if (data.section === 'body' && data.row.index === tableData.length) {
        data.cell.styles.fontStyle = 'bold';
        data.cell.styles.fillColor = [240, 240, 240];
      }
    },
    didDrawCell: function (data) {
      if (data.section === 'body' && data.column.index === 8 && data.cell.raw === 'IMAGE_PLACEHOLDER') {
        // Determine original item group id for this row... this can be tricky because we lost reference.
        // Let's use the item index or we can map it back. 
        // Good thing the row index corresponds to tableData index (for normal rows).
        if (data.row.index < items.length) {
          const originalItem = items[data.row.index];
          const gId = originalItem.group_id || originalItem.id;
          if (loadedImages[gId]) {
            const x = data.cell.x + 5;
            const y = data.cell.y + 5;
            const dim = 50; // Image box 50x50
            doc.addImage(loadedImages[gId], 'JPEG', x, y, dim, dim);
          }
        }
      }
    },
  })

  const fileName = `Laporan_PDF_Pengeluaran_${mName}_${report.year}.pdf`
  doc.save(fileName)
}
