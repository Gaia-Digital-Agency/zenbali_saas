/* Zen Bali - Auth JavaScript */

function requireAuth() {
    if (!Auth.isLoggedIn()) {
        window.location.href = '/creator/login.html';
        return false;
    }
    return true;
}

function requireAdminAuth() {
    const token = localStorage.getItem('zenbali_admin_token');
    if (!token) {
        window.location.href = '/admin/login.html';
        return false;
    }
    return true;
}
