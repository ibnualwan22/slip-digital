const formatRupiah = (number) => {
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
        maximumFractionDigits: 0
    }).format(number);
};

const formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric'
    }).format(date);
};

const getMonthName = (monthNumber) => {
    const months = [
        'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
        'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ];
    return months[monthNumber - 1] || '-';
};

const getCategoryLabel = (code) => {
    const categories = {
        'BUK': 'BUK',
        'S2_GEL1': 'S2 Gelombang 1',
        'S2_GEL2': 'S2 Gelombang 2',
        'REGULER': 'Reguler',
        'KSU': 'KSU'
    };
    return categories[code] || code;
};

const getCategoryBadge = (categoryObjOrCode) => {
    if (!categoryObjOrCode) return `<span class="badge badge-gray">-</span>`;
    
    let code = categoryObjOrCode;
    let name = categoryObjOrCode;
    
    if (typeof categoryObjOrCode === 'object') {
        code = categoryObjOrCode.code;
        name = categoryObjOrCode.name;
    } else {
        name = getCategoryLabel(code);
    }

    const colors = {
        'BUK': 'badge-primary',
        'S2_GEL1': 'badge-success',
        'S2_GEL2': 'badge-success',
        'REGULER': 'badge-warning',
        'KSU': 'badge-danger'
    };
    return `<span class="badge ${colors[code] || 'badge-gray'}">${name || '-'}</span>`;
};

const getStatusBadge = (status) => {
    const config = {
        'DRAFT': { text: 'Draft', class: 'badge-warning' },
        'CONFIRMED': { text: 'Terkonfirmasi', class: 'badge-primary' },
        'PAID': { text: 'Dibayar', class: 'badge-success' }
    };
    const c = config[status] || { text: status, class: 'badge-gray' };
    return `<span class="badge ${c.class}">${c.text}</span>`;
};

const getActivityTypeBadge = (type) => {
    if (type === 'ADDITION') return `<span class="badge badge-success">Penambahan</span>`;
    if (type === 'DEDUCTION') return `<span class="badge badge-danger">Potongan</span>`;
    return type;
}
