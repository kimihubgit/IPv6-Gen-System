package main

import (
	"net/http"
)

func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="vi">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>IPv6 Proxy Hub - Cloud Enterprise Manager</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-dark: #080c14;
      --bg-surface: rgba(15, 23, 42, 0.75);
      --bg-surface-elevated: rgba(26, 38, 64, 0.85);
      --border-subtle: rgba(255, 255, 255, 0.08);
      --border-glow: rgba(99, 102, 241, 0.4);
      --primary: #6366f1;
      --primary-hover: #4f46e5;
      --primary-glow: rgba(99, 102, 241, 0.35);
      --accent-cyan: #06b6d4;
      --accent-emerald: #10b981;
      --accent-amber: #f59e0b;
      --accent-rose: #f43f5e;
      --text-main: #f8fafc;
      --text-muted: #94a3b8;
      --text-dim: #64748b;
      --radius-sm: 8px;
      --radius-md: 12px;
      --radius-lg: 16px;
      --radius-xl: 20px;
      --transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    }

    * { box-sizing: border-box; margin: 0; padding: 0; font-family: 'Plus Jakarta Sans', -apple-system, sans-serif; }
    body { background-color: var(--bg-dark); color: var(--text-main); min-height: 100vh; overflow-x: hidden; position: relative; }
    
    /* Background Ambient Gradients */
    .ambient-bg {
      position: fixed; inset: 0; pointer-events: none; z-index: 0;
      background: radial-gradient(circle at 15% 10%, rgba(99, 102, 241, 0.12) 0%, transparent 45%),
                  radial-gradient(circle at 85% 80%, rgba(6, 182, 212, 0.08) 0%, transparent 50%),
                  radial-gradient(circle at 50% 50%, rgba(16, 185, 129, 0.04) 0%, transparent 60%);
    }

    /* Layout Wrapper */
    .app-layout { position: relative; z-index: 1; display: flex; flex-direction: column; min-height: 100vh; }

    /* Top Sticky Header */
    header.site-header {
      position: sticky; top: 0; z-index: 100;
      backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px);
      background: rgba(8, 12, 20, 0.85);
      border-bottom: 1px solid var(--border-subtle);
      box-shadow: 0 4px 24px rgba(0,0,0,0.4);
    }

    .header-top {
      max-width: 1440px; margin: 0 auto; padding: 14px 24px;
      display: flex; align-items: center; justify-content: space-between; gap: 16px;
    }

    .brand-section { display: flex; align-items: center; gap: 14px; text-decoration: none; }
    .brand-logo {
      width: 42px; height: 42px; border-radius: var(--radius-md);
      background: linear-gradient(135deg, var(--primary), var(--accent-cyan));
      display: flex; align-items: center; justify-content: center;
      box-shadow: 0 0 20px var(--primary-glow); color: #fff; font-size: 20px; font-weight: 800;
    }
    .brand-title { font-size: 1.15rem; font-weight: 800; letter-spacing: -0.02em; color: #fff; line-height: 1.2; }
    .brand-subtitle { font-size: 0.72rem; color: var(--text-muted); font-weight: 500; display: flex; align-items: center; gap: 6px; }
    
    .badge-pill {
      display: inline-flex; align-items: center; gap: 5px;
      padding: 3px 8px; border-radius: 999px; font-size: 0.7rem; font-weight: 600;
      background: rgba(255,255,255,0.06); border: 1px solid var(--border-subtle);
    }
    .badge-pill.success { background: rgba(16, 185, 129, 0.15); color: #34d399; border-color: rgba(16, 185, 129, 0.3); }
    .badge-pill.primary { background: rgba(99, 102, 241, 0.15); color: #a5b4fc; border-color: rgba(99, 102, 241, 0.3); }
    .badge-pill.warning { background: rgba(245, 158, 11, 0.15); color: #fbbf24; border-color: rgba(245, 158, 11, 0.3); }
    .status-dot { width: 7px; height: 7px; border-radius: 50%; background: currentColor; box-shadow: 0 0 8px currentColor; }

    /* Header Nav Tabs */
    .header-nav {
      display: flex; align-items: center; gap: 6px;
      background: rgba(15, 23, 42, 0.6); padding: 4px; border-radius: var(--radius-md);
      border: 1px solid var(--border-subtle);
    }
    .nav-tab-btn {
      padding: 8px 16px; border-radius: var(--radius-sm); border: none; background: transparent;
      color: var(--text-muted); font-size: 0.85rem; font-weight: 600; cursor: pointer;
      display: flex; align-items: center; gap: 8px; transition: var(--transition);
    }
    .nav-tab-btn:hover { color: var(--text-main); background: rgba(255, 255, 255, 0.04); }
    .nav-tab-btn.active {
      color: #fff; background: var(--primary);
      box-shadow: 0 2px 12px var(--primary-glow);
    }

    .header-actions { display: flex; align-items: center; gap: 10px; }

    /* Buttons */
    .btn {
      display: inline-flex; align-items: center; justify-content: center; gap: 8px;
      padding: 9px 16px; border-radius: var(--radius-sm); font-size: 0.84rem; font-weight: 600;
      cursor: pointer; transition: var(--transition); border: 1px solid transparent; text-decoration: none;
    }
    .btn-primary {
      background: linear-gradient(135deg, var(--primary), #4f46e5); color: #fff;
      box-shadow: 0 4px 14px var(--primary-glow);
    }
    .btn-primary:hover { opacity: 0.95; transform: translateY(-1px); }
    .btn-secondary {
      background: rgba(255,255,255,0.06); border-color: var(--border-subtle); color: var(--text-main);
    }
    .btn-secondary:hover { background: rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.15); }
    .btn-accent {
      background: linear-gradient(135deg, #0891b2, var(--accent-cyan)); color: #fff;
      box-shadow: 0 4px 14px rgba(6, 182, 212, 0.3);
    }
    .btn-accent:hover { opacity: 0.95; transform: translateY(-1px); }
    .btn-danger {
      background: rgba(244, 63, 94, 0.15); border-color: rgba(244, 63, 94, 0.3); color: #fb7185;
    }
    .btn-danger:hover { background: rgba(244, 63, 94, 0.25); }
    .btn-sm { padding: 6px 10px; font-size: 0.78rem; border-radius: 6px; }

    /* Main Container */
    main.main-content {
      max-width: 1440px; margin: 0 auto; width: 100%; padding: 24px;
      flex: 1; display: flex; flex-direction: column; gap: 24px;
    }

    /* Subnet Banner */
    .banner-bar {
      background: linear-gradient(90deg, rgba(30, 41, 59, 0.7), rgba(15, 23, 42, 0.8));
      border: 1px solid var(--border-subtle); border-radius: var(--radius-lg);
      padding: 14px 20px; display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px;
    }
    .banner-info { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
    .info-item { display: flex; flex-direction: column; gap: 2px; }
    .info-label { font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-dim); font-weight: 600; }
    .info-value { font-size: 0.88rem; font-family: 'JetBrains Mono', monospace; font-weight: 600; color: #fff; }

    /* Page Views */
    .page-section { display: none; }
    .page-section.active { display: block; animation: fadeIn 0.25s ease-out; }
    @keyframes fadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }

    /* Cards Grid */
    .kpi-grid {
      display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
      gap: 16px; margin-bottom: 24px;
    }
    .kpi-card {
      background: var(--bg-surface); border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg); padding: 20px;
      display: flex; flex-direction: column; justify-content: space-between;
      transition: var(--transition); position: relative; overflow: hidden;
    }
    .kpi-card:hover { border-color: rgba(255,255,255,0.15); transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.3); }
    .kpi-card::before {
      content: ''; position: absolute; top: 0; left: 0; right: 0; height: 3px;
      background: linear-gradient(90deg, transparent, currentColor, transparent); opacity: 0.6;
    }
    .kpi-card.c-primary { color: var(--primary); }
    .kpi-card.c-cyan { color: var(--accent-cyan); }
    .kpi-card.c-emerald { color: var(--accent-emerald); }
    .kpi-card.c-amber { color: var(--accent-amber); }

    .kpi-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
    .kpi-title { font-size: 0.82rem; font-weight: 600; color: var(--text-muted); }
    .kpi-icon { width: 36px; height: 36px; border-radius: var(--radius-sm); background: rgba(255,255,255,0.05); display: flex; align-items: center; justify-content: center; font-size: 1.1rem; }
    .kpi-val { font-size: 1.85rem; font-weight: 800; color: #fff; line-height: 1.1; margin-bottom: 6px; }
    .kpi-sub { font-size: 0.74rem; color: var(--text-dim); }

    /* Charts & Analytics Dual Panel */
    .analytics-row {
      display: grid; grid-template-columns: 2fr 1fr; gap: 20px; margin-bottom: 24px;
    }
    @media (max-width: 1024px) { .analytics-row { grid-template-columns: 1fr; } }

    .panel-card {
      background: var(--bg-surface); border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg); padding: 22px;
      display: flex; flex-direction: column; gap: 16px;
    }
    .panel-head { display: flex; align-items: center; justify-content: space-between; }
    .panel-title { font-size: 0.98rem; font-weight: 700; color: #fff; display: flex; align-items: center; gap: 8px; }

    /* SVG Bandwidth Chart */
    .chart-container {
      width: 100%; height: 210px; position: relative;
      background: rgba(10, 15, 28, 0.6); border-radius: var(--radius-md);
      padding: 12px; border: 1px solid rgba(255,255,255,0.04);
    }
    svg.live-chart { width: 100%; height: 100%; overflow: visible; }

    /* Progress Breakdown */
    .progress-list { display: flex; flex-direction: column; gap: 14px; }
    .progress-item { display: flex; flex-direction: column; gap: 6px; }
    .progress-meta { display: flex; justify-content: space-between; font-size: 0.8rem; font-weight: 600; }
    .progress-meta span:first-child { color: var(--text-muted); }
    .progress-bar-bg { height: 8px; background: rgba(255,255,255,0.06); border-radius: 999px; overflow: hidden; }
    .progress-bar-fill { height: 100%; border-radius: 999px; transition: width 0.4s ease; }

    /* Safe Guard Banner */
    .guard-banner {
      background: linear-gradient(135deg, rgba(16, 185, 129, 0.1), rgba(6, 182, 212, 0.05));
      border: 1px solid rgba(16, 185, 129, 0.25); border-radius: var(--radius-lg);
      padding: 18px 22px; display: flex; align-items: flex-start; gap: 16px; margin-bottom: 24px;
    }
    .guard-icon { font-size: 26px; }
    .guard-text h4 { font-size: 0.94rem; color: #34d399; font-weight: 700; margin-bottom: 4px; }
    .guard-text p { font-size: 0.8rem; color: var(--text-muted); line-height: 1.5; }

    /* Tables */
    .table-wrapper {
      background: var(--bg-surface); border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg); overflow: hidden;
      box-shadow: 0 4px 20px rgba(0,0,0,0.25);
    }
    .table-toolbar {
      padding: 16px 20px; border-bottom: 1px solid var(--border-subtle);
      display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 14px;
    }
    .table-search-box {
      display: flex; align-items: center; gap: 10px;
      background: rgba(10, 15, 28, 0.8); border: 1px solid var(--border-subtle);
      padding: 7px 14px; border-radius: var(--radius-sm); width: 320px; max-width: 100%;
    }
    .table-search-box input {
      background: transparent; border: none; outline: none; color: #fff; font-size: 0.85rem; width: 100%;
    }
    .filter-tabs { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
    .filter-btn {
      padding: 6px 12px; border-radius: var(--radius-sm); border: 1px solid var(--border-subtle);
      background: transparent; color: var(--text-muted); font-size: 0.78rem; font-weight: 600; cursor: pointer;
    }
    .filter-btn.active { background: rgba(99, 102, 241, 0.15); border-color: var(--primary); color: #fff; }

    table.data-table { width: 100%; border-collapse: collapse; text-align: left; }
    table.data-table th {
      padding: 12px 18px; font-size: 0.74rem; text-transform: uppercase; letter-spacing: 0.05em;
      color: var(--text-dim); background: rgba(10, 15, 28, 0.5); font-weight: 700; border-bottom: 1px solid var(--border-subtle);
    }
    table.data-table td {
      padding: 14px 18px; font-size: 0.84rem; border-bottom: 1px solid rgba(255,255,255,0.04);
      color: var(--text-main); vertical-align: middle;
    }
    table.data-table tr:hover td { background: rgba(255,255,255,0.02); }
    .port-badge {
      font-family: 'JetBrains Mono', monospace; font-size: 0.9rem; font-weight: 700;
      color: #fff; background: rgba(255,255,255,0.06); padding: 3px 8px; border-radius: 6px;
    }
    .user-pass-box { font-family: 'JetBrains Mono', monospace; font-size: 0.78rem; color: var(--text-muted); }
    .user-pass-box strong { color: #fff; }
    .rot-badge {
      display: inline-flex; align-items: center; gap: 4px; padding: 3px 8px;
      border-radius: 6px; font-size: 0.74rem; font-weight: 600;
    }
    .rot-sticky { background: rgba(6, 182, 212, 0.15); color: #22d3ee; border: 1px solid rgba(6, 182, 212, 0.3); }
    .rot-request { background: rgba(99, 102, 241, 0.15); color: #a5b4fc; border: 1px solid rgba(99, 102, 241, 0.3); }
    .rot-static { background: rgba(148, 163, 184, 0.15); color: #cbd5e1; border: 1px solid rgba(148, 163, 184, 0.3); }

    /* Switch Toggle */
    .switch { position: relative; display: inline-block; width: 38px; height: 22px; }
    .switch input { opacity: 0; width: 0; height: 0; }
    .slider {
      position: absolute; cursor: pointer; inset: 0; background-color: rgba(255,255,255,0.12);
      transition: .25s; border-radius: 24px;
    }
    .slider:before {
      position: absolute; content: ""; height: 16px; width: 16px; left: 3px; bottom: 3px;
      background-color: white; transition: .25s; border-radius: 50%;
    }
    input:checked + .slider { background-color: var(--accent-emerald); }
    input:checked + .slider:before { transform: translateX(16px); }

    /* Modals */
    .modal-overlay {
      position: fixed; inset: 0; z-index: 1000;
      background: rgba(0, 0, 0, 0.75); backdrop-filter: blur(10px);
      display: none; align-items: center; justify-content: center; padding: 20px;
    }
    .modal-overlay.open { display: flex; animation: fadeIn 0.2s ease-out; }
    .modal-card {
      background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.12);
      border-radius: var(--radius-xl); width: 100%; max-width: 580px; max-height: 90vh;
      overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.6);
      display: flex; flex-direction: column;
    }
    .modal-header {
      padding: 20px 24px; border-bottom: 1px solid var(--border-subtle);
      display: flex; align-items: center; justify-content: space-between;
    }
    .modal-title { font-size: 1.1rem; font-weight: 700; color: #fff; }
    .modal-body { padding: 24px; display: flex; flex-direction: column; gap: 16px; }
    .modal-footer {
      padding: 16px 24px; border-top: 1px solid var(--border-subtle);
      display: flex; align-items: center; justify-content: flex-end; gap: 10px;
      background: rgba(0,0,0,0.15);
    }
    .close-btn { background: none; border: none; color: var(--text-dim); font-size: 1.3rem; cursor: pointer; }
    .close-btn:hover { color: #fff; }

    /* Form Fields */
    .form-group { display: flex; flex-direction: column; gap: 6px; }
    .form-group label { font-size: 0.8rem; font-weight: 600; color: var(--text-muted); }
    .form-control {
      background: rgba(15, 23, 42, 0.8); border: 1px solid var(--border-subtle);
      border-radius: var(--radius-sm); padding: 10px 14px; font-size: 0.88rem; color: #fff;
      outline: none; transition: var(--transition);
    }
    .form-control:focus { border-color: var(--primary); box-shadow: 0 0 0 3px var(--primary-glow); }
    .form-grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
    .form-hint { font-size: 0.74rem; color: var(--text-dim); line-height: 1.4; }

    /* Code Snippet Box */
    .code-box {
      background: #090d16; border: 1px solid var(--border-subtle); border-radius: var(--radius-md);
      padding: 14px; font-family: 'JetBrains Mono', monospace; font-size: 0.8rem; color: #cbd5e1;
      white-space: pre-wrap; word-break: break-all; position: relative; max-height: 250px; overflow-y: auto;
    }

    /* Auth Overlay */
    #login-overlay {
      position: fixed; inset: 0; z-index: 2000;
      background: radial-gradient(circle at center, #0e172a 0%, #060911 100%);
      display: flex; align-items: center; justify-content: center; padding: 20px;
    }
    .login-card {
      width: 100%; max-width: 400px; background: rgba(15, 23, 42, 0.85);
      backdrop-filter: blur(20px); border: 1px solid rgba(255,255,255,0.1);
      border-radius: var(--radius-xl); padding: 36px 30px; box-shadow: 0 25px 60px rgba(0,0,0,0.7);
      display: flex; flex-direction: column; gap: 20px; text-align: center;
    }

    /* Toast Notification */
    #toast-box {
      position: fixed; bottom: 24px; right: 24px; z-index: 9999;
      display: flex; flex-direction: column; gap: 10px; pointer-events: none;
    }
    .toast-msg {
      background: #1e293b; color: #fff; border: 1px solid var(--border-subtle);
      border-radius: var(--radius-sm); padding: 12px 18px; font-size: 0.84rem; font-weight: 500;
      box-shadow: 0 10px 30px rgba(0,0,0,0.5); pointer-events: auto; animation: slideIn 0.25s ease-out;
      display: flex; align-items: center; gap: 10px;
    }
    .toast-msg.success { border-left: 4px solid var(--accent-emerald); }
    .toast-msg.error { border-left: 4px solid var(--accent-rose); }
    @keyframes slideIn { from { transform: translateX(50px); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
  </style>
</head>
<body>
  <div class="ambient-bg"></div>

  <!-- LOGIN OVERLAY -->
  <div id="login-overlay" style="display: none;">
    <div class="login-card">
      <div style="display:flex; justify-content:center;">
        <div class="brand-logo" style="width:52px; height:52px; font-size:24px;">6</div>
      </div>
      <div>
        <h2 style="font-size:1.4rem; font-weight:800; color:#fff;">IPv6 Proxy Hub</h2>
        <p style="font-size:0.8rem; color:var(--text-muted); margin-top:4px;">Hệ Thống Quản Trị Proxy Doanh Nghiệp</p>
      </div>
      <form id="login-form" onsubmit="doLogin(event)" style="display:flex; flex-direction:column; gap:14px; text-align:left;">
        <div class="form-group">
          <label>Tài khoản</label>
          <input type="text" id="login-user" class="form-control" placeholder="admin" required autofocus>
        </div>
        <div class="form-group">
          <label>Mật khẩu</label>
          <input type="password" id="login-pass" class="form-control" placeholder="••••••••" required>
        </div>
        <button type="submit" class="btn btn-primary" style="margin-top:8px; padding:12px;">Đăng Nhập Quản Trị</button>
      </form>
    </div>
  </div>

  <div class="app-layout">
    <!-- SITE HEADER -->
    <header class="site-header">
      <div class="header-top">
        <!-- Brand & Badges -->
        <div class="brand-section">
          <div class="brand-logo">6</div>
          <div>
            <div class="brand-title">IPv6 PROXY HUB</div>
            <div class="brand-subtitle">
              <span>Linux Engine Enterprise</span>
              <span class="badge-pill success"><span class="status-dot"></span> NDP Safe Shield</span>
            </div>
          </div>
        </div>

        <!-- Distinct Header Navigation Tabs -->
        <nav class="header-nav">
          <button class="nav-tab-btn active" id="tab-btn-stats" onclick="switchTab('stats')">
            <span>📊</span> Thống Kê & Giám Sát
          </button>
          <button class="nav-tab-btn" id="tab-btn-proxies" onclick="switchTab('proxies')">
            <span>⚡</span> Quản Lý Cổng Proxy
          </button>
          <button class="nav-tab-btn" id="tab-btn-settings" onclick="switchTab('settings')">
            <span>🛠️</span> Cài Đặt & Hướng Dẫn
          </button>
        </nav>

        <!-- Right Quick Actions -->
        <div class="header-actions">
          <button class="btn btn-secondary btn-sm" onclick="loadAllData()" title="Làm mới dữ liệu">
            🔄 Làm Mới
          </button>
          <button class="btn btn-accent btn-sm" onclick="openBulkModal()">
            ⚡ Tạo Hàng Loạt
          </button>
          <button class="btn btn-primary btn-sm" onclick="openCreateModal()">
            + Thêm Cổng
          </button>
          <button class="btn btn-danger btn-sm" onclick="doLogout()" title="Đăng xuất">
            🚪
          </button>
        </div>
      </div>
    </header>

    <!-- MAIN BODY -->
    <main class="main-content">
      <!-- Subnet Overview Strip -->
      <div class="banner-bar">
        <div class="banner-info">
          <div class="info-item">
            <span class="info-label">IPv4 Máy Chủ (Host)</span>
            <span class="info-value" id="disp-ipv4">...</span>
          </div>
          <div class="info-item">
            <span class="info-label">Subnet Định Tuyến IPv6</span>
            <span class="info-value" id="disp-subnet">...</span>
          </div>
          <div class="info-item">
            <span class="info-label">Giao Thức Proxy</span>
            <span class="info-value" style="color:var(--accent-cyan);">HTTP & SOCKS5 (Auto)</span>
          </div>
          <div class="info-item">
            <span class="info-label">Uptime Hệ Thống</span>
            <span class="info-value" id="disp-uptime">0h 0m</span>
          </div>
        </div>
        <div style="display:flex; align-items:center; gap:8px;">
          <span class="badge-pill primary" id="disp-total-badge">0 Cổng Cấu Hình</span>
          <span class="badge-pill success" id="disp-active-badge">0 Đang Chạy</span>
        </div>
      </div>

      <!-- TAB 1: THỐNG KÊ & GIÁM SÁT (ANALYTICS & MONITORING) -->
      <section id="section-stats" class="page-section active">
        <!-- KPI Cards -->
        <div class="kpi-grid">
          <div class="kpi-card c-primary">
            <div class="kpi-head">
              <span class="kpi-title">CỔNG HOẠT ĐỘNG</span>
              <div class="kpi-icon">⚡</div>
            </div>
            <div class="kpi-val" id="kpi-active-ports">0 / 0</div>
            <div class="kpi-sub" id="kpi-active-desc">0% tổng số cổng đang mở</div>
          </div>

          <div class="kpi-card c-cyan">
            <div class="kpi-head">
              <span class="kpi-title">TỔNG LƯU LƯỢNG ĐÃ DÙNG</span>
              <div class="kpi-icon">📶</div>
            </div>
            <div class="kpi-val" id="kpi-total-traffic">0 MB</div>
            <div class="kpi-sub">Băng thông luân chuyển qua IPv6</div>
          </div>

          <div class="kpi-card c-emerald">
            <div class="kpi-head">
              <span class="kpi-title">TÀI NGUYÊN RAM VPS</span>
              <div class="kpi-icon">💻</div>
            </div>
            <div class="kpi-val" id="kpi-ram">0 MB</div>
            <div class="kpi-sub" id="kpi-goroutines">Goroutines: 0</div>
          </div>

          <div class="kpi-card c-amber">
            <div class="kpi-head">
              <span class="kpi-title">CƠ CHẾ XOAY (ROTATION)</span>
              <div class="kpi-icon">🔄</div>
            </div>
            <div class="kpi-val" id="kpi-rot-summary">0 Sticky / 0 Req</div>
            <div class="kpi-sub">Bảo toàn kết nối, chống chặn ISP</div>
          </div>
        </div>

        <!-- Safe Shield Explanation Banner -->
        <div class="guard-banner">
          <div class="guard-icon">🛡️</div>
          <div class="guard-text">
            <h4>Hệ Thống Bảo Vệ Băng Thông & Phân Giải NDP An Toàn (Anti-Ban VPS)</h4>
            <p>
              Khác với các tool thông thường thêm hàng ngàn IP ảo trực tiếp vào card mạng (gây sập bảng Neighbor Table của Gateway nhà mạng và bị khoá VPS),
              hệ thống sử dụng cơ chế <strong>Linux Non-Local Bind + NDP Proxying theo nhu cầu kết nối</strong>. Toàn bộ dải <code>/64</code> (18 tỷ tỷ IPv6)
              xoay ngẫu nhiên vô tận mà không gây nghẽn gateway nhà mạng.
            </p>
          </div>
        </div>

        <!-- Analytics Two Column Row -->
        <div class="analytics-row">
          <!-- Left: Live SVG Traffic Chart -->
          <div class="panel-card">
            <div class="panel-head">
              <span class="panel-title">📈 Biểu Đồ Lưu Lượng Realtime (Băng Thông / Giây)</span>
              <span class="badge-pill success"><span class="status-dot"></span> Đang Theo Dõi Trực Tiếp</span>
            </div>
            <div class="chart-container">
              <svg class="live-chart" id="traffic-svg" viewBox="0 0 500 180" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="#6366f1" stop-opacity="0.5"/>
                    <stop offset="100%" stop-color="#6366f1" stop-opacity="0.0"/>
                  </linearGradient>
                </defs>
                <path id="chart-area" d="" fill="url(#chartGrad)"></path>
                <path id="chart-line" d="" fill="none" stroke="#6366f1" stroke-width="2.5"></path>
              </svg>
            </div>
            <div style="display:flex; justify-content:space-between; font-size:0.75rem; color:var(--text-dim);">
              <span>Lịch sử 60 giây gần nhất</span>
              <span id="chart-peak-val">Đỉnh: 0 KB/s</span>
            </div>
          </div>

          <!-- Right: Rotation Breakdown -->
          <div class="panel-card">
            <div class="panel-head">
              <span class="panel-title">🔄 Phân Bổ Chế Độ Xoay IP</span>
            </div>
            <div class="progress-list">
              <div class="progress-item">
                <div class="progress-meta">
                  <span>Sticky (Giữ IP 5s/10s/30s)</span>
                  <span id="rot-sticky-count">0</span>
                </div>
                <div class="progress-bar-bg">
                  <div class="progress-bar-fill" id="rot-sticky-bar" style="width:0%; background:var(--accent-cyan);"></div>
                </div>
              </div>
              <div class="progress-item">
                <div class="progress-meta">
                  <span>Mỗi Request 1 IP Mới</span>
                  <span id="rot-request-count">0</span>
                </div>
                <div class="progress-bar-bg">
                  <div class="progress-bar-fill" id="rot-request-bar" style="width:0%; background:var(--primary);"></div>
                </div>
              </div>
              <div class="progress-item">
                <div class="progress-meta">
                  <span>IP Tĩnh Cố Định</span>
                  <span id="rot-static-count">0</span>
                </div>
                <div class="progress-bar-bg">
                  <div class="progress-bar-fill" id="rot-static-bar" style="width:0%; background:var(--text-muted);"></div>
                </div>
              </div>
            </div>

            <div style="background:rgba(255,255,255,0.03); border-radius:var(--radius-sm); padding:12px; font-size:0.75rem; color:var(--text-muted); line-height:1.4; margin-top:auto;">
              💡 <strong>Mẹo nuôi tài khoản:</strong> Hãy chọn chế độ <em>Sticky 10s - 30s</em> hoặc dùng <em>Session Key</em> để các luồng không bị đổi IP liên tục giữa chừng làm checkpoint tài khoản!
            </div>
          </div>
        </div>

        <!-- Top Bandwidth Proxies Table -->
        <div class="table-wrapper">
          <div class="table-toolbar">
            <span style="font-weight:700; font-size:0.95rem; color:#fff;">🔥 Top 5 Cổng Proxy Tiêu Thụ Nhiều Data Nhất</span>
            <button class="btn btn-secondary btn-sm" onclick="switchTab('proxies')">Xem Toàn Bộ Cổng →</button>
          </div>
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:60px;">Top</th>
                <th>Cổng (Port)</th>
                <th>Tài Khoản</th>
                <th>Chế Độ Xoay</th>
                <th>Lưu Lượng Dùng</th>
                <th>Trạng Thái</th>
              </tr>
            </thead>
            <tbody id="top-proxies-body">
              <tr><td colspan="6" style="text-align:center; color:var(--text-dim); padding:30px;">Đang tải dữ liệu...</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- TAB 2: QUẢN LÝ PROXY (PROXY MANAGER) -->
      <section id="section-proxies" class="page-section">
        <div class="table-wrapper">
          <!-- Toolbar with Search & Filter Tabs -->
          <div class="table-toolbar">
            <div style="display:flex; align-items:center; gap:12px; flex-wrap:wrap;">
              <div class="table-search-box">
                <span>🔍</span>
                <input type="text" id="proxy-search" placeholder="Tìm theo Port, User..." oninput="filterProxies()">
              </div>
              <div class="filter-tabs">
                <button class="filter-btn active" onclick="setFilter('all', this)">Tất Cả (<span id="count-all">0</span>)</button>
                <button class="filter-btn" onclick="setFilter('active', this)">Đang Bật (<span id="count-active">0</span>)</button>
                <button class="filter-btn" onclick="setFilter('sticky', this)">Sticky (Xoay theo giây)</button>
                <button class="filter-btn" onclick="setFilter('request', this)">Mỗi Request 1 IP</button>
              </div>
            </div>

            <div style="display:flex; align-items:center; gap:8px;">
              <button class="btn btn-secondary btn-sm" onclick="openExportModal()">
                📥 Xuất File (Export)
              </button>
              <button class="btn btn-accent btn-sm" onclick="openBulkModal()">
                ⚡ Tạo Hàng Loạt
              </button>
              <button class="btn btn-primary btn-sm" onclick="openCreateModal()">
                + Thêm Cổng
              </button>
            </div>
          </div>

          <!-- Proxy Data Table -->
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:50px;">STT</th>
                <th>Cổng Lắng Nghe</th>
                <th>Giao Thức</th>
                <th>Tài Khoản Xác Thực</th>
                <th>Chế Độ Xoay</th>
                <th>Lưu Lượng / Giới Hạn</th>
                <th>Hạn Sử Dụng</th>
                <th style="text-align:center;">Bật / Tắt</th>
                <th style="text-align:right;">Thao Tác</th>
              </tr>
            </thead>
            <tbody id="proxy-table-body">
              <tr><td colspan="9" style="text-align:center; color:var(--text-dim); padding:40px;">Đang tải danh sách proxy...</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- TAB 3: HƯỚNG DẪN & CÀI ĐẶT (INTEGRATION & SETTINGS) -->
      <section id="section-settings" class="page-section">
        <div style="display:grid; grid-template-columns: 1fr 1fr; gap:24px;">
          <!-- Left: Antidetect & Client Guide -->
          <div class="panel-card">
            <div class="panel-head">
              <span class="panel-title">🌐 Hướng Dẫn Sử Dụng Trình Duyệt Antidetect & Tool</span>
            </div>
            <div style="font-size:0.84rem; color:var(--text-muted); display:flex; flex-direction:column; gap:14px; line-height:1.5;">
              <div>
                <strong style="color:#fff;">1. Cấu hình vào Antidetect (Gologin, AdsPower, Hidemyacc...):</strong>
                <p style="margin-top:4px;">
                  - Loại Proxy: Chọn <code>HTTP</code> hoặc <code>SOCKS5</code> (hệ thống tự nhận diện).<br>
                  - IP / Host: <code id="guide-host">103.199.11.9</code><br>
                  - Port: Cổng bạn đã tạo (VD: <code>10000</code>).<br>
                  - User / Password: Tài khoản bạn đã đặt.
                </p>
              </div>

              <div style="background:rgba(99,102,241,0.08); border:1px solid rgba(99,102,241,0.25); border-radius:var(--radius-sm); padding:12px;">
                <strong style="color:#a5b4fc;">2. Cách dùng 50 luồng đồng thời không lo trùng IP (Session-Sticky):</strong>
                <p style="margin-top:4px; font-size:0.78rem;">
                  Hệ thống hỗ trợ <strong>Session-based Sticky IP</strong>! Bạn có thể cho 50 luồng dùng chung 1 cổng duy nhất,
                  chỉ cần trong username nối thêm mã luồng:
                  <br>
                  Luồng 1: <code>user-session-1</code> &nbsp;|&nbsp;
                  Luồng 2: <code>user-session-2</code> &nbsp;|&nbsp;
                  Luồng 50: <code>user-session-50</code>
                  <br>
                  Mỗi luồng sẽ nhận <strong>1 IPv6 riêng biệt</strong>, giữ nguyên trong số giây đã cài (5s/10s/30s) mà không bao giờ bị đụng độ nhau!
                </p>
              </div>

              <div>
                <strong style="color:#fff;">3. Test Nhanh Bằng cURL:</strong>
                <div class="code-box" id="guide-curl-box">curl -x http://user:pass@103.199.11.9:10000 https://api64.ipify.org?format=json</div>
              </div>
            </div>
          </div>

          <!-- Right: System Password & Config -->
          <div class="panel-card">
            <div class="panel-head">
              <span class="panel-title">🔒 Đổi Mật Khẩu Quản Trị Web</span>
            </div>
            <form onsubmit="changeAdminSettings(event)" style="display:flex; flex-direction:column; gap:14px;">
              <div class="form-group">
                <label>Mật Khẩu Hiện Tại</label>
                <input type="password" id="curr-pass" class="form-control" placeholder="••••••••" required>
              </div>
              <div class="form-group">
                <label>Tên Đăng Nhập Mới (Để trống nếu không đổi)</label>
                <input type="text" id="new-user" class="form-control" placeholder="admin">
              </div>
              <div class="form-group">
                <label>Mật Khẩu Mới</label>
                <input type="password" id="new-pass" class="form-control" placeholder="••••••••" required>
              </div>
              <button type="submit" class="btn btn-primary" style="margin-top:6px;">Lưu Thay Đổi</button>
            </form>

            <div style="margin-top:20px; padding-top:20px; border-top:1px solid var(--border-subtle);">
              <span style="font-size:0.85rem; font-weight:700; color:#fff;">Thông tin tiến trình dịch vụ:</span>
              <ul style="font-size:0.78rem; color:var(--text-muted); margin-top:8px; line-height:1.6; padding-left:18px;">
                <li>Systemd Service: <code>ipv6-proxy.service</code></li>
                <li>Thư mục lưu trữ tài khoản: <code>/root/IPv6-Gen-System/linux/proxies_data.json</code></li>
                <li>Kernel Nonlocal Bind: <code>net.ipv6.ip_nonlocal_bind = 1</code> (Đang hoạt động)</li>
              </ul>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>

  <!-- MODAL: TẠO / SỬA PROXY ĐƠN -->
  <div class="modal-overlay" id="modal-proxy">
    <div class="modal-card">
      <div class="modal-header">
        <h3 class="modal-title" id="modal-proxy-title">Thêm Cổng Proxy Mới</h3>
        <button class="close-btn" onclick="closeModal('modal-proxy')">&times;</button>
      </div>
      <form id="form-proxy" onsubmit="saveProxy(event)">
        <input type="hidden" id="p-id">
        <div class="modal-body">
          <div class="form-grid-2">
            <div class="form-group">
              <label>Cổng Lắng Nghe (Port) *</label>
              <input type="number" id="p-port" class="form-control" placeholder="10000" min="1024" max="65535" required>
            </div>
            <div class="form-group">
              <label>Hạn Sử Dụng (Số ngày)</label>
              <input type="number" id="p-days" class="form-control" placeholder="0 = Không giới hạn" min="0" value="30">
            </div>
          </div>

          <div class="form-grid-2">
            <div class="form-group">
              <label>Tài Khoản (Username)</label>
              <input type="text" id="p-user" class="form-control" placeholder="Để trống = Không pass">
            </div>
            <div class="form-group">
              <label>Mật Khẩu (Password)</label>
              <div style="display:flex; gap:6px;">
                <input type="text" id="p-pass" class="form-control" placeholder="Để trống = Không pass">
                <button type="button" class="btn btn-secondary btn-sm" onclick="genRandPass('p-pass')">🎲</button>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>Chế Độ Xoay IPv6 *</label>
            <select id="p-rot-type" class="form-control" onchange="toggleStickyInput(this.value, 'p-sticky-group')">
              <option value="sticky">Xoay Giữ IP Theo Thời Gian (Sticky Seconds)</option>
              <option value="request">Mỗi Request 1 IP Mới (Xoay Liên Tục)</option>
              <option value="static">Cố Định 1 IPv6 (Static IP)</option>
            </select>
          </div>

          <div class="form-group" id="p-sticky-group">
            <label>Thời Gian Giữ 1 IP (Giây) *</label>
            <input type="number" id="p-sticky-sec" class="form-control" value="10" min="2" max="86400">
            <span class="form-hint">Ví dụ: Điền <strong>5</strong> (5 giây), <strong>10</strong> (10 giây), hoặc <strong>30</strong> (30 giây). Hết thời gian này lượt gọi tiếp theo sẽ tự nhảy sang IPv6 mới.</span>
          </div>

          <div class="form-group">
            <label>Giới Hạn Dung Lượng (GB)</label>
            <input type="number" id="p-maxgb" class="form-control" placeholder="0 = Không giới hạn" min="0" step="0.5" value="0">
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick="closeModal('modal-proxy')">Hủy</button>
          <button type="submit" class="btn btn-primary">Lưu Cổng Proxy</button>
        </div>
      </form>
    </div>
  </div>

  <!-- MODAL: TẠO HÀNG LOẠT (BULK GENERATE) -->
  <div class="modal-overlay" id="modal-bulk">
    <div class="modal-card">
      <div class="modal-header">
        <h3 class="modal-title">⚡ Tạo Cổng Proxy Hàng Loạt (Bulk Generator)</h3>
        <button class="close-btn" onclick="closeModal('modal-bulk')">&times;</button>
      </div>
      <form onsubmit="submitBulkProxies(event)">
        <div class="modal-body">
          <div class="form-grid-2">
            <div class="form-group">
              <label>Cổng Bắt Đầu (Start Port) *</label>
              <input type="number" id="b-start-port" class="form-control" value="10000" min="1024" max="65500" required>
            </div>
            <div class="form-group">
              <label>Số Lượng Cổng Cần Tạo *</label>
              <input type="number" id="b-count" class="form-control" value="10" min="1" max="500" required>
            </div>
          </div>

          <div class="form-grid-2">
            <div class="form-group">
              <label>Tiền Tố Username</label>
              <input type="text" id="b-prefix" class="form-control" value="proxy" placeholder="proxy">
            </div>
            <div class="form-group">
              <label>Mật Khẩu Chung</label>
              <input type="text" id="b-pass" class="form-control" placeholder="Để trống = Random riêng">
            </div>
          </div>

          <div class="form-group">
            <label>Chế Độ Xoay IPv6 *</label>
            <select id="b-rot-type" class="form-control" onchange="toggleStickyInput(this.value, 'b-sticky-group')">
              <option value="sticky">Xoay Giữ IP Theo Thời Gian (Sticky Seconds)</option>
              <option value="request">Mỗi Request 1 IP Mới (Xoay Liên Tục)</option>
              <option value="static">Cố Định 1 IPv6 (Static IP)</option>
            </select>
          </div>

          <div class="form-group" id="b-sticky-group">
            <label>Thời Gian Giữ IP (Giây) *</label>
            <input type="number" id="b-sticky-sec" class="form-control" value="10" min="2" max="86400">
            <span class="form-hint">Mỗi cổng trong danh sách sẽ có 1 IP riêng biệt và tự đổi IP sau đúng số giây này!</span>
          </div>

          <div class="form-grid-2">
            <div class="form-group">
              <label>Hạn Sử Dụng (Ngày)</label>
              <input type="number" id="b-days" class="form-control" value="30" min="0">
            </div>
            <div class="form-group">
              <label>Giới Hạn GB (Mỗi Cổng)</label>
              <input type="number" id="b-maxgb" class="form-control" value="0" min="0">
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick="closeModal('modal-bulk')">Hủy</button>
          <button type="submit" class="btn btn-accent">⚡ Bắt Đầu Tạo Ngay</button>
        </div>
      </form>
    </div>
  </div>

  <!-- MODAL: XUẤT DANH SÁCH (EXPORT) -->
  <div class="modal-overlay" id="modal-export">
    <div class="modal-card">
      <div class="modal-header">
        <h3 class="modal-title">📥 Xuất Danh Sách Proxy</h3>
        <button class="close-btn" onclick="closeModal('modal-export')">&times;</button>
      </div>
      <div class="modal-body">
        <div style="display:flex; gap:10px;">
          <button class="btn btn-secondary btn-sm" onclick="fetchExport('ip_port_user_pass')">Dạng IP:Port:User:Pass</button>
          <button class="btn btn-secondary btn-sm" onclick="fetchExport('json')">Dạng JSON</button>
        </div>
        <div class="code-box" id="export-box" style="height:250px;">Đang tải...</div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick="closeModal('modal-export')">Đóng</button>
        <button type="button" class="btn btn-primary" onclick="copyBoxText('export-box')">📋 Sao Chép Toàn Bộ</button>
      </div>
    </div>
  </div>

  <!-- MODAL: XEM CODE SNIPPET -->
  <div class="modal-overlay" id="modal-snippet">
    <div class="modal-card">
      <div class="modal-header">
        <h3 class="modal-title" id="snippet-title">Code Mẫu Tích Hợp</h3>
        <button class="close-btn" onclick="closeModal('modal-snippet')">&times;</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>cURL Command</label>
          <div class="code-box" id="snip-curl">...</div>
        </div>
        <div class="form-group">
          <label>Python (requests) với Session Sticky</label>
          <div class="code-box" id="snip-python">...</div>
        </div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick="closeModal('modal-snippet')">Đóng</button>
      </div>
    </div>
  </div>

  <div id="toast-box"></div>

  <!-- JAVASCRIPT LOGIC -->
  <script>
    var globalProxies = [];
    var serverStats = {};
    var currentFilter = 'all';
    var chartHistory = [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0];
    var prevTrafficBytes = 0;

    function showToast(msg, type) {
      type = type || 'success';
      var box = document.getElementById('toast-box');
      var div = document.createElement('div');
      div.className = 'toast-msg ' + type;
      div.innerHTML = (type === 'success' ? '✅ ' : '⚠️ ') + msg;
      box.appendChild(div);
      setTimeout(function() { div.remove(); }, 3500);
    }

    function openModal(id) { document.getElementById(id).classList.add('open'); }
    function closeModal(id) { document.getElementById(id).classList.remove('open'); }

    function switchTab(tab) {
      document.querySelectorAll('.page-section').forEach(function(s) { s.classList.remove('active'); });
      document.querySelectorAll('.nav-tab-btn').forEach(function(b) { b.classList.remove('active'); });
      
      var sec = document.getElementById('section-' + tab);
      var btn = document.getElementById('tab-btn-' + tab);
      if (sec) sec.classList.add('active');
      if (btn) btn.classList.add('active');
    }

    function genRandPass(targetId) {
      var chars = 'abcdefghjkmnpqrstuvwxyz23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
      var res = '';
      for (var i = 0; i < 8; i++) res += chars[Math.floor(Math.random() * chars.length)];
      document.getElementById(targetId).value = res;
    }

    function toggleStickyInput(val, targetId) {
      var el = document.getElementById(targetId);
      if (el) el.style.display = (val === 'sticky') ? 'flex' : 'none';
    }

    function formatBytes(bytes) {
      if (!bytes || bytes <= 0) return '0 B';
      var k = 1024;
      var sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      var i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    async function doLogin(e) {
      e.preventDefault();
      var u = document.getElementById('login-user').value.trim();
      var p = document.getElementById('login-pass').value.trim();
      try {
        var res = await fetch('/api/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: u, password: p })
        });
        var data = await res.json();
        if (res.ok) {
          document.getElementById('login-overlay').style.display = 'none';
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
      if (!confirm('Bạn có chắc chắn muốn đăng xuất?')) return;
      await fetch('/api/logout', { method: 'POST' });
      document.getElementById('login-overlay').style.display = 'flex';
      showToast('Đã đăng xuất');
    }

    async function loadAllData() {
      try {
        var results = await Promise.all([
          fetch('/api/stats'),
          fetch('/api/proxies')
        ]);
        var statsRes = results[0];
        var proxiesRes = results[1];

        if (statsRes.status === 401 || proxiesRes.status === 401) {
          document.getElementById('login-overlay').style.display = 'flex';
          return;
        }

        serverStats = await statsRes.json();
        globalProxies = await proxiesRes.json();

        renderHeaderAndBanners();
        renderStats();
        renderProxiesTable();
      } catch (err) {
        console.error('Error loading data:', err);
      }
    }

    function renderHeaderAndBanners() {
      var ip = serverStats.public_ip || serverStats.public_ipv4 || '103.199.11.9';
      document.getElementById('disp-ipv4').textContent = ip;
      document.getElementById('disp-subnet').textContent = serverStats.prefix || 'N/A';
      document.getElementById('guide-host').textContent = ip;
      document.getElementById('disp-uptime').textContent = serverStats.uptime || '0h 0m';

      document.getElementById('disp-total-badge').textContent = (serverStats.total_proxies || 0) + ' Cổng Cấu Hình';
      document.getElementById('disp-active-badge').textContent = (serverStats.active_proxies || 0) + ' Đang Chạy';
    }

    function renderStats() {
      var total = serverStats.total_proxies || 0;
      var active = serverStats.active_proxies || 0;
      document.getElementById('kpi-active-ports').textContent = active + ' / ' + total;
      var pct = total > 0 ? Math.round((active / total) * 100) : 0;
      document.getElementById('kpi-active-desc').textContent = pct + '% tổng số cổng đang mở';

      document.getElementById('kpi-total-traffic').textContent = formatBytes(serverStats.total_bytes);
      document.getElementById('kpi-ram').textContent = serverStats.ram_usage || '1.5 MB';
      document.getElementById('kpi-goroutines').textContent = 'Goroutines: ' + (serverStats.goroutines || 0) + ' | Cores: ' + (serverStats.cpu_cores || 1);

      var rotCounts = serverStats.rotation_counts || serverStats.rotations || { sticky: 0, request: 0, static: 0 };
      document.getElementById('kpi-rot-summary').textContent = (rotCounts.sticky || 0) + ' Sticky / ' + (rotCounts.request || 0) + ' Req';

      var sumRot = (rotCounts.sticky || 0) + (rotCounts.request || 0) + (rotCounts.static || 0) || 1;
      document.getElementById('rot-sticky-count').textContent = rotCounts.sticky || 0;
      document.getElementById('rot-request-count').textContent = rotCounts.request || 0;
      document.getElementById('rot-static-count').textContent = rotCounts.static || 0;

      document.getElementById('rot-sticky-bar').style.width = Math.round(((rotCounts.sticky || 0) / sumRot) * 100) + '%';
      document.getElementById('rot-request-bar').style.width = Math.round(((rotCounts.request || 0) / sumRot) * 100) + '%';
      document.getElementById('rot-static-bar').style.width = Math.round(((rotCounts.static || 0) / sumRot) * 100) + '%';

      updateTrafficChart(serverStats.total_bytes);
      renderTopProxies(serverStats.top_proxies || []);
    }

    function updateTrafficChart(currentBytes) {
      if (prevTrafficBytes === 0) prevTrafficBytes = currentBytes;
      var diffBytes = Math.max(0, currentBytes - prevTrafficBytes);
      prevTrafficBytes = currentBytes;

      var rateKB = diffBytes / 1024 / 5;
      chartHistory.push(rateKB);
      chartHistory.shift();

      var maxVal = Math.max.apply(null, chartHistory);
      if (maxVal < 10) maxVal = 10;
      document.getElementById('chart-peak-val').textContent = 'Đỉnh: ' + maxVal.toFixed(1) + ' KB/s';

      var width = 500;
      var height = 180;
      var step = width / (chartHistory.length - 1);

      var points = [];
      for (var i = 0; i < chartHistory.length; i++) {
        var x = i * step;
        var y = height - (chartHistory[i] / maxVal) * (height - 20) - 10;
        points.push({ x: x, y: y });
      }

      var lineParts = [];
      for (var j = 0; j < points.length; j++) {
        lineParts.push((j === 0 ? 'M' : 'L') + points[j].x + ',' + points[j].y);
      }
      var lineD = lineParts.join(' ');
      var areaD = lineD + ' L' + width + ',' + height + ' L0,' + height + ' Z';

      document.getElementById('chart-line').setAttribute('d', lineD);
      document.getElementById('chart-area').setAttribute('d', areaD);
    }

    function renderTopProxies(topList) {
      var tbody = document.getElementById('top-proxies-body');
      if (!topList || topList.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center; color:var(--text-dim); padding:20px;">Chưa có lưu lượng phát sinh</td></tr>';
        return;
      }
      var html = '';
      for (var idx = 0; idx < topList.length; idx++) {
        var p = topList[idx];
        var badgeCls = (idx === 0 ? 'warning' : 'primary');
        var rotCls = (p.rotation_type === 'sticky' ? 'rot-sticky' : 'rot-request');
        var stateCls = (p.enabled ? 'success' : 'primary');
        var stateText = (p.enabled ? 'Đang Chạy' : 'Tạm Dừng');
        html += '<tr>' +
          '<td><span class="badge-pill ' + badgeCls + '">#' + (idx + 1) + '</span></td>' +
          '<td><span class="port-badge">:' + p.port + '</span></td>' +
          '<td><span class="user-pass-box"><strong>' + (p.username || 'No Auth') + '</strong></span></td>' +
          '<td><span class="rot-badge ' + rotCls + '">' + p.rotation_type + '</span></td>' +
          '<td style="font-weight:700; color:#fff;">' + formatBytes(p.bytes_used) + '</td>' +
          '<td><span class="badge-pill ' + stateCls + '">' + stateText + '</span></td>' +
        '</tr>';
      }
      tbody.innerHTML = html;
    }

    function setFilter(type, btn) {
      currentFilter = type;
      document.querySelectorAll('.filter-btn').forEach(function(b) { b.classList.remove('active'); });
      btn.classList.add('active');
      filterProxies();
    }

    function filterProxies() {
      var q = (document.getElementById('proxy-search').value || '').toLowerCase();
      var list = globalProxies.filter(function(p) {
        var matchQ = (p.port + '').indexOf(q) !== -1 || (p.username || '').toLowerCase().indexOf(q) !== -1;
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
      document.getElementById('count-active').textContent = globalProxies.filter(function(p) { return p.enabled; }).length;
      filterProxies();
    }

    function renderProxyRows(list) {
      var tbody = document.getElementById('proxy-table-body');
      if (list.length === 0) {
        tbody.innerHTML = '<tr><td colspan="9" style="text-align:center; color:var(--text-dim); padding:40px;">Không tìm thấy cổng proxy nào phù hợp</td></tr>';
        return;
      }

      var html = '';
      for (var idx = 0; idx < list.length; idx++) {
        var p = list[idx];
        var rotBadge = '';
        if (p.rotation_type === 'sticky') {
          rotBadge = '<span class="rot-badge rot-sticky">⏱️ Sticky ' + (p.sticky_sec || 10) + 's</span>';
        } else if (p.rotation_type === 'static') {
          rotBadge = '<span class="rot-badge rot-static">📌 Static</span>';
        } else {
          rotBadge = '<span class="rot-badge rot-request">⚡ Mỗi Req 1 IP</span>';
        }

        var maxGBText = p.max_bytes > 0 ? (' / ' + formatBytes(p.max_bytes)) : '';
        var expireText = p.expires_at ? new Date(p.expires_at).toLocaleDateString('vi-VN') : 'Vĩnh viễn';
        var authText = p.username ? ('<strong>' + p.username + '</strong> : ' + p.password) : '<span style="color:var(--text-dim);">Không mật khẩu</span>';
        var chk = p.enabled ? 'checked' : '';

        html += '<tr>' +
          '<td style="color:var(--text-dim); font-size:0.75rem;">' + (idx + 1) + '</td>' +
          '<td><span class="port-badge">:' + p.port + '</span></td>' +
          '<td><span class="badge-pill primary">HTTP/SOCKS5</span></td>' +
          '<td><div class="user-pass-box">' + authText + '</div></td>' +
          '<td>' + rotBadge + '</td>' +
          '<td><strong style="color:#fff;">' + formatBytes(p.bytes_used) + '</strong><span style="color:var(--text-dim); font-size:0.75rem;">' + maxGBText + '</span></td>' +
          '<td style="font-size:0.8rem; color:var(--text-muted);">' + expireText + '</td>' +
          '<td style="text-align:center;"><label class="switch"><input type="checkbox" ' + chk + ' onchange="toggleProxy(\'' + p.id + '\', this.checked)"><span class="slider"></span></label></td>' +
          '<td style="text-align:right;"><div style="display:inline-flex; gap:6px;">' +
            '<button class="btn btn-secondary btn-sm" title="Copy Host:Port:User:Pass" onclick="copyProxyString(' + p.port + ', \'' + (p.username||'') + '\', \'' + (p.password||'') + '\')">📋</button>' +
            '<button class="btn btn-secondary btn-sm" title="Xem Code Mẫu" onclick="openSnippetModal(\'' + p.id + '\')">💻</button>' +
            '<button class="btn btn-secondary btn-sm" title="Chỉnh sửa" onclick="openEditModal(\'' + p.id + '\')">✏️</button>' +
            '<button class="btn btn-danger btn-sm" title="Xóa" onclick="deleteProxy(\'' + p.id + '\')">🗑️</button>' +
          '</div></td>' +
        '</tr>';
      }
      tbody.innerHTML = html;
    }

    function copyProxyString(port, u, p) {
      var host = serverStats.public_ipv4 || '103.199.11.9';
      var str = host + ':' + port;
      if (u && p) str += ':' + u + ':' + p;
      navigator.clipboard.writeText(str);
      showToast('Đã copy: ' + str);
    }

    async function toggleProxy(id, enabled) {
      try {
        var res = await fetch('/api/proxies/' + id, {
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
        var res = await fetch('/api/proxies/' + id, { method: 'DELETE' });
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
      document.getElementById('p-days').value = '30';
      document.getElementById('p-rot-type').value = 'sticky';
      document.getElementById('p-sticky-sec').value = '10';
      document.getElementById('p-maxgb').value = '0';
      toggleStickyInput('sticky', 'p-sticky-group');
      openModal('modal-proxy');
    }

    function openEditModal(id) {
      var p = globalProxies.find(function(x) { return x.id === id; });
      if (!p) return;
      document.getElementById('modal-proxy-title').textContent = 'Chỉnh Sửa Cổng Proxy :' + p.port;
      document.getElementById('p-id').value = p.id;
      document.getElementById('p-port').value = p.port;
      document.getElementById('p-port').disabled = false;
      document.getElementById('p-user').value = p.username || '';
      document.getElementById('p-pass').value = p.password || '';
      document.getElementById('p-days').value = '0';
      document.getElementById('p-rot-type').value = p.rotation_type || 'sticky';
      document.getElementById('p-sticky-sec').value = p.sticky_sec || 10;
      document.getElementById('p-maxgb').value = p.max_bytes ? (p.max_bytes / 1024 / 1024 / 1024).toFixed(1) : '0';
      toggleStickyInput(p.rotation_type || 'sticky', 'p-sticky-group');
      openModal('modal-proxy');
    }

    async function saveProxy(e) {
      e.preventDefault();
      var id = document.getElementById('p-id').value;
      var port = parseInt(document.getElementById('p-port').value);
      var user = document.getElementById('p-user').value.trim();
      var pass = document.getElementById('p-pass').value.trim();
      var rotType = document.getElementById('p-rot-type').value;
      var stickySec = parseInt(document.getElementById('p-sticky-sec').value) || 10;
      var maxGB = parseFloat(document.getElementById('p-maxgb').value) || 0;
      var days = parseInt(document.getElementById('p-days').value) || 0;

      var payload = {
        port: port,
        username: user,
        password: pass,
        rotation_type: rotType,
        sticky_sec: stickySec,
        max_gb: maxGB,
        expire_days: days
      };

      try {
        var url = id ? ('/api/proxies/' + id) : '/api/proxies';
        var method = id ? 'PUT' : 'POST';
        var res = await fetch(url, {
          method: method,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        var data = await res.json();
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

    function openBulkModal() {
      openModal('modal-bulk');
    }

    async function submitBulkProxies(e) {
      e.preventDefault();
      var startPort = parseInt(document.getElementById('b-start-port').value);
      var count = parseInt(document.getElementById('b-count').value);
      var prefix = document.getElementById('b-prefix').value.trim();
      var pass = document.getElementById('b-pass').value.trim();
      var rotType = document.getElementById('b-rot-type').value;
      var stickySec = parseInt(document.getElementById('b-sticky-sec').value) || 10;
      var days = parseInt(document.getElementById('b-days').value) || 30;
      var maxgb = parseFloat(document.getElementById('b-maxgb').value) || 0;

      var payload = {
        start_port: startPort,
        count: count,
        username_prefix: prefix,
        password: pass,
        rotation_type: rotType,
        sticky_sec: stickySec,
        expire_days: days,
        max_gb: maxgb
      };

      try {
        var res = await fetch('/api/proxies/bulk', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        var data = await res.json();
        if (res.ok) {
          showToast('Đã tạo thành công ' + (data.created_count || count) + ' proxy!');
          closeModal('modal-bulk');
          loadAllData();
          switchTab('proxies');
        } else {
          showToast(data.error || 'Lỗi khi tạo hàng loạt', 'error');
        }
      } catch (err) { showToast('Lỗi kết nối', 'error'); }
    }

    function openExportModal() {
      openModal('modal-export');
      fetchExport('ip_port_user_pass');
    }

    async function fetchExport(fmt) {
      var box = document.getElementById('export-box');
      box.textContent = 'Đang trích xuất dữ liệu...';
      try {
        var res = await fetch('/api/export?format=' + fmt);
        var text = await res.text();
        box.textContent = text || 'Chưa có proxy nào đang bật.';
      } catch (e) { box.textContent = 'Lỗi trích xuất'; }
    }

    function copyBoxText(id) {
      var txt = document.getElementById(id).textContent;
      navigator.clipboard.writeText(txt);
      showToast('Đã copy danh sách vào Clipboard!');
    }

    function openSnippetModal(id) {
      var p = globalProxies.find(function(x) { return x.id === id; });
      if (!p) return;
      var host = serverStats.public_ipv4 || '103.199.11.9';
      document.getElementById('snippet-title').textContent = 'Code Mẫu Proxy Cổng :' + p.port;
      
      var auth = (p.username && p.password) ? (p.username + ':' + p.password + '@') : '';
      var curlCmd = 'curl -x http://' + auth + host + ':' + p.port + ' https://api64.ipify.org?format=json';
      
      var pyCode = 'import requests\n\n' +
        '# Với Phương thức 1 (Session Sticky): Thay đổi session_id cho mỗi luồng riêng biệt\n' +
        '# để 50 luồng đồng thời dùng chung 1 cổng mà nhận 50 IP IPv6 khác nhau:\n' +
        'session_id = "thread_1"\n' +
        'username = "' + (p.username || 'user') + '-session-" + session_id\n\n' +
        'proxies = {\n' +
        '    "http": f"http://{username}:' + (p.password || 'pass') + '@' + host + ':' + p.port + '",\n' +
        '    "https": f"http://{username}:' + (p.password || 'pass') + '@' + host + ':' + p.port + '",\n' +
        '}\n\n' +
        'r = requests.get("https://api64.ipify.org?format=json", proxies=proxies, timeout=10)\n' +
        'print("IPv6 của luồng:", r.json()["ip"])';

      document.getElementById('snip-curl').textContent = curlCmd;
      document.getElementById('snip-python').textContent = pyCode;
      openModal('modal-snippet');
    }

    async function changeAdminSettings(e) {
      e.preventDefault();
      var currPass = document.getElementById('curr-pass').value;
      var newUser = document.getElementById('new-user').value.trim();
      var newPass = document.getElementById('new-pass').value;

      try {
        var res = await fetch('/api/settings', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ current_password: currPass, new_username: newUser, new_password: newPass })
        });
        var data = await res.json();
        if (res.ok) {
          showToast('Đã đổi mật khẩu thành công!');
          document.getElementById('curr-pass').value = '';
          document.getElementById('new-pass').value = '';
        } else {
          showToast(data.error || 'Lỗi đổi thông tin', 'error');
        }
      } catch (e) { showToast('Lỗi kết nối', 'error'); }
    }

    setInterval(function() {
      loadAllData();
    }, 5000);

    window.addEventListener('DOMContentLoaded', function() {
      loadAllData();
    });
  </script>
</body>
</html>`
