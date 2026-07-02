let employeesData = [];

async function renderEmployees(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header">
                <div class="filter-bar" style="margin-bottom:0">
                    <select id="filter-cat" class="form-control" style="width:200px">
                        <option value="">Semua Kategori (Loading...)</option>
                    </select>
                    <label style="display:flex; align-items:center; gap:8px; cursor:pointer;">
                        <input type="checkbox" id="filter-active" checked> Aktif Saja
                    </label>
                </div>
                <button class="btn btn-primary" onclick="showEmployeeForm()">
                    <i class='bx bx-plus'></i> Tambah Pegawai
                </button>
            </div>
            <div class="card-body p-0">
                <div class="table-responsive">
                    <table class="table" id="emp-table">
                        <thead>
                            <tr>
                                <th>Nama</th>
                                <th>Kategori</th>
                                <th>Jabatan</th>
                                <th>Gaji Pokok</th>
                                <th>Status</th>
                                <th style="width:100px">Aksi</th>
                            </tr>
                        </thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
    `;

    document.getElementById('filter-cat').addEventListener('change', loadEmployees);
    document.getElementById('filter-active').addEventListener('change', loadEmployees);

    // Fetch categories for filter
    try {
        const res = await api.get('/categories');
        const filterCat = document.getElementById('filter-cat');
        filterCat.innerHTML = '<option value="">Semua Kategori</option>' + res.data.map(c => `<option value="${c.id}">${c.name}</option>`).join('');
    } catch(e) {}

    await loadEmployees();
}

async function loadEmployees() {
    const tbody = document.querySelector('#emp-table tbody');
    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center">Memuat data...</td></tr>';
    
    const cat = document.getElementById('filter-cat').value;
    const active = document.getElementById('filter-active').checked;
    
    let url = `/employees?active=${active}`;
    if (cat) url += `&category=${cat}`;

    try {
        const res = await api.get(url);
        employeesData = res.data || [];
        
        if (employeesData.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" style="text-align:center">Tidak ada data pegawai</td></tr>';
            return;
        }

        tbody.innerHTML = employeesData.map(emp => `
            <tr>
                <td style="font-weight:600">${emp.name}</td>
                <td>${getCategoryBadge(emp.category)}</td>
                <td>${emp.role || '-'}</td>
                <td>${formatRupiah(emp.category?.fixed_salary || 0)}</td>
                <td>
                    ${emp.is_active ? 
                        '<span class="badge badge-success">Aktif</span>' : 
                        '<span class="badge badge-danger">Nonaktif</span>'}
                </td>
                <td>
                    <button class="btn-icon" onclick="showEmployeeForm('${emp.id}')" title="Edit"><i class='bx bx-edit-alt'></i></button>
                    <button class="btn-icon delete" onclick="deleteEmployee('${emp.id}')" title="Hapus"><i class='bx bx-trash'></i></button>
                </td>
            </tr>
        `).join('');
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;color:red">Gagal memuat data</td></tr>';
    }
}

async function showEmployeeForm(id = null) {
    let emp = {
        name: '', category_id: '', role: '', 
        is_active: true
    };
    let title = 'Tambah Pegawai Baru';

    if (id) {
        emp = employeesData.find(e => e.id === id);
        title = 'Edit Pegawai';
    }

    // load categories
    let cats = [];
    try {
        const res = await api.get('/categories');
        cats = res.data;
    } catch(e) {}

    const { value: formValues } = await Swal.fire({
        title: title,
        html: `
            <div class="form-group">
                <label class="form-label">Nama Pegawai</label>
                <input id="swal-name" class="form-control" value="${emp.name}">
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label class="form-label">Kategori</label>
                    <select id="swal-cat" class="form-control">
                        <option value="">-- Pilih Kategori --</option>
                        ${cats.map(c => `<option value="${c.id}" ${c.id===emp.category_id?'selected':''}>${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Jabatan (Role)</label>
                    <input id="swal-role" class="form-control" value="${emp.role}">
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label class="form-label">Tunjangan Khusus <span style="font-size:11px;color:var(--text-muted)">(opsional, override kategori)</span></label>
                    <input type="number" id="swal-allowance" class="form-control" value="${emp.structural_allowance || ''}" placeholder="Kosongkan = ikut kategori">
                </div>
                <div class="form-group">
                    <label class="form-label">Tarif Jam Mengajar <span style="font-size:11px;color:var(--text-muted)">(opsional, override kategori)</span></label>
                    <input type="number" id="swal-hourly" class="form-control" value="${emp.hourly_rate || ''}" placeholder="Kosongkan = ikut kategori">
                </div>
            </div>
            <div class="form-group" style="margin-top: 15px;">
                <label style="display:flex; align-items:center; gap:8px; justify-content:start;">
                    <input type="checkbox" id="swal-active" ${emp.is_active?'checked':''}> Status Aktif
                </label>
            </div>
        `,
        focusConfirm: false,
        showCancelButton: true,
        confirmButtonText: 'Simpan',
        cancelButtonText: 'Batal',
        confirmButtonColor: 'var(--primary-color)',
        preConfirm: () => {
            const catId = document.getElementById('swal-cat').value;
            if (!catId) {
                Swal.showValidationMessage('Silakan pilih kategori');
                return false;
            }
            const allowanceVal = document.getElementById('swal-allowance').value;
            const hourlyVal = document.getElementById('swal-hourly').value;
            return {
                name: document.getElementById('swal-name').value,
                category_id: catId,
                role: document.getElementById('swal-role').value,
                structural_allowance: allowanceVal !== '' ? parseFloat(allowanceVal) : null,
                hourly_rate: hourlyVal !== '' ? parseFloat(hourlyVal) : null,
                is_active: document.getElementById('swal-active').checked
            }
        }
    });

    if (formValues) {
        try {
            if (id) {
                await api.put(`/employees/${id}`, formValues);
                Swal.fire('Tersimpan', 'Data pegawai berhasil diupdate', 'success');
            } else {
                await api.post('/employees', formValues);
                Swal.fire('Tersimpan', 'Pegawai baru berhasil ditambahkan', 'success');
            }
            loadEmployees();
        } catch (e) {
            // error handled by api
        }
    }
}

async function deleteEmployee(id) {
    const result = await Swal.fire({
        title: 'Hapus Pegawai?',
        text: "Data yang dihapus tidak dapat dikembalikan!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: 'var(--danger)',
        cancelButtonColor: 'var(--text-muted)',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    });

    if (result.isConfirmed) {
        try {
            await api.delete(`/employees/${id}`);
            Swal.fire('Terhapus!', 'Data pegawai berhasil dihapus.', 'success');
            loadEmployees();
        } catch (e) {
            // error handled by api
        }
    }
}
