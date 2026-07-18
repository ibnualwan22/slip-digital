export const formatRupiah = (number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(number || 0);
};

export const formatRupiahInput = (value) => {
  const cleanValue = String(value).replace(/\./g, '').trim();
  if (cleanValue === '') return '';
  const num = parseFloat(cleanValue);
  if (isNaN(num)) return '';
  return num.toLocaleString('id-ID');
};

export const parseRupiahInput = (value) => {
  return parseFloat(String(value).replace(/\./g, '').replace(',', '.')) || 0;
};

export const formatDate = (dateString) => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(date);
};

export const getMonthName = (monthNumber) => {
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
  ];
  return months[monthNumber - 1] || '-';
};

export const getCategoryBadgeClass = (code) => {
  const colors = {
    BUK: 'badge-primary',
    S2_GEL1: 'badge-success',
    S2_GEL2: 'badge-success',
    REGULER: 'badge-warning',
    KSU: 'badge-danger',
  };
  return colors[code] || 'badge-gray';
};

export const getStatusConfig = (status) => {
  const config = {
    DRAFT: { text: 'Draft', cls: 'badge-warning' },
    CONFIRMED: { text: 'Terkonfirmasi', cls: 'badge-primary' },
    PAID: { text: 'Dibayar', cls: 'badge-success' },
  };
  return config[status] || { text: status, cls: 'badge-gray' };
};
