let categoriesData = [];

async function renderCategories(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header">
                <h3 class="card-title">Master Kategori Pegawai</h3>
                <button class="btn btn-primary" onclick="showCategoryForm()">
                    <i class='bx bx-plus'></i> Tambah Kategori
                </button>
            </div>
            <div class="card-body p-0">
                <div class="table-responsive">
                    <table class="table" id="cat-table">
                        <thead>
                            <tr>
                                <th>Nama Kategori</th>
                                <th>Kode</th>
                                <th>Gaji Pokok</th>
                                <th>Struktural</th>
                                <th>Insentif / Rate</th>
                                <th>Kalkulasi</th>
                                <th style="width:100px">Aksi</th>
                            </tr>
                        </thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
    `;

    await loadCategories();
}

async function loadCategories() {
    const tbody = document.querySelector('#cat-table tbody');
    tbody.innerHTML = '<tr><td colspan="7" style="text-align:center">Memuat data...</td></tr>';
    
    try {
        const res = await api.get('/categories');
        categoriesData = res.data || [];
        
        if (categoriesData.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align:center">Tidak ada data kategori</td></tr>';
            return;
        }

        tbody.innerHTML = categoriesData.map(cat => `
            <tr>
                <td style="font-weight:600">${cat.name}</td>
                <td><span class="badge" style="background:var(--gray-200);color:var(--text-color)">${cat.code}</span></td>
                <td>${formatRupiah(cat.fixed_salary)}</td>
                <td>${formatRupiah(cat.structural_allowance)}</td>
                <td>
                    ${cat.target_incentive > 0 ? `Target: ${formatRupiah(cat.target_incentive)}<br>` : ''}
                    ${cat.hourly_rate > 0 ? `Rate: ${formatRupiah(cat.hourly_rate)}` : ''}
                    ${cat.target_incentive == 0 && cat.hourly_rate == 0 ? '-' : ''}
                </td>
                <td>${cat.calc_method}</td>
                <td>
                    <button class="btn-icon" onclick="showCategoryForm('${cat.id}')" title="Edit"><i class='bx bx-edit-alt'></i></button>
                    <!-- <button class="btn-icon delete" onclick="deleteCategory('${cat.id}')" title="Hapus"><i class='bx bx-trash'></i></button> -->
                </td>
            </tr>
        `).join('');
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:red">Gagal memuat data</td></tr>';
    }
}

async function showCategoryForm(id = null) {
    let cat = {
        name: '', code: '', calc_method: 'FIXED', 
        fixed_salary: 0, structural_allowance: 0, 
        target_incentive: 0, hourly_rate: 0
    };
    let title = 'Tambah Kategori Baru';

    if (id) {
        cat = categoriesData.find(c => c.id === id);
        title = 'Edit Kategori';
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
                    <option value="FIXED" ${cat.calc_method=='FIXED'?'selected':''}>FIXED (Tanpa Rate)</option>
                    <option value="HOURLY" ${cat.calc_method=='HOURLY'?'selected':''}>HOURLY (Jam x Tarif Lokal)</option>
                    <option value="PROPORTIONAL" ${cat.calc_method=='PROPORTIONAL'?'selected':''}>PROPORTIONAL (Jam / 70 x Target)</option>
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
            const name = document.getElementById('swal-name').value;
            const code = document.getElementById('swal-code').value;
            if (!name || !code) {
                Swal.showValidationMessage('Nama dan Kode tidak boleh kosong');
                return false;
            }
            return {
                name: name,
                code: code,
                calc_method: document.getElementById('swal-calc').value,
                fixed_salary: parseFloat(document.getElementById('swal-salary').value) || 0,
                structural_allowance: parseFloat(document.getElementById('swal-allowance').value) || 0,
                hourly_rate: parseFloat(document.getElementById('swal-hourly').value) || 0,
                target_incentive: parseFloat(document.getElementById('swal-incentive').value) || 0
            }
        }
    });

    if (formValues) {
        try {
            if (id) {
                await api.put(`/categories/${id}`, formValues);
                Swal.fire('Tersimpan', 'Kategori berhasil diupdate', 'success');
            } else {
                await api.post('/categories', formValues);
                Swal.fire('Tersimpan', 'Kategori baru berhasil ditambahkan', 'success');
            }
            loadCategories();
        } catch (e) {
            Swal.fire('Gagal', e.message, 'error');
        }
    }
}

async function deleteCategory(id) {
    const result = await Swal.fire({
        title: 'Hapus Kategori?',
        text: "Kategori tidak dapat dihapus jika masih ada pegawai yang terhubung!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: 'var(--danger)',
        cancelButtonColor: 'var(--text-muted)',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    });

    if (result.isConfirmed) {
        try {
            await api.delete(`/categories/${id}`);
            Swal.fire('Terhapus!', 'Kategori berhasil dihapus.', 'success');
            loadCategories();
        } catch (e) {
            Swal.fire('Gagal', e.message, 'error');
        }
    }
}
