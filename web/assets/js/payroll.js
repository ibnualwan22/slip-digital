async function renderPayroll(container) {
    const now = new Date();
    const currentMonth = now.getMonth() + 1;
    const currentYear = now.getFullYear();

    container.innerHTML = `
        <div class="card">
            <div class="card-header">
                <div class="filter-bar" style="margin-bottom:0">
                    <select id="filter-month" class="form-control" style="width:150px">
                        ${Array.from({length: 12}, (_, i) => `<option value="${i+1}" ${i+1===currentMonth?'selected':''}>${getMonthName(i+1)}</option>`).join('')}
                    </select>
                    <input type="number" id="filter-year" class="form-control" value="${currentYear}" style="width:100px">
                </div>
                <button class="btn btn-primary" onclick="showPayrollForm()">
                    <i class='bx bx-plus'></i> Buat Transaksi
                </button>
            </div>
            <div class="card-body p-0">
                <div class="table-responsive">
                    <table class="table" id="payroll-table">
                        <thead>
                            <tr>
                                <th>Pegawai</th>
                                <th>Periode</th>
                                <th>Pendapatan (Gross)</th>
                                <th>Potongan</th>
                                <th>Take Home Pay</th>
                                <th>Status</th>
                                <th style="width:80px">Aksi</th>
                            </tr>
                        </thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
    `;

    document.getElementById('filter-month').addEventListener('change', loadPayrolls);
    document.getElementById('filter-year').addEventListener('change', loadPayrolls);

    await loadPayrolls();
}

async function loadPayrolls() {
    const tbody = document.querySelector('#payroll-table tbody');
    tbody.innerHTML = '<tr><td colspan="7" style="text-align:center">Memuat data...</td></tr>';
    
    const month = document.getElementById('filter-month').value;
    const year = document.getElementById('filter-year').value;

    try {
        const res = await api.get(`/payroll?month=${month}&year=${year}`);
        const data = res.data || [];
        
        if (data.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align:center">Tidak ada transaksi pada periode ini</td></tr>';
            return;
        }

        tbody.innerHTML = data.map(p => `
            <tr>
                <td>
                    <div style="font-weight:600">${p.employee?.name || 'Unknown'}</div>
                    <div style="font-size:12px;color:var(--text-muted)">${p.employee?.category?.name || '-'}</div>
                </td>
                <td>${getMonthName(p.month)} ${p.year}</td>
                <td style="color:var(--success)">${formatRupiah(p.gross_income)}</td>
                <td style="color:var(--danger)">${formatRupiah(p.total_deductions)}</td>
                <td style="font-weight:700;color:var(--primary-color)">${formatRupiah(p.take_home_pay)}</td>
                <td>${getStatusBadge(p.status)}</td>
                <td>
                    <a href="#/payroll/${p.id}" class="btn-icon" title="Lihat Detail"><i class='bx bx-right-arrow-alt'></i></a>
                    <button class="btn-icon delete" onclick="deleteTransaction('${p.id}')" title="Hapus"><i class='bx bx-trash'></i></button>
                </td>
            </tr>
        `).join('');
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:red">Gagal memuat data</td></tr>';
    }
}

async function showPayrollForm() {
    // Fetch employees for dropdown
    const empRes = await api.get('/employees?active=true');
    const employees = empRes.data || [];
    
    if(employees.length === 0) {
        Swal.fire('Oops', 'Tidak ada data pegawai aktif', 'error');
        return;
    }

    const now = new Date();

    const { value: formValues } = await Swal.fire({
        title: 'Buat Transaksi Payroll Baru',
        html: `
            <div class="form-group">
                <label class="form-label">Pegawai</label>
                <select id="swal-pay-emp" class="form-control">
                    ${employees.map(e => `<option value="${e.id}">${e.name} (${e.category?.name || '-'})</option>`).join('')}
                </select>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label class="form-label">Bulan</label>
                    <select id="swal-pay-month" class="form-control">
                        ${Array.from({length: 12}, (_, i) => `<option value="${i+1}" ${i+1===now.getMonth()+1?'selected':''}>${getMonthName(i+1)}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Tahun</label>
                    <input type="number" id="swal-pay-year" class="form-control" value="${now.getFullYear()}">
                </div>
            </div>
            <div class="form-group">
                <label class="form-label">Catatan (Opsional)</label>
                <input id="swal-pay-notes" class="form-control">
            </div>
        `,
        focusConfirm: false,
        showCancelButton: true,
        confirmButtonText: 'Buat Transaksi',
        cancelButtonText: 'Batal',
        confirmButtonColor: 'var(--primary-color)',
        preConfirm: () => {
            return {
                employee_id: document.getElementById('swal-pay-emp').value,
                month: parseInt(document.getElementById('swal-pay-month').value),
                year: parseInt(document.getElementById('swal-pay-year').value),
                notes: document.getElementById('swal-pay-notes').value
            }
        }
    });

    if (formValues) {
        try {
            const res = await api.post('/payroll', formValues);
            Swal.fire('Berhasil', 'Transaksi dibuat', 'success').then(() => {
                window.location.hash = `#/payroll/${res.data.id}`;
            });
        } catch (e) {
            // err handled
        }
    }
}

async function deleteTransaction(id) {
    const result = await Swal.fire({
        title: 'Hapus Transaksi?',
        text: "Seluruh detail rincian di dalam transaksi ini juga akan terhapus secara permanen!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: 'var(--danger)',
        cancelButtonColor: 'var(--text-muted)',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    });

    if (result.isConfirmed) {
        try {
            await api.delete(`/payroll/${id}`);
            Swal.fire('Terhapus!', 'Transaksi berhasil dihapus.', 'success');
            loadPayrolls();
        } catch (e) {
            // error handled by api
        }
    }
}
