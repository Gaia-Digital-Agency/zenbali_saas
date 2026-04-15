/* ===========================================
   Zen Bali - Main JavaScript
   Developed by net1io.com
   Copyright (C) 2024
   =========================================== */

// API Configuration
const API_BASE = '/api';

// Token Management
const Auth = {
    getToken() {
        return localStorage.getItem('zenbali_token');
    },
    setToken(token) {
        localStorage.setItem('zenbali_token', token);
    },
    removeToken() {
        localStorage.removeItem('zenbali_token');
    },
    getUser() {
        const user = localStorage.getItem('zenbali_user');
        return user ? JSON.parse(user) : null;
    },
    setUser(user) {
        localStorage.setItem('zenbali_user', JSON.stringify(user));
    },
    removeUser() {
        localStorage.removeItem('zenbali_user');
    },
    isLoggedIn() {
        return !!this.getToken();
    },
    logout() {
        this.removeToken();
        this.removeUser();
        window.location.href = '/';
    }
};

// API Client
const API = {
    async request(endpoint, options = {}) {
        const url = `${API_BASE}${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        const token = Auth.getToken();
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        try {
            const response = await fetch(url, {
                ...options,
                headers
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Request failed');
            }

            return data;
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    },

    get(endpoint) {
        return this.request(endpoint, { method: 'GET' });
    },

    post(endpoint, body) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(body)
        });
    },

    put(endpoint, body) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(body)
        });
    },

    delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    },

    async upload(endpoint, formData) {
        const url = `${API_BASE}${endpoint}`;
        const token = Auth.getToken();
        const headers = {};

        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(url, {
            method: 'POST',
            headers,
            body: formData
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Upload failed');
        }

        return data;
    }
};

// Utility Functions
const Utils = {
    formatDate(dateStr) {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            weekday: 'short',
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    },

    formatDateTime(dateStr) {
        const date = new Date(dateStr);
        return date.toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    },

    formatCurrency(amount, currency = 'USD') {
        if (amount === 0) return 'Free';
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency
        }).format(amount);
    },

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    truncate(text, length = 100) {
        if (!text) return '';
        if (text.length <= length) return text;
        return text.substring(0, length) + '...';
    },

    showAlert(message, type = 'success') {
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type}`;
        alertDiv.textContent = message;
        
        const container = document.querySelector('.container') || document.body;
        container.insertBefore(alertDiv, container.firstChild);

        setTimeout(() => {
            alertDiv.remove();
        }, 5000);
    },

    showError(message) {
        this.showAlert(message, 'error');
    },

    showSuccess(message) {
        this.showAlert(message, 'success');
    },

    getQueryParam(name) {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get(name);
    },

    setQueryParams(params) {
        const url = new URL(window.location);
        Object.entries(params).forEach(([key, value]) => {
            if (value) {
                url.searchParams.set(key, value);
            } else {
                url.searchParams.delete(key);
            }
        });
        window.history.pushState({}, '', url);
    }
};

// Form Handling
const Form = {
    serialize(form) {
        const formData = new FormData(form);
        const data = {};
        for (const [key, value] of formData.entries()) {
            if (value) {
                // Handle numeric fields
                if (['location_id', 'event_type_id', 'entrance_type_id'].includes(key)) {
                    data[key] = parseInt(value, 10);
                } else if (key === 'entrance_fee') {
                    data[key] = parseFloat(value) || 0;
                } else {
                    data[key] = value;
                }
            }
        }
        return data;
    },

    validate(form) {
        let isValid = true;
        const errors = [];

        form.querySelectorAll('[required]').forEach(field => {
            if (!field.value.trim()) {
                isValid = false;
                field.classList.add('error');
                errors.push(`${field.name || 'Field'} is required`);
            } else {
                field.classList.remove('error');
            }
        });

        return { isValid, errors };
    },

    reset(form) {
        form.reset();
        form.querySelectorAll('.error').forEach(el => el.classList.remove('error'));
    }
};

// Modal Handling
const Modal = {
    show(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.add('active');
            document.body.style.overflow = 'hidden';
        }
    },

    hide(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.remove('active');
            document.body.style.overflow = '';
        }
    },

    init() {
        document.querySelectorAll('.modal-overlay').forEach(overlay => {
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) {
                    overlay.classList.remove('active');
                    document.body.style.overflow = '';
                }
            });
        });

        document.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', () => {
                const overlay = btn.closest('.modal-overlay');
                if (overlay) {
                    overlay.classList.remove('active');
                    document.body.style.overflow = '';
                }
            });
        });
    }
};

// Visitor Tracking
async function trackVisitor() {
    try {
        await API.post('/visitors', {
            user_agent: navigator.userAgent
        });
    } catch (error) {
        console.log('Visitor tracking failed:', error);
    }
}

async function loadVisitorStats() {
    try {
        const response = await API.get('/visitors/stats');
        if (response.success && response.data) {
            const stats = response.data;
            
            const countEl = document.getElementById('visitorCount');
            if (countEl) {
                countEl.textContent = stats.total_visitors.toLocaleString();
            }

            const lastEl = document.getElementById('lastVisitor');
            if (lastEl && stats.last_visitor_date) {
                const date = new Date(stats.last_visitor_date);
                const location = [stats.last_visitor_city, stats.last_visitor_country]
                    .filter(Boolean)
                    .join(', ');
                lastEl.textContent = `${date.toLocaleString()} from ${location || 'Unknown'}`;
            }
        }
    } catch (error) {
        console.log('Failed to load visitor stats:', error);
    }
}

// Initialize modals when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    Modal.init();
});
