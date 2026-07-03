const API_BASE = '/api/v1';

let currentSection = 'dashboard';
let editingProductId = null;
let productColors = [];

function addColor() {
    const nameInput = document.getElementById('color-name-input');
    const hexInput = document.getElementById('color-hex-input');
    const stockInput = document.getElementById('color-stock-input');
    const name = nameInput.value.trim();
    const hex = hexInput.value.trim();
    const stock = parseInt(stockInput.value) || 0;
    if (!name || !hex) return;
    productColors.push({ name, hex_code: hex, stock });
    nameInput.value = '';
    hexInput.value = '';
    stockInput.value = '0';
    renderColorList();
}

function removeColor(index) {
    productColors.splice(index, 1);
    renderColorList();
}

function renderColorList() {
    const container = document.getElementById('color-list');
    container.innerHTML = productColors.map((c, i) => `
        <div class="color-chip">
            <span class="color-swatch" style="background:${c.hex_code}"></span>
            <span>${c.name}</span>
            <span class="color-stock-badge">Stock: ${c.stock || 0}</span>
            <button type="button" class="color-remove" onclick="removeColor(${i})">&times;</button>
        </div>
    `).join('');
}

async function editProduct(productId) {
    try {
        const response = await fetch(`${API_BASE}/products/${productId}`);
        const product = await response.json();
        editingProductId = productId;
        showProductModal(product);
    } catch (error) {
        console.error('Error loading product:', error);
        alert('Error al cargar el producto');
    }
}

function showProductModal(product = null) {
    const modal = document.getElementById('productModal');
    const form = document.getElementById('productForm');
    const title = document.getElementById('productModalTitle');

    if (product) {
        title.textContent = 'Editar Producto';
        form.name.value = product.name;
        form.description.value = product.description;
        form.price.value = product.price;
        form.stock.value = product.stock;
        form.category.value = product.category;
        form.image_url.value = product.imageUrl || '';
        form.material.value = product.material || '';
        form.sizes.value = (product.sizes || []).join(', ');
        productColors = (product.colors || []).map(c => ({ name: c.name, hex_code: c.hex, stock: c.stock || 0 }));
        renderColorList();
    } else {
        title.textContent = 'Nuevo Producto';
        form.reset();
        productColors = [];
        renderColorList();
    }

    modal.classList.add('active');
}

function closeProductModal() {
    document.getElementById('productModal').classList.remove('active');
    editingProductId = null;
}

document.getElementById('productForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const form = e.target;
    const payload = {
        name: form.name.value,
        description: form.description.value,
        price: parseFloat(form.price.value) || 0,
        stock: parseInt(form.stock.value) || 0,
        category: form.category.value,
        imageUrl: form.image_url.value,
        material: form.material.value,
        sizes: form.sizes.value.split(',').map(s => s.trim()).filter(Boolean),
        colors: productColors.map(c => ({ name: c.name, hex: c.hex_code, stock: c.stock || 0 }))
    };

    try {
        const url = editingProductId
            ? `${API_BASE}/admin/products/${editingProductId}`
            : `${API_BASE}/admin/products`;
        const method = editingProductId ? 'PUT' : 'POST';
        const response = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (response.ok) {
            closeProductModal();
            loadProducts();
        } else {
            const err = await response.json();
            alert(err.error || 'Error al guardar producto');
        }
    } catch (error) {
        console.error('Error saving product:', error);
        alert('Error al guardar producto');
    }
});

async function deleteProduct(productId) {
    if (!confirm('¿Eliminar este producto?')) return;
    try {
        const response = await fetch(`${API_BASE}/admin/products/${productId}`, { method: 'DELETE' });
        if (response.ok) {
            loadProducts();
        } else {
            alert('Error al eliminar producto');
        }
    } catch (error) {
        console.error('Error deleting product:', error);
        alert('Error al eliminar producto');
    }
}

async function loadPaymentMethods() {
    try {
        const response = await fetch(`${API_BASE}/payment-methods`);
        const methods = await response.json();
        renderPaymentMethods(methods || []);
    } catch (error) {
        console.error('Error loading payment methods:', error);
    }
}

function renderPaymentMethods(methods) {
    const container = document.getElementById('paymentMethods');
    if (methods.length === 0) {
        container.innerHTML = '<p class="loading">No hay métodos de pago</p>';
        return;
    }
    container.innerHTML = methods.map(m => `
        <div class="payment-card">
            <h3>${m.name}</h3>
            <p>${m.description || ''}</p>
            <span class="status-badge ${m.enabled ? 'paid' : 'disabled'}">${m.enabled ? 'Activo' : 'Inactivo'}</span>
        </div>
    `).join('');
}

async function loadUsers() {
    try {
        const response = await fetch(`${API_BASE}/admin/users`);
        const users = await response.json();
        renderUsers(users || []);
    } catch (error) {
        console.error('Error loading users:', error);
    }
}

function renderUsers(users) {
    const container = document.getElementById('usersTable');
    if (users.length === 0) {
        container.innerHTML = '<p class="loading">No hay usuarios</p>';
        return;
    }
    container.innerHTML = `
        <table>
            <thead>
                <tr><th>ID</th><th>Usuario</th><th>Email</th><th>Rol</th><th>Registro</th></tr>
            </thead>
            <tbody>
                ${users.map(user => `
                    <tr>
                        <td>${user.id}</td>
                        <td>${user.username}</td>
                        <td>${user.email}</td>
                        <td><span class="status-badge ${user.role}">${user.role}</span></td>
                        <td>${formatDate(user.createdAt)}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
}

async function loadSettings() {
    try {
        const response = await fetch(`${API_BASE}/admin/settings`);
        const settings = await response.json();
        renderSettings(settings || []);
    } catch (error) {
        console.error('Error loading settings:', error);
    }
}

function renderSettings(settings) {
    const container = document.getElementById('settingsForm');
    if (settings.length === 0) {
        container.innerHTML = '<p class="loading">No hay configuraciones</p>';
        return;
    }
    const html = settings.map(setting => `
        <div class="form-group">
            <label>${setting.key}</label>
            <input type="text" value="${setting.value}" onchange="updateSetting('${setting.key}', this.value)">
        </div>
    `).join('');
    container.innerHTML = html;
}

async function updateSetting(key, value) {
    try {
        const response = await fetch(`${API_BASE}/admin/settings`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ key, value })
        });
        if (response.ok) {
            alert('Configuración actualizada');
        } else {
            alert('Error al actualizar la configuración');
        }
    } catch (error) {
        console.error('Error updating setting:', error);
        alert('Error al actualizar la configuración');
    }
}

async function loadLogs() {
    try {
        const response = await fetch(`${API_BASE}/admin/logs`);
        const logs = await response.json();
        renderLogs(logs || []);
    } catch (error) {
        console.error('Error loading logs:', error);
    }
}

function renderLogs(logs) {
    const container = document.getElementById('logsTable');
    if (logs.length === 0) {
        container.innerHTML = '<p class="loading">No hay logs</p>';
        return;
    }
    const html = `
        <table>
            <thead>
                <tr><th>Fecha</th><th>Usuario</th><th>Acción</th><th>Entidad</th><th>IP</th></tr>
            </thead>
            <tbody>
                ${logs.map(log => `
                    <tr>
                        <td>${formatDate(log.createdAt)}</td>
                        <td>${log.userId || 'N/A'}</td>
                        <td>${log.action}</td>
                        <td>${log.entity || 'N/A'}</td>
                        <td>${log.ipAddress || 'N/A'}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    container.innerHTML = html;
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('es-CO', {
        style: 'currency',
        currency: 'COP',
        minimumFractionDigits: 0
    }).format(amount);
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('es-CO', {
        year: 'numeric', month: 'short', day: 'numeric',
        hour: '2-digit', minute: '2-digit'
    }).format(date);
}

async function loadDashboard() {
    try {
        const resp = await fetch(`${API_BASE}/admin/dashboard/stats`);
        const stats = await resp.json();
        renderDashboardStats(stats);
    } catch (err) {
        console.error('Error loading dashboard:', err);
    }
}

function renderDashboardStats(stats) {
    document.getElementById('totalRevenue').textContent = formatCurrency(stats.totalRevenue);
    document.getElementById('totalOrders').textContent = stats.totalOrders;
    document.getElementById('pendingOrders').textContent = stats.pendingOrders;
    document.getElementById('totalCustomers').textContent = stats.totalCustomers;

    const orders = stats.recentOrders || [];
    document.getElementById('recentOrders').innerHTML = orders.length === 0
        ? '<p class="loading">No hay órdenes recientes</p>'
        : `<table><thead><tr><th>ID</th><th>Total</th><th>Estado</th><th>Pago</th><th>Fecha</th></tr></thead><tbody>
            ${orders.map(o => `
                <tr>
                    <td>#${o.id}</td>
                    <td>${formatCurrency(o.total)}</td>
                    <td><span class="status-badge ${o.status}">${o.status}</span></td>
                    <td><span class="status-badge ${o.paymentStatus}">${o.paymentStatus || 'pending'}</span></td>
                    <td>${formatDate(o.createdAt)}</td>
                </tr>
            `).join('')}
        </tbody></table>`;

    const products = stats.topProducts || [];
    document.getElementById('topProducts').innerHTML = products.length === 0
        ? '<p class="loading">Sin ventas aún</p>'
        : products.map(p => {
            const img = p.imageUrl ? `<img src="${imgUrl(p.imageUrl)}" alt="${p.name}" class="top-product-img">` : '<div class="top-product-img placeholder"></div>';
            return `
            <div class="top-product-item">
                ${img}
                <div class="top-product-info">
                    <span class="top-product-name">${p.name}</span>
                    <span class="top-product-price">${formatCurrency(p.price)}</span>
                </div>
            </div>`;
        }).join('');
}

async function loadOrders() {
    const statusFilter = document.getElementById('orderStatusFilter')?.value || '';
    const paymentFilter = document.getElementById('paymentStatusFilter')?.value || '';
    let url = `${API_BASE}/admin/orders`;
    const params = new URLSearchParams();
    if (statusFilter) params.set('status', statusFilter);
    if (paymentFilter) params.set('payment_status', paymentFilter);
    const qs = params.toString();
    if (qs) url += '?' + qs;

    try {
        const resp = await fetch(url);
        const orders = await resp.json();
        renderOrders(orders || []);
    } catch (err) {
        console.error('Error loading orders:', err);
    }
}

function renderOrders(orders) {
    const container = document.getElementById('ordersTable');
    if (orders.length === 0) {
        container.innerHTML = '<p class="loading">No hay órdenes</p>';
        return;
    }
    container.innerHTML = `
        <table>
            <thead>
                <tr>
                    <th>ID</th><th>Cliente</th><th>Total</th><th>Estado</th>
                    <th>Pago</th><th>Método</th><th>Fecha</th><th>Acciones</th>
                </tr>
            </thead>
            <tbody>
                ${orders.map(o => `
                    <tr>
                        <td>#${o.id}</td>
                        <td>Usuario #${o.userId}</td>
                        <td>${formatCurrency(o.total)}</td>
                        <td><span class="status-badge ${o.status}">${o.status}</span></td>
                        <td><span class="status-badge ${o.paymentStatus}">${o.paymentStatus}</span></td>
                        <td>${o.paymentMethod}</td>
                        <td>${formatDate(o.createdAt)}</td>
                        <td><button class="btn-success btn-sm" onclick="viewOrder(${o.id})">Ver</button></td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
}

async function viewOrder(orderId) {
    try {
        const resp = await fetch(`${API_BASE}/orders/${orderId}`);
        const order = await resp.json();
        renderOrderDetails(order);
    } catch (err) {
        console.error('Error loading order details:', err);
        alert('Error al cargar detalles de la orden');
    }
}

function renderOrderDetails(order) {
    const container = document.getElementById('orderDetails');
    container.innerHTML = `
        <div class="order-details">
            <h3>Información de Envío</h3>
            <p><strong>Dirección:</strong> ${order.shippingAddress}</p>
            <p><strong>Ciudad:</strong> ${order.shippingCity}</p>
            <p><strong>Teléfono:</strong> ${order.shippingPhone}</p>
            <h3>Productos</h3>
            ${(order.items || []).map(item => `
                <div class="order-item">
                    <p><strong>${item.productName}</strong> × ${item.quantity}</p>
                    <p>Precio: ${formatCurrency(item.productPrice)}</p>
                </div>
            `).join('')}
            <h3>Total: ${formatCurrency(order.total)}</h3>
            <h3>Actualizar Estado</h3>
            <div class="form-group">
                <label>Estado de la Orden</label>
                <select id="orderStatusSelect">
                    <option value="pending" ${order.status === 'pending' ? 'selected' : ''}>Pendiente</option>
                    <option value="processing" ${order.status === 'processing' ? 'selected' : ''}>Procesando</option>
                    <option value="completed" ${order.status === 'completed' ? 'selected' : ''}>Completado</option>
                    <option value="cancelled" ${order.status === 'cancelled' ? 'selected' : ''}>Cancelado</option>
                </select>
            </div>
            <div class="form-group">
                <label>Estado del Pago</label>
                <select id="paymentStatusSelect">
                    <option value="pending" ${order.paymentStatus === 'pending' ? 'selected' : ''}>Pendiente</option>
                    <option value="paid" ${order.paymentStatus === 'paid' ? 'selected' : ''}>Pagado</option>
                    <option value="failed" ${order.paymentStatus === 'failed' ? 'selected' : ''}>Fallido</option>
                </select>
            </div>
            <button class="btn-primary" onclick="updateOrderStatus(${order.id})">Actualizar</button>
        </div>
    `;
    document.getElementById('orderModal').classList.add('active');
}

async function updateOrderStatus(orderId) {
    const status = document.getElementById('orderStatusSelect').value;
    const paymentStatus = document.getElementById('paymentStatusSelect').value;
    try {
        const resp = await fetch(`${API_BASE}/admin/orders/${orderId}/status`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status, paymentStatus })
        });
        if (resp.ok) {
            alert('Orden actualizada');
            document.getElementById('orderModal').classList.remove('active');
            loadOrders();
            loadDashboard();
        }
    } catch (err) {
        console.error('Error updating order:', err);
        alert('Error al actualizar la orden');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const modal = document.getElementById('orderModal');
    if (modal) {
        modal.querySelector('.close').onclick = () => modal.classList.remove('active');
    }
});

function imgUrl(url) {
    if (!url || url.startsWith('http') || url.startsWith('/')) return url;
    return '/assets/img/' + url;
}

async function loadProducts() {
    try {
        const resp = await fetch(`${API_BASE}/products`);
        const data = await resp.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        renderProducts(list);
    } catch (err) {
        console.error('Error loading products:', err);
    }
}

function renderProducts(products) {
    const container = document.getElementById('productsGrid');
    if (products.length === 0) {
        container.innerHTML = '<p class="loading">No hay productos</p>';
        return;
    }
    container.innerHTML = products.map(p => `
        <div class="product-card">
            <img src="${imgUrl(p.imageUrl)}" alt="${p.name}" class="product-image" loading="lazy">
            <div class="product-info">
                <div class="product-name">${p.name}</div>
                <div class="product-price">${formatCurrency(p.price)}</div>
                <div class="product-stock">Stock: ${p.stock} · ${p.category}</div>
                <div class="product-actions">
                    <button class="btn-success btn-sm" onclick="editProduct(${p.id})">Editar</button>
                    <button class="btn-danger btn-sm" onclick="deleteProduct(${p.id})">Eliminar</button>
                </div>
            </div>
        </div>
    `).join('');
}

document.getElementById('addProductBtn')?.addEventListener('click', () => showProductModal());

const loginOverlay = document.getElementById('loginOverlay');
const loginForm = document.getElementById('adminLoginForm');
const loginError = document.getElementById('loginError');

function showLoginError(msg) {
    loginError.textContent = msg;
    loginError.style.display = 'block';
}

function hideLoginError() {
    loginError.style.display = 'none';
}

function showAdminContent() {
    loginOverlay.classList.remove('active');
    document.querySelector('.admin-container').style.display = 'flex';
}

function showLoginOverlay() {
    loginOverlay.classList.add('active');
    document.querySelector('.admin-container').style.display = 'none';
}

async function checkAdminAuth() {
    try {
        const resp = await fetch(`${API_BASE}/admin/users`);
        if (resp.ok) {
            showAdminContent();
            loadDashboard();
            return;
        }
    } catch (_) {}
    showLoginOverlay();
}

loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    hideLoginError();
    const email = document.getElementById('loginEmail').value.trim();
    const password = document.getElementById('loginPassword').value;
    const btn = loginForm.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Ingresando...';
    try {
        const resp = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        if (resp.ok) {
            const user = await resp.json();
            document.getElementById('loginEmail').value = '';
            document.getElementById('loginPassword').value = '';
            showAdminContent();
            document.getElementById('adminUsername').textContent = user.username;
            loadDashboard();
        } else {
            const err = await resp.json();
            showLoginError(err.error || 'Credenciales inválidas');
        }
    } catch (err) {
        showLoginError('Error de conexión con el servidor');
    } finally {
        btn.disabled = false;
        btn.textContent = 'Ingresar';
    }
});

document.getElementById('logoutBtn').addEventListener('click', async () => {
    try {
        await fetch(`${API_BASE}/auth/logout`, { method: 'POST' });
    } catch (_) {}
    document.getElementById('adminUsername').textContent = 'Admin';
    showLoginOverlay();
});

document.addEventListener('DOMContentLoaded', () => {
    document.querySelector('.admin-container').style.display = 'none';
    checkAdminAuth();
});

document.querySelectorAll('.nav-item').forEach(item => {
    item.addEventListener('click', (e) => {
        e.preventDefault();
        const section = item.dataset.section;
        if (!section) return;

        document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
        item.classList.add('active');

        document.querySelectorAll('.content-section').forEach(s => s.classList.remove('active'));
        document.getElementById(`section-${section}`)?.classList.add('active');

        document.getElementById('pageTitle').textContent = item.textContent.trim();

        switch (section) {
            case 'dashboard': loadDashboard(); break;
            case 'orders': loadOrders(); break;
            case 'products': loadProducts(); break;
            case 'payments': loadPaymentMethods(); break;
            case 'users': loadUsers(); break;
            case 'settings': loadSettings(); break;
            case 'logs': loadLogs(); break;
        }
    });
});

document.addEventListener('DOMContentLoaded', () => {
    const pm = document.getElementById('productModal');
    if (pm) {
        const existingClose = pm.querySelector('.close');
        if (existingClose) {
            existingClose.onclick = () => closeProductModal();
        }
    }
});
