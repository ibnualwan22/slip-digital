const API_BASE = '/api/v1';
export const TOKEN_KEY = 'emaliyah_token';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function getRole() {
  return localStorage.getItem('emaliyah_role');
}

export function getName() {
  return localStorage.getItem('emaliyah_name') || '';
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem('emaliyah_role');
  localStorage.removeItem('emaliyah_name');
}

function handleUnauthorized() {
  clearSession();
  if (window.location.pathname !== '/login') {
    window.location.href = '/login';
  }
}

async function request(endpoint, method = 'GET', data = null) {
  const options = {
    method,
    headers: { 'Content-Type': 'application/json' },
    cache: 'no-store',
  };
  const token = getToken();
  if (token) options.headers['Authorization'] = `Bearer ${token}`;
  if (data) {
    options.body = JSON.stringify(data);
  }

  const response = await fetch(`${API_BASE}${endpoint}`, options);
  if (response.status === 401) {
    handleUnauthorized();
    throw new Error('Sesi berakhir. Silakan login kembali.');
  }
  const result = await response.json();

  if (!result.success) {
    throw new Error(result.message || 'Terjadi kesalahan pada server');
  }
  return result;
}

const api = {
  get: (endpoint) => request(endpoint, 'GET'),
  post: (endpoint, data) => request(endpoint, 'POST', data),
  put: (endpoint, data) => request(endpoint, 'PUT', data),
  delete: (endpoint) => request(endpoint, 'DELETE'),
  upload: async (endpoint, formData) => {
    const headers = {};
    const token = getToken();
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const response = await fetch(`${API_BASE}${endpoint}`, {
      method: 'POST',
      headers,
      body: formData,
      cache: 'no-store',
    });
    if (response.status === 401) {
      handleUnauthorized();
      throw new Error('Sesi berakhir. Silakan login kembali.');
    }
    const result = await response.json();
    if (!result.success) throw new Error(result.message || 'Upload gagal');
    return result;
  },
};

export default api;
