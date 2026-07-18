const API_BASE = '/api/v1';

async function request(endpoint, method = 'GET', data = null) {
  const options = {
    method,
    headers: { 'Content-Type': 'application/json' },
    cache: 'no-store',
  };
  if (data) {
    options.body = JSON.stringify(data);
  }

  const response = await fetch(`${API_BASE}${endpoint}`, options);
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
    const response = await fetch(`${API_BASE}${endpoint}`, {
      method: 'POST',
      body: formData,
      cache: 'no-store',
    });
    const result = await response.json();
    if (!result.success) throw new Error(result.message || 'Upload gagal');
    return result;
  },
};

export default api;
