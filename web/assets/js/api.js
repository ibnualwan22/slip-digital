const API_BASE = '/api/v1';

const api = {
    async request(endpoint, method = 'GET', data = null) {
        try {
            const options = {
                method,
                headers: {
                    'Content-Type': 'application/json'
                }
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
        } catch (error) {
            Swal.fire({
                icon: 'error',
                title: 'Oops...',
                text: error.message,
            });
            throw error;
        }
    },

    get(endpoint) {
        return this.request(endpoint, 'GET');
    },

    post(endpoint, data) {
        return this.request(endpoint, 'POST', data);
    },

    put(endpoint, data) {
        return this.request(endpoint, 'PUT', data);
    },

    delete(endpoint) {
        return this.request(endpoint, 'DELETE');
    }
};
