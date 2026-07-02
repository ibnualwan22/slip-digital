// Sidebar Toggle Logic
document.getElementById('toggle-sidebar').addEventListener('click', () => {
    document.getElementById('sidebar').classList.add('active');
});
document.getElementById('close-sidebar').addEventListener('click', () => {
    document.getElementById('sidebar').classList.remove('active');
});

// Simple Hash Router
const routes = {
    '#/dashboard': { title: 'Dashboard', render: renderDashboard },
    '#/categories': { title: 'Master Kategori Asatidz', render: renderCategories },
    '#/employees': { title: 'Data Asatidz', render: renderEmployees },
    '#/activities': { title: 'Master Aktivitas', render: renderActivities },
    '#/payroll': { title: 'Transaksi Payroll', render: renderPayroll },
};

const contentEl = document.getElementById('app-content');
const titleEl = document.getElementById('page-title');

async function handleRoute() {
    let hash = window.location.hash || '#/dashboard';

    // Check dynamic routes
    let matchedRoute = routes[hash];
    let params = {};

    if (!matchedRoute) {
        if (hash.startsWith('#/payroll/')) {
            matchedRoute = { title: 'Detail Slip Gaji', render: renderPayrollDetail };
            params.id = hash.replace('#/payroll/', '');
        }
    }

    if (!matchedRoute) {
        hash = '#/dashboard';
        matchedRoute = routes[hash];
    }

    // Update UI
    titleEl.textContent = matchedRoute.title;

    // Update active menu
    document.querySelectorAll('.menu-item').forEach(el => el.classList.remove('active'));

    let menuId = hash.split('/')[1]; // 'dashboard', 'employees', etc
    const activeMenu = document.getElementById(`menu-${menuId}`);
    if (activeMenu) activeMenu.classList.add('active');

    // Render Page
    contentEl.innerHTML = `
        <div class="loading-state">
            <i class='bx bx-loader-alt bx-spin'></i>
            <p>Memuat...</p>
        </div>
    `;

    try {
        await matchedRoute.render(contentEl, params);
    } catch (e) {
        contentEl.innerHTML = `
            <div class="card"><div class="card-body" style="color:var(--danger)">
                Gagal memuat halaman: ${e.message}
            </div></div>
        `;
    }
}

// Listen to hash changes
window.addEventListener('hashchange', handleRoute);

// Initialize
handleRoute();
