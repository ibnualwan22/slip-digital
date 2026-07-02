let activitiesData = [];

async function renderActivities(container) {
    container.innerHTML = `
        <div class="card">
            <div class="card-header">
                <h3 class="card-title">Master Aktivitas</h3>
                <button class="btn btn-primary" onclick="showActivityForm()">
                    <i class='bx bx-plus'></i> Tambah Aktivitas
                </button>
            </div>
            <div class="card-body p-0">
                <div class="table-responsive">
                    <table class="table" id="act-table">
                        <thead>
                            <tr>
                                <th>Nama Aktivitas</th>
                                <th>Kode</th>
                                <th>Tipe</th>
                                <th>Rate Default</th>
                                <th>Deskripsi</th>
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
    await loadActivities();
}

async function loadActivities() {
    const tbody = document.querySelector('#act-table tbody');
    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center">Memuat data...</td></tr>';
    
    try {
        const res = await api.get('/activities');
        activitiesData = res.data || [];
        
        if (activitiesData.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align:center">Tidak ada aktivitas</td></tr>';
            return;
        }

        tbody.innerHTML = activitiesData.map(a => `
            <tr>
                <td style="font-weight:600">${a.activity_name}</td>
                <td><span style="font-family:monospace; background:#F1F5F9; padding:2px 6px; border-radius:4px">${a.code}</span></td>
                <td>${getActivityTypeBadge(a.type)}</td>
                <td>${formatRupiah(a.default_rate)}</td>
                <td>${a.description || '-'}</td>
                <td>
                    ${a.is_active ? 
                        '<span class="badge badge-success">Aktif</span>' : 
                        '<span class="badge badge-danger">Nonaktif</span>'}
                </td>
                <td>
                    <button class="btn-icon" onclick="showActivityForm('${a.id}')" title="Edit"><i class='bx bx-edit-alt'></i></button>
                </td>
            </tr>
        `).join('');
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:red">Gagal memuat data</td></tr>';
    }
}

async function showActivityForm(id = null) {
    let act = {
        activity_name: '', code: '', type: 'ADDITION', default_rate: 0, description: '', is_active: true
    };
    let title = 'Tambah Master Aktivitas';

    if (id) {
        act = activitiesData.find(a => a.id === id);
        title = 'Edit Master Aktivitas';
    }

    const { value: formValues } = await Swal.fire({
        title: title,
        html: `
            <div class="form-group">
                <label class="form-label">Nama Aktivitas</label>
                <input id="swal-act-name" class="form-control" placeholder="Cth: Jam Mengajar" value="${act.activity_name}">
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label class="form-label">Kode (Unik)</label>
                    <input id="swal-act-code" class="form-control" placeholder="Cth: JAM_AJAR" style="text-transform:uppercase" value="${act.code}" ${id ? 'readonly' : ''}>
                </div>
                <div class="form-group">
                    <label class="form-label">Tipe</label>
                    <select id="swal-act-type" class="form-control">
                        <option value="ADDITION" ${act.type=='ADDITION'?'selected':''}>Penambahan (Income)</option>
                        <option value="DEDUCTION" ${act.type=='DEDUCTION'?'selected':''}>Potongan (Deduction)</option>
                    </select>
                </div>
            </div>
            <div class="form-group">
                <label class="form-label">Rate Default (Rp)</label>
                <input type="number" id="swal-act-rate" class="form-control" value="${act.default_rate}">
            </div>
            <div class="form-group">
                <label class="form-label">Deskripsi</label>
                <input id="swal-act-desc" class="form-control" value="${act.description || ''}">
            </div>
            <div class="form-group">
                <label style="display:flex; align-items:center; gap:8px; justify-content:start;">
                    <input type="checkbox" id="swal-act-active" ${act.is_active?'checked':''}> Status Aktif
                </label>
            </div>
        `,
        focusConfirm: false,
        showCancelButton: true,
        confirmButtonText: 'Simpan',
        cancelButtonText: 'Batal',
        confirmButtonColor: 'var(--primary-color)',
        preConfirm: () => {
            return {
                activity_name: document.getElementById('swal-act-name').value,
                code: document.getElementById('swal-act-code').value.toUpperCase(),
                type: document.getElementById('swal-act-type').value,
                default_rate: parseFloat(document.getElementById('swal-act-rate').value),
                description: document.getElementById('swal-act-desc').value,
                is_active: document.getElementById('swal-act-active').checked
            }
        }
    });

    if (formValues) {
        try {
            if (id) {
                await api.put(`/activities/${id}`, formValues);
                Swal.fire('Tersimpan', 'Aktivitas berhasil diupdate', 'success');
            } else {
                await api.post('/activities', formValues);
                Swal.fire('Tersimpan', 'Aktivitas baru berhasil ditambahkan', 'success');
            }
            loadActivities();
        } catch (e) {
            // err handled by api
        }
    }
}
