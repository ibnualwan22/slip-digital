async function renderPayrollDetail(container, params) {
    container.innerHTML = `<div class="loading-state"><i class='bx bx-loader-alt bx-spin'></i><p>Memuat data slip gaji...</p></div>`;

    try {
        const res = await api.get(`/payroll/${params.id}`);
        const tx = res.data;
        const details = tx.details || [];

        const additions = details.filter(d => d.type === 'ADDITION');
        const deductions = details.filter(d => d.type === 'DEDUCTION');

        container.innerHTML = `
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom: 20px;">
                <a href="#/payroll" class="btn btn-outline"><i class='bx bx-left-arrow-alt'></i> Kembali</a>
                <div>
                    <button class="btn btn-outline" onclick="loadPayrollDetailData('${tx.id}')"><i class='bx bx-refresh'></i> Refresh</button>
                    ${tx.status !== 'DRAFT' ? `<button class="btn btn-primary" style="background:#25D366;border-color:#25D366" onclick="previewAndSendSlip('${tx.id}')"><i class='bx bxl-whatsapp'></i> Kirim via WA</button>` : `<button class="btn btn-primary" onclick="calculateTHP('${tx.id}')"><i class='bx bx-calculator'></i> Hitung THP</button>`}
                </div>
            </div>

            <div class="info-grid">
                <div class="info-item">
                    <div class="label">Nama Asatidz</div>
                    <div class="value">${tx.employee?.name || '-'}</div>
                </div>
                <div class="info-item">
                    <div class="label">Kategori</div>
                    <div class="value">${tx.employee?.category?.name || getCategoryBadge(tx.employee?.category)}</div>
                </div>
                <div class="info-item">
                    <div class="label">Periode</div>
                    <div class="value">${getMonthName(tx.month)} ${tx.year}</div>
                </div>
                <div class="info-item">
                    <div class="label">Status</div>
                    <div class="value">${getStatusBadge(tx.status)}</div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Rincian Pendapatan (Income)</h3>
                    ${tx.status === 'DRAFT' ? `<button class="btn btn-sm btn-outline" onclick="showAddDetailForm('${tx.id}', 'ADDITION', '${tx.employee?.category?.calc_method || ''}', ${tx.employee?.category?.target_incentive || 0}, ${tx.employee?.hourly_rate || tx.employee?.category?.hourly_rate || 0})"><i class='bx bx-plus'></i> Tambah Item</button>` : ''}
                </div>
                <div class="card-body p-0">
                    <table class="table">
                        <thead>
                            <tr>
                                <th>Deskripsi / Aktivitas</th>
                                <th>Qty</th>
                                <th>Rate</th>
                                <th style="text-align:right">Total</th>
                                <th>Waktu Catat</th>
                                ${tx.status === 'DRAFT' ? '<th style="width:50px"></th>' : ''}
                            </tr>
                        </thead>
                        <tbody>
                            ${additions.length === 0 ? '<tr><td colspan="4" style="text-align:center">Belum ada data</td></tr>' : additions.map(a => `
                                <tr>
                                    <td>${a.description || a.activity?.activity_name || '-'}</td>
                                    <td>${a.quantity}</td>
                                    <td>${formatRupiah(a.rate)}</td>
                                    <td style="text-align:right; font-weight:600">${formatRupiah(a.total_amount)}</td>
                                    <td style="font-size:12px; color:var(--text-muted)">${a.recorded_at || '-'}</td>
                                    ${tx.status === 'DRAFT' ? `
                                        <td>
                                            <button class="btn-icon delete" onclick="removeDetail('${a.id}')" title="Hapus"><i class='bx bx-x'></i></button>
                                        </td>
                                    ` : ''}
                                </tr>
                            `).join('')}
                        </tbody>
                        ${additions.length > 0 ? `
                        <tfoot>
                            <tr style="background:#F8FAFC">
                                <td colspan="3" style="text-align:right; font-weight:600">Total Pendapatan:</td>
                                <td style="text-align:right; font-weight:700; color:var(--success)">${formatRupiah(tx.gross_income)}</td>
                                <td></td>
                                ${tx.status === 'DRAFT' ? '<td></td>' : ''}
                            </tr>
                        </tfoot>` : ''}
                    </table>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Rincian Potongan (Deduction)</h3>
                    ${tx.status === 'DRAFT' ? `<button class="btn btn-sm btn-outline" onclick="showAddDetailForm('${tx.id}', 'DEDUCTION', '${tx.employee?.category?.calc_method || ''}', ${tx.employee?.category?.target_incentive || 0}, ${tx.employee?.hourly_rate || tx.employee?.category?.hourly_rate || 0})"><i class='bx bx-plus'></i> Tambah Item</button>` : ''}
                </div>
                <div class="card-body p-0">
                    <table class="table">
                        <thead>
                            <tr>
                                <th>Deskripsi / Aktivitas</th>
                                <th>Qty</th>
                                <th>Rate</th>
                                <th style="text-align:right">Total</th>
                                <th>Waktu Catat</th>
                                ${tx.status === 'DRAFT' ? '<th style="width:50px"></th>' : ''}
                            </tr>
                        </thead>
                        <tbody>
                            ${deductions.length === 0 ? '<tr><td colspan="4" style="text-align:center">Belum ada data</td></tr>' : deductions.map(d => `
                                <tr>
                                    <td>${d.description || d.activity?.activity_name || '-'}</td>
                                    <td>${d.quantity}</td>
                                    <td>${formatRupiah(d.rate)}</td>
                                    <td style="text-align:right; font-weight:600">${formatRupiah(d.total_amount)}</td>
                                    <td style="font-size:12px; color:var(--text-muted)">${d.recorded_at || '-'}</td>
                                    ${tx.status === 'DRAFT' ? `
                                        <td>
                                            <button class="btn-icon delete" onclick="removeDetail('${d.id}')" title="Hapus"><i class='bx bx-x'></i></button>
                                        </td>
                                    ` : ''}
                                </tr>
                            `).join('')}
                        </tbody>
                        ${deductions.length > 0 ? `
                        <tfoot>
                            <tr style="background:#F8FAFC">
                                <td colspan="3" style="text-align:right; font-weight:600">Total Potongan:</td>
                                <td style="text-align:right; font-weight:700; color:var(--danger)">${formatRupiah(tx.total_deductions)}</td>
                                <td></td>
                                ${tx.status === 'DRAFT' ? '<td></td>' : ''}
                            </tr>
                        </tfoot>` : ''}
                    </table>
                </div>
            </div>

            <div class="card" style="background:var(--primary-light); border-color:#BFDBFE">
                <div class="card-body" style="display:flex; justify-content:space-between; align-items:center;">
                    <div>
                        <h2 style="color:var(--primary-hover); margin-bottom:4px">Take Home Pay</h2>
                        <p style="color:var(--primary-color); font-size:13px">Total yang diterima setelah dikurangi potongan</p>
                    </div>
                    <div style="font-size:32px; font-weight:700; color:var(--primary-hover)">
                        ${formatRupiah(tx.take_home_pay)}
                    </div>
                </div>
            </div>
        `;

    } catch (e) {
        container.innerHTML = `<div class="card"><div class="card-body" style="color:var(--danger)">Gagal memuat detail transaksi: ${e.message}</div></div>`;
    }
}

// Global helper to reload current route
window.loadPayrollDetailData = function (id) {
    handleRoute();
}

let allActivitiesCache = [];

async function showAddDetailForm(txId, type, calcMethod = '', targetIncentive = 0, hourlyRate = 0) {
    if (allActivitiesCache.length === 0) {
        try {
            const res = await api.get('/activities?active=true');
            allActivitiesCache = res.data || [];
        } catch (e) { }
    }

    const filteredActs = allActivitiesCache.filter(a => a.type === type);

    let htmlOptions = '<option value="" data-rate="0">-- Pilih Aktivitas --</option>';
    for (let a of filteredActs) {
        let defaultRate = 0; // Default 0 for all activities in the dropdown
        let rateText = "";

        if (a.activity_code === 'JAM_AJAR') {
            // Priority: employee hourly_rate override > category rate > 0
            if (calcMethod === 'PROPORTIONAL') {
                defaultRate = Math.round(targetIncentive / 70);
                rateText = ` (Target/70)`;
            } else if (calcMethod === 'HOURLY' && hourlyRate > 0) {
                defaultRate = hourlyRate;
                rateText = ` (Tarif Jam)`;
            }
            // else stays 0, user can input manually
        } else {
            // For non-JAM_AJAR, keep the activity's default rate
            defaultRate = a.default_rate || 0;
        }

        htmlOptions += `<option value="${a.id}" data-rate="${defaultRate}">${a.activity_name}${rateText}</option>`;
    }

    const { value: formValues } = await Swal.fire({
        title: `Tambah ${type === 'ADDITION' ? 'Pendapatan' : 'Potongan'}`,
        html: `
            <div class="form-group">
                <label class="form-label">Pilih Aktivitas</label>
                <select id="swal-det-act" class="form-control" onchange="
                    var rate = this.options[this.selectedIndex].dataset.rate || 0;
                    document.getElementById('swal-det-rate').value = formatRupiahInput(rate);
                ">
                    ${htmlOptions}
                </select>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label class="form-label">Quantity / Jumlah Jam</label>
                    <input type="number" id="swal-det-qty" class="form-control" value="1" min="0" step="0.5">
                </div>
                <div class="form-group">
                    <label class="form-label">Rate (Rp)</label>
                    <input type="text" id="swal-det-rate" class="form-control" value="0"
                        oninput="this.value = formatRupiahInput(this.value.replace(/\\./g, ''))">
                </div>
            </div>
            <div class="form-group">
                <label class="form-label">Deskripsi Tambahan (Opsional)</label>
                <input id="swal-det-desc" class="form-control">
            </div>
        `,
        focusConfirm: false,
        showCancelButton: true,
        confirmButtonText: 'Tambahkan',
        cancelButtonText: 'Batal',
        confirmButtonColor: 'var(--primary-color)',
        preConfirm: () => {
            const actId = document.getElementById('swal-det-act').value;
            if (!actId) {
                Swal.showValidationMessage('Silakan pilih aktivitas');
                return false;
            }
            return {
                activity_id: actId,
                quantity: parseFloat(document.getElementById('swal-det-qty').value),
                rate: parseFloat(document.getElementById('swal-det-rate').value.replace(/\./g, '').replace(',', '.')) || 0,
                type: type,
                description: document.getElementById('swal-det-desc').value
            }
        }
    });

    if (formValues) {
        try {
            await api.post(`/payroll/${txId}/details`, formValues);
            Swal.fire({
                title: 'Berhasil',
                text: 'Item ditambahkan',
                icon: 'success',
                timer: 1500,
                showConfirmButton: false
            });
            handleRoute(); // reload page
        } catch (e) { }
    }
}

async function calculateTHP(txId) {
    try {
        Swal.fire({
            title: 'Menghitung...',
            text: 'Sedang mengkalkulasi komponen gaji',
            allowOutsideClick: false,
            didOpen: () => {
                Swal.showLoading();
            }
        });

        await api.post(`/payroll/${txId}/calculate`);

        Swal.fire({
            title: 'Berhasil',
            text: 'Kalkulasi THP berhasil disimpan',
            icon: 'success',
            timer: 1500,
            showConfirmButton: false
        });

        handleRoute(); // reload page
    } catch (e) { }
}

async function removeDetail(detailId, txId) {
    const result = await Swal.fire({
        title: 'Hapus Item?',
        text: "Item rincian ini akan dihapus dari slip gaji.",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: 'var(--danger)',
        cancelButtonColor: 'var(--text-muted)',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    });

    if (result.isConfirmed) {
        try {
            await api.delete(`/payroll/details/${detailId}`);

            // Auto recalculate after delete
            await api.post(`/payroll/${txId}/calculate`);

            Swal.fire({
                title: 'Terhapus!',
                text: 'Item berhasil dihapus.',
                icon: 'success',
                timer: 1500,
                showConfirmButton: false
            });
            handleRoute(); // reload page
        } catch (e) {
            // err
        }
    }
}

async function previewAndSendSlip(id) {
    try {
        Swal.fire({
            title: 'Membuat Preview...',
            allowOutsideClick: false,
            didOpen: () => Swal.showLoading()
        });

        const res = await api.get('/payroll/' + id + '/preview-wa');
        const b64 = res.data.image_base64;

        const { isConfirmed } = await Swal.fire({
            title: 'Preview Slip Gaji',
            html: `<img src="data:image/jpeg;base64,${b64}" style="width:100%;max-width:500px;border:1px solid #ccc;border-radius:8px;">`,
            width: '600px',
            showCancelButton: true,
            confirmButtonText: '<i class="bx bxl-whatsapp"></i> Kirim via WhatsApp',
            cancelButtonText: 'Batal',
            confirmButtonColor: '#25D366'
        });

        if (isConfirmed) {
            Swal.fire({
                title: 'Mengirim Pesan...',
                allowOutsideClick: false,
                didOpen: () => Swal.showLoading()
            });

            await api.post('/payroll/' + id + '/send-wa');
            Swal.fire('Terkirim!', 'Slip berhasil dikirim ke WhatsApp pegawai.', 'success');
        }
    } catch (e) {
        // error already handled by api.js
    }
}
