// IPv6 Proxy Hub - Enterprise Frontend Engine
let globalProxies = [];
let serverStats = {};
let currentFilter = 'all';
let chartHistory = Array(20).fill(0);
let prevTrafficBytes = 0;

// Toast Notifications
function showToast(msg, type = 'success') {
  const container = document.getElementById('toast-container');
  const div = document.createElement('div');
  div.className = `toast-item ${type}`;
  div.innerHTML = `${type === 'success' ? '✅' : '⚠️'} <span>${msg}</span>`;
  container.appendChild(div);
  setTimeout(() => div.remove(), 3500);
}

// Modal Control
function openModal(id) {
  document.getElementById(id).classList.add('open');
}

function closeModal(id) {
  document.getElementById(id).classList.remove('open');
}

// Tab / View Switching via Sidebar
const pageTitles = {
  stats: {
    title: 'Thống Kê & Giám Sát',
    sub: 'Tổng quan lưu lượng, hiệu năng VPS và cảnh báo an toàn mạng'
  },
  proxies: {
    title: 'Quản Lý Cổng Proxy',
    sub: 'Quản lý, tạo hàng loạt, cấu hình cổng HTTP & SOCKS5'
  },
  integration: {
    title: 'Tích Hợp & Hướng Dẫn',
    sub: 'Cấu hình Antidetect Browser, cURL và code mẫu đa luồng'
  },
  settings: {
    title: 'Cài Đặt Hệ Thống',
    sub: 'Đổi tài khoản quản trị và giám sát dịch vụ hạ tầng'
  }
};

function switchView(viewName) {
  // Update sidebar active
  document.querySelectorAll('.sidebar-menu .nav-item').forEach(el => {
    el.classList.remove('active');
  });
  const navBtn = document.getElementById(`nav-${viewName}`);
  if (navBtn) navBtn.classList.add('active');

  // Update main content tabs
  document.querySelectorAll('.tab-content').forEach(el => {
    el.classList.remove('active');
  });
  const tabEl = document.getElementById(`tab-${viewName}`);
  if (tabEl) tabEl.classList.add('active');

  // Update Breadcrumb Header
  const info = pageTitles[viewName] || pageTitles.stats;
  document.getElementById('header-breadcrumb-title').textContent = info.title;
  document.getElementById('header-breadcrumb-sub').textContent = info.sub;
}

// Format Bytes
function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Random Generator
function genRandPass(targetId) {
  const chars = 'abcdefghjkmnpqrstuvwxyz23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
  let res = '';
  for (let i = 0; i < 8; i++) res += chars[Math.floor(Math.random() * chars.length)];
  document.getElementById(targetId).value = res;
}

function toggleStickyInput(val, targetId) {
  const el = document.getElementById(targetId);
  if (el) el.style.display = (val === 'sticky') ? 'flex' : 'none';
}

// Auth Handlers
async function doLogin(e) {
  e.preventDefault();
  const u = document.getElementById('login-user').value.trim();
  const p = document.getElementById('login-pass').value.trim();
  try {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: u, password: p })
    });
    const data = await res.json();
    if (res.ok) {
      document.getElementById('login-modal').style.display = 'none';
      showToast('Đăng nhập thành công!');
      loadAllData();
    } else {
      showToast(data.error || 'Đăng nhập thất bại', 'error');
    }
  } catch (err) {
    showToast('Lỗi kết nối máy chủ', 'error');
  }
}

async function doLogout() {
  if (!confirm('Bạn có chắc chắn muốn đăng xuất khỏi trang quản trị?')) return;
  await fetch('/api/logout', { method: 'POST' });
  document.getElementById('login-modal').style.display = 'flex';
  showToast('Đã đăng xuất');
}

// Fetch Data
async function loadAllData() {
  try {
    const [statsRes, proxiesRes] = await Promise.all([
      fetch('/api/stats'),
      fetch('/api/proxies')
    ]);

    if (statsRes.status === 401 || proxiesRes.status === 401) {
      document.getElementById('login-modal').style.display = 'flex';
      return;
    }

    serverStats = await statsRes.json();
    globalProxies = await proxiesRes.json();

    renderSidebarAndHeader();
    renderStats();
    renderProxiesTable();
  } catch (err) {
    console.error('Lỗi nạp dữ liệu:', err);
  }
}

function renderSidebarAndHeader() {
  const ip = serverStats.public_ip || serverStats.public_ipv4 || '103.199.11.9';
  document.getElementById('disp-ipv4').textContent = ip;
  document.getElementById('disp-subnet').textContent = serverStats.prefix || 'N/A';
  document.getElementById('guide-host').textContent = ip;
  document.getElementById('disp-uptime').textContent = serverStats.uptime || '0h 0m';

  document.getElementById('disp-total-badge').textContent = `${serverStats.total_proxies || 0} Cổng Cấu Hình`;
  document.getElementById('disp-active-badge').textContent = `${serverStats.active_proxies || 0} Đang Chạy`;

  // Sidebar count badge
  const sbBadge = document.getElementById('sidebar-proxy-count');
  if (sbBadge) sbBadge.textContent = globalProxies.length;
}

function renderStats() {
  // KPI Cards
  const total = serverStats.total_proxies || 0;
  const active = serverStats.active_proxies || 0;
  document.getElementById('kpi-active-ports').textContent = `${active} / ${total}`;
  const pct = total > 0 ? Math.round((active / total) * 100) : 0;
  document.getElementById('kpi-active-desc').textContent = `${pct}% tổng số cổng đang mở`;

  document.getElementById('kpi-total-traffic').textContent = formatBytes(serverStats.total_bytes);
  document.getElementById('kpi-ram').textContent = serverStats.ram_usage || '1.5 MB';
  document.getElementById('kpi-goroutines').textContent = `Goroutines: ${serverStats.goroutines || 0} | Cores: ${serverStats.cpu_cores || 1}`;

  const rotCounts = serverStats.rotation_counts || serverStats.rotations || { sticky: 0, request: 0, static: 0 };
  document.getElementById('kpi-rot-summary').textContent = `${rotCounts.sticky || 0} Sticky / ${rotCounts.request || 0} Req`;

  // Rotation Progress Bars
  const sumRot = (rotCounts.sticky || 0) + (rotCounts.request || 0) + (rotCounts.static || 0) || 1;
  document.getElementById('rot-sticky-count').textContent = rotCounts.sticky || 0;
  document.getElementById('rot-request-count').textContent = rotCounts.request || 0;
  document.getElementById('rot-static-count').textContent = rotCounts.static || 0;

  document.getElementById('rot-sticky-bar').style.width = Math.round(((rotCounts.sticky || 0) / sumRot) * 100) + '%';
  document.getElementById('rot-request-bar').style.width = Math.round(((rotCounts.request || 0) / sumRot) * 100) + '%';
  document.getElementById('rot-static-bar').style.width = Math.round(((rotCounts.static || 0) / sumRot) * 100) + '%';

  // Live Chart
  updateTrafficChart(serverStats.total_bytes);

  // Top Proxies
  renderTopProxies(serverStats.top_proxies || []);
}

function updateTrafficChart(currentBytes) {
  if (prevTrafficBytes === 0) prevTrafficBytes = currentBytes;
  const diffBytes = Math.max(0, currentBytes - prevTrafficBytes);
  prevTrafficBytes = currentBytes;

  const rateKB = diffBytes / 1024 / 5;
  chartHistory.push(rateKB);
  chartHistory.shift();

  let maxVal = Math.max(...chartHistory);
  if (maxVal < 10) maxVal = 10;
  document.getElementById('chart-peak-val').textContent = `Đỉnh: ${maxVal.toFixed(1)} KB/s`;

  const width = 500;
  const height = 180;
  const step = width / (chartHistory.length - 1);

  const points = chartHistory.map((val, idx) => {
    const x = idx * step;
    const y = height - (val / maxVal) * (height - 20) - 10;
    return { x, y };
  });

  const lineD = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ');
  const areaD = `${lineD} L${width},${height} L0,${height} Z`;

  document.getElementById('chart-line').setAttribute('d', lineD);
  document.getElementById('chart-area').setAttribute('d', areaD);
}

function renderTopProxies(topList) {
  const tbody = document.getElementById('top-proxies-body');
  if (!topList || topList.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center; color:var(--text-muted); padding:20px;">Chưa có lưu lượng phát sinh</td></tr>';
    return;
  }

  tbody.innerHTML = topList.map((p, idx) => `
    <tr>
      <td><span class="badge-pill ${idx === 0 ? 'warning' : 'primary'}">#${idx + 1}</span></td>
      <td><span class="port-tag">:${p.port}</span></td>
      <td><span class="auth-info"><strong>${p.username || p.name || 'No Auth'}</strong></span></td>
      <td><span class="rot-tag ${p.rotation_type === 'sticky' ? 'rot-sticky' : 'rot-request'}">${p.rotation_type}</span></td>
      <td style="font-weight:700; color:#fff;">${formatBytes(p.bytes_used)}</td>
      <td><span class="badge-pill ${p.enabled ? 'success' : 'primary'}">${p.enabled ? 'Đang Chạy' : 'Tạm Dừng'}</span></td>
    </tr>
  `).join('');
}

// Proxies Table & Filter
function setFilter(type, btn) {
  currentFilter = type;
  document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  filterProxies();
}

function filterProxies() {
  const q = (document.getElementById('proxy-search').value || '').toLowerCase();
  const list = globalProxies.filter(p => {
    const matchQ = (p.port + '').includes(q) || (p.username || '').toLowerCase().includes(q);
    if (!matchQ) return false;
    if (currentFilter === 'active') return p.enabled;
    if (currentFilter === 'sticky') return p.rotation_type === 'sticky';
    if (currentFilter === 'request') return p.rotation_type === 'request';
    return true;
  });
  renderProxyRows(list);
}

function renderProxiesTable() {
  document.getElementById('count-all').textContent = globalProxies.length;
  document.getElementById('count-active').textContent = globalProxies.filter(p => p.enabled).length;
  filterProxies();
}

function renderProxyRows(list) {
  const tbody = document.getElementById('proxy-table-body');
  if (list.length === 0) {
    tbody.innerHTML = '<tr><td colspan="9" style="text-align:center; color:var(--text-muted); padding:40px;">Không tìm thấy cổng proxy nào phù hợp</td></tr>';
    return;
  }

  tbody.innerHTML = list.map((p, idx) => {
    let rotBadge = '';
    if (p.rotation_type === 'sticky') {
      rotBadge = `<span class="rot-tag rot-sticky">⏱️ Sticky ${p.sticky_sec || 10}s</span>`;
    } else if (p.rotation_type === 'static') {
      rotBadge = `<span class="rot-tag rot-static">📌 Static</span>`;
    } else {
      rotBadge = `<span class="rot-tag rot-request">⚡ Mỗi Req 1 IP</span>`;
    }

    const maxGBText = p.max_bytes > 0 ? ` / ${formatBytes(p.max_bytes)}` : '';
    const expireText = p.expires_at ? new Date(p.expires_at).toLocaleDateString('vi-VN') : 'Vĩnh viễn';
    const authText = p.username ? `<strong>${p.username}</strong> : ${p.password}` : '<span style="color:var(--text-muted);">Không mật khẩu</span>';

    let protoBadge = '';
    if (p.proto === 'http') {
      protoBadge = `<span class="badge-pill primary">HTTP(S)</span>`;
    } else if (p.proto === 'socks5') {
      protoBadge = `<span class="badge-pill warning">SOCKS5</span>`;
    } else {
      protoBadge = `<span class="badge-pill success">HTTP & SOCKS5</span>`;
    }

    return `
      <tr>
        <td style="color:var(--text-muted); font-size:0.75rem;">${idx + 1}</td>
        <td><span class="port-tag">:${p.port}</span></td>
        <td>${protoBadge}</td>
        <td><div class="auth-info">${authText}</div></td>
        <td>${rotBadge}</td>
        <td><strong style="color:#fff;">${formatBytes(p.bytes_used)}</strong><span style="color:var(--text-muted); font-size:0.75rem;">${maxGBText}</span></td>
        <td style="font-size:0.8rem; color:var(--text-secondary);">${expireText}</td>
        <td style="text-align:center;">
          <label class="switch">
            <input type="checkbox" ${p.enabled ? 'checked' : ''} onchange="toggleProxy('${p.id}', this.checked)">
            <span class="slider"></span>
          </label>
        </td>
        <td style="text-align:right;">
          <div style="display:inline-flex; gap:6px;">
            <button class="btn btn-secondary btn-sm" title="Copy Host:Port:User:Pass" onclick="copyProxyString(${p.port}, '${p.username||''}', '${p.password||''}')">📋</button>
            <button class="btn btn-secondary btn-sm" title="Xem Code Mẫu" onclick="openSnippetModal('${p.id}')">💻</button>
            <button class="btn btn-secondary btn-sm" title="Chỉnh sửa" onclick="openEditModal('${p.id}')">✏️</button>
            <button class="btn btn-danger btn-sm" title="Xóa" onclick="deleteProxy('${p.id}')">🗑️</button>
          </div>
        </td>
      </tr>
    `;
  }).join('');
}

// Proxy CRUD Operations
function copyProxyString(port, u, p) {
  const host = serverStats.public_ip || serverStats.public_ipv4 || '103.199.11.9';
  let str = `${host}:${port}`;
  if (u && p) str += `:${u}:${p}`;
  navigator.clipboard.writeText(str);
  showToast(`Đã copy: ${str}`);
}

async function toggleProxy(id, enabled) {
  try {
    const res = await fetch(`/api/proxies/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: enabled })
    });
    if (res.ok) {
      showToast(enabled ? 'Đã bật cổng proxy' : 'Đã tạm dừng cổng');
      loadAllData();
    } else {
      showToast('Lỗi khi cập nhật trạng thái', 'error');
    }
  } catch (e) { showToast('Lỗi kết nối', 'error'); }
}

async function deleteProxy(id) {
  if (!confirm('Bạn có chắc chắn muốn xóa cổng proxy này?')) return;
  try {
    const res = await fetch(`/api/proxies/${id}`, { method: 'DELETE' });
    if (res.ok) {
      showToast('Đã xóa proxy');
      loadAllData();
    } else {
      showToast('Không thể xóa proxy', 'error');
    }
  } catch (e) { showToast('Lỗi kết nối', 'error'); }
}

function openCreateModal() {
  document.getElementById('modal-proxy-title').textContent = 'Thêm Cổng Proxy Mới';
  document.getElementById('p-id').value = '';
  document.getElementById('p-port').value = '';
  document.getElementById('p-port').disabled = false;
  document.getElementById('p-user').value = '';
  document.getElementById('p-pass').value = '';
  document.getElementById('p-proto').value = 'both';
  document.getElementById('p-days').value = '30';
  document.getElementById('p-rot-type').value = 'sticky';
  document.getElementById('p-sticky-sec').value = '10';
  document.getElementById('p-maxgb').value = '0';
  toggleStickyInput('sticky', 'p-sticky-group');
  openModal('modal-proxy');
}

function openEditModal(id) {
  const p = globalProxies.find(x => x.id === id);
  if (!p) return;
  document.getElementById('modal-proxy-title').textContent = `Chỉnh Sửa Cổng Proxy :${p.port}`;
  document.getElementById('p-id').value = p.id;
  document.getElementById('p-port').value = p.port;
  document.getElementById('p-port').disabled = false;
  document.getElementById('p-user').value = p.username || '';
  document.getElementById('p-pass').value = p.password || '';
  document.getElementById('p-proto').value = p.proto || 'both';
  document.getElementById('p-days').value = '0';
  document.getElementById('p-rot-type').value = p.rotation_type || 'sticky';
  document.getElementById('p-sticky-sec').value = p.sticky_sec || 10;
  document.getElementById('p-maxgb').value = p.max_bytes ? (p.max_bytes / 1024 / 1024 / 1024).toFixed(1) : '0';
  toggleStickyInput(p.rotation_type || 'sticky', 'p-sticky-group');
  openModal('modal-proxy');
}

async function saveProxy(e) {
  e.preventDefault();
  const id = document.getElementById('p-id').value;
  const port = parseInt(document.getElementById('p-port').value);
  const user = document.getElementById('p-user').value.trim();
  const pass = document.getElementById('p-pass').value.trim();
  const proto = document.getElementById('p-proto').value;
  const rotType = document.getElementById('p-rot-type').value;
  const stickySec = parseInt(document.getElementById('p-sticky-sec').value) || 10;
  const maxGB = parseFloat(document.getElementById('p-maxgb').value) || 0;
  const days = parseInt(document.getElementById('p-days').value) || 0;

  const payload = {
    port: port,
    username: user,
    password: pass,
    proto: proto,
    rotation_type: rotType,
    sticky_sec: stickySec,
    max_gb: maxGB,
    expire_days: days
  };

  try {
    const url = id ? `/api/proxies/${id}` : '/api/proxies';
    const method = id ? 'PUT' : 'POST';
    const res = await fetch(url, {
      method: method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await res.json();
    if (res.ok) {
      showToast(id ? 'Cập nhật thành công!' : 'Tạo proxy thành công!');
      closeModal('modal-proxy');
      loadAllData();
    } else {
      showToast(data.error || 'Có lỗi xảy ra', 'error');
    }
  } catch (err) {
    showToast('Lỗi kết nối máy chủ', 'error');
  }
}

// Bulk Generation
function openBulkModal() {
  document.getElementById('b-proto').value = 'both';
  openModal('modal-bulk');
}

async function submitBulkProxies(e) {
  e.preventDefault();
  const startPort = parseInt(document.getElementById('b-start-port').value);
  const count = parseInt(document.getElementById('b-count').value);
  const prefix = document.getElementById('b-prefix').value.trim();
  const pass = document.getElementById('b-pass').value.trim();
  const proto = document.getElementById('b-proto').value;
  const rotType = document.getElementById('b-rot-type').value;
  const stickySec = parseInt(document.getElementById('b-sticky-sec').value) || 10;
  const days = parseInt(document.getElementById('b-days').value) || 30;
  const maxgb = parseFloat(document.getElementById('b-maxgb').value) || 0;

  const payload = {
    start_port: startPort,
    count: count,
    username_prefix: prefix,
    password: pass,
    proto: proto,
    rotation_type: rotType,
    sticky_sec: stickySec,
    expire_days: days,
    max_gb: maxgb
  };

  try {
    const res = await fetch('/api/proxies/bulk', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await res.json();
    if (res.ok) {
      showToast(`Đã tạo thành công ${data.created_count || count} proxy!`);
      closeModal('modal-bulk');
      loadAllData();
      switchView('proxies');
    } else {
      showToast(data.error || 'Lỗi khi tạo hàng loạt', 'error');
    }
  } catch (err) { showToast('Lỗi kết nối', 'error'); }
}

// Export Modal
function openExportModal() {
  openModal('modal-export');
  fetchExport('ip_port_user_pass');
}

async function fetchExport(fmt) {
  const box = document.getElementById('export-box');
  box.textContent = 'Đang trích xuất dữ liệu...';
  try {
    const res = await fetch(`/api/export?format=${fmt}`);
    const text = await res.text();
    box.textContent = text || 'Chưa có proxy nào đang bật.';
  } catch (e) { box.textContent = 'Lỗi trích xuất'; }
}

function copyBoxText(id) {
  const txt = document.getElementById(id).textContent;
  navigator.clipboard.writeText(txt);
  showToast('Đã copy danh sách vào Clipboard!');
}

// Snippet Modal
function openSnippetModal(id) {
  const p = globalProxies.find(x => x.id === id);
  if (!p) return;
  const host = serverStats.public_ip || serverStats.public_ipv4 || '103.199.11.9';
  document.getElementById('snippet-title').textContent = `Code Mẫu Proxy Cổng :${p.port}`;
  
  const auth = (p.username && p.password) ? `${p.username}:${p.password}@` : '';
  const curlCmd = `curl -x http://${auth}${host}:${p.port} https://api64.ipify.org?format=json`;
  
  const pyCode = `import requests

# Với Phương thức 1 (Session Sticky): Thay đổi 'session_id' cho mỗi luồng riêng biệt
# để 50 luồng đồng thời dùng chung 1 cổng mà nhận 50 IP IPv6 khác nhau:
session_id = "thread_1"
username = "${p.username || 'user'}-session-" + session_id

proxies = {
    "http": f"http://{username}:${p.password || 'pass'}@{host}:${p.port}",
    "https": f"http://{username}:${p.password || 'pass'}@{host}:${p.port}",
}

r = requests.get("https://api64.ipify.org?format=json", proxies=proxies, timeout=10)
print("IPv6 của luồng:", r.json()["ip"])`;

  document.getElementById('snip-curl').textContent = curlCmd;
  document.getElementById('snip-python').textContent = pyCode;
  openModal('modal-snippet');
}

// Settings
async function changeAdminSettings(e) {
  e.preventDefault();
  const currPass = document.getElementById('curr-pass').value;
  const newUser = document.getElementById('new-user').value.trim();
  const newPass = document.getElementById('new-pass').value;

  try {
    const res = await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ current_password: currPass, new_username: newUser, new_password: newPass })
    });
    const data = await res.json();
    if (res.ok) {
      showToast('Đã đổi mật khẩu thành công!');
      document.getElementById('curr-pass').value = '';
      document.getElementById('new-pass').value = '';
    } else {
      showToast(data.error || 'Lỗi đổi thông tin', 'error');
    }
  } catch (e) { showToast('Lỗi kết nối', 'error'); }
}

// Auto Refresh
setInterval(() => {
  loadAllData();
}, 5000);

window.addEventListener('DOMContentLoaded', () => {
  loadAllData();
});
