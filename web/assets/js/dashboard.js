async function renderDashboard(container) {
    // Fetch data
    const empRes = await api.get('/employees');
    const employees = empRes.data || [];
    
    const now = new Date();
    const currentMonth = now.getMonth() + 1;
    const currentYear = now.getFullYear();
    
    const payRes = await api.get(`/payroll?month=${currentMonth}&year=${currentYear}`);
    const payrolls = payRes.data || [];

    const totalActive = employees.filter(e => e.is_active).length;
    const totalTransactions = payrolls.length;
    
    let totalExpense = 0;
    payrolls.forEach(p => {
        totalExpense += parseFloat(p.take_home_pay);
    });

    container.innerHTML = `
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-icon primary">
                    <i class='bx bx-group'></i>
                </div>
                <div class="stat-info">
                    <h3>${employees.length}</h3>
                    <p>Total Pegawai</p>
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-icon success">
                    <i class='bx bx-user-check'></i>
                </div>
                <div class="stat-info">
                    <h3>${totalActive}</h3>
                    <p>Pegawai Aktif</p>
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-icon warning">
                    <i class='bx bx-receipt'></i>
                </div>
                <div class="stat-info">
                    <h3>${totalTransactions}</h3>
                    <p>Slip Gaji (${getMonthName(currentMonth)} ${currentYear})</p>
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-icon danger">
                    <i class='bx bx-wallet'></i>
                </div>
                <div class="stat-info">
                    <h3>${formatRupiah(totalExpense)}</h3>
                    <p>Estimasi Pengeluaran</p>
                </div>
            </div>
        </div>
        
        <div class="card">
            <div class="card-header">
                <h3 class="card-title">Informasi Sistem</h3>
            </div>
            <div class="card-body">
                <p>Selamat datang di sistem E-Rekap Markaz Arabiyah. Gunakan menu di sebelah kiri untuk mengelola data pegawai, master aktivitas, dan transaksi payroll bulanan.</p>
            </div>
        </div>
    `;
}
