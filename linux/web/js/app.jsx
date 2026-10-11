const { useState, useEffect, useMemo, useRef, useCallback } = React;

// ----------------------------------------------------------------------
// Minimalist Crisp SVG Icons (Monochrome, No Childish Emojis)
// ----------------------------------------------------------------------
function Icon({ name, className = "w-4 h-4", ...props }) {
  switch (name) {
    case "server":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
          <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
          <line x1="6" y1="6" x2="6.01" y2="6"></line>
          <line x1="6" y1="18" x2="6.01" y2="18"></line>
        </svg>
      );
    case "network":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <rect x="16" y="16" width="6" height="6" rx="1"></rect>
          <rect x="2" y="16" width="6" height="6" rx="1"></rect>
          <rect x="9" y="2" width="6" height="6" rx="1"></rect>
          <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"></path>
          <path d="M12 12V8"></path>
        </svg>
      );
    case "folder":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"></path>
        </svg>
      );
    case "folder-plus":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"></path>
          <line x1="12" y1="10" x2="12" y2="16"></line>
          <line x1="9" y1="13" x2="15" y2="13"></line>
        </svg>
      );
    case "plus":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" {...props}>
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
      );
    case "refresh":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"></path>
        </svg>
      );
    case "activity":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
        </svg>
      );
    case "clock":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <circle cx="12" cy="12" r="10"></circle>
          <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
      );
    case "copy":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
        </svg>
      );
    case "code":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <polyline points="16 18 22 12 16 6"></polyline>
          <polyline points="8 6 2 12 8 18"></polyline>
        </svg>
      );
    case "edit":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
        </svg>
      );
    case "trash":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <polyline points="3 6 5 6 21 6"></polyline>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
        </svg>
      );
    case "search":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <circle cx="11" cy="11" r="8"></circle>
          <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
        </svg>
      );
    case "download":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
          <polyline points="7 10 12 15 17 10"></polyline>
          <line x1="12" y1="15" x2="12" y2="3"></line>
        </svg>
      );
    case "layers":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
          <polyline points="2 17 12 22 22 17"></polyline>
          <polyline points="2 12 12 17 22 12"></polyline>
        </svg>
      );
    case "sliders":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <line x1="4" y1="21" x2="4" y2="14"></line>
          <line x1="4" y1="10" x2="4" y2="3"></line>
          <line x1="12" y1="21" x2="12" y2="12"></line>
          <line x1="12" y1="8" x2="12" y2="3"></line>
          <line x1="20" y1="21" x2="20" y2="16"></line>
          <line x1="20" y1="12" x2="20" y2="3"></line>
          <line x1="1" y1="14" x2="7" y2="14"></line>
          <line x1="9" y1="8" x2="15" y2="8"></line>
          <line x1="17" y1="16" x2="23" y2="16"></line>
        </svg>
      );
    case "check":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" {...props}>
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      );
    case "x":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" {...props}>
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      );
    case "logout":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
          <polyline points="16 17 21 12 16 7"></polyline>
          <line x1="21" y1="12" x2="9" y2="12"></line>
        </svg>
      );
    case "book":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="1.75" viewBox="0 0 24 24" {...props}>
          <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
          <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
        </svg>
      );
    case "menu":
      return (
        <svg className={className} fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" {...props}>
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      );
    default:
      return null;
  }
}

// Format Byte Utility
function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

// ----------------------------------------------------------------------
// Realtime Bandwidth Chart with Chart.js
// ----------------------------------------------------------------------
function BandwidthChart({ totalBytes, isPaused }) {
  const canvasRef = useRef(null);
  const chartInstance = useRef(null);
  const prevBytesRef = useRef(0);
  const historyRef = useRef(new Array(30).fill(0));
  const labelsRef = useRef(new Array(30).fill(""));

  const [currentSpeed, setCurrentSpeed] = useState(0);
  const [peakSpeed, setPeakSpeed] = useState(0);

  // Initialize Chart.js once
  useEffect(() => {
    if (!canvasRef.current) return;
    const ctx = canvasRef.current.getContext("2d");

    // Initialize labels (-30s to 0s)
    const initLabels = [];
    for (let i = 29; i >= 0; i--) {
      initLabels.push(i === 0 ? "Hiện tại" : `-${i * 2}s`);
    }
    labelsRef.current = initLabels;

    // Gradient background
    const gradient = ctx.createLinearGradient(0, 0, 0, 260);
    gradient.addColorStop(0, "rgba(99, 102, 241, 0.35)");
    gradient.addColorStop(1, "rgba(99, 102, 241, 0.0)");

    chartInstance.current = new Chart(ctx, {
      type: "line",
      data: {
        labels: labelsRef.current,
        datasets: [
          {
            label: "Tốc độ tải (KB/s)",
            data: historyRef.current,
            borderColor: "#6366f1",
            borderWidth: 2,
            backgroundColor: gradient,
            fill: true,
            tension: 0.3,
            pointRadius: 2,
            pointHoverRadius: 5,
            pointHoverBackgroundColor: "#818cf8",
            pointHoverBorderColor: "#fff",
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        animation: { duration: 300 },
        interaction: {
          mode: "index",
          intersect: false,
        },
        plugins: {
          legend: { display: false },
          tooltip: {
            backgroundColor: "#0f172a",
            titleColor: "#94a3b8",
            bodyColor: "#f8fafc",
            borderColor: "#334155",
            borderWidth: 1,
            padding: 10,
            callbacks: {
              label: (ctx) => ` Tốc độ: ${ctx.parsed.y.toFixed(1)} KB/s (${(ctx.parsed.y / 1024).toFixed(2)} MB/s)`,
            },
          },
        },
        scales: {
          x: {
            grid: { color: "rgba(51, 65, 85, 0.25)" },
            ticks: {
              color: "#64748b",
              font: { size: 11 },
              maxTicksLimit: 7,
            },
          },
          y: {
            beginAtZero: true,
            grid: { color: "rgba(51, 65, 85, 0.25)" },
            ticks: {
              color: "#64748b",
              font: { size: 11 },
              callback: (val) => `${val} KB/s`,
            },
          },
        },
      },
    });

    return () => {
      if (chartInstance.current) {
        chartInstance.current.destroy();
      }
    };
  }, []);

  // Update chart data whenever totalBytes changes
  useEffect(() => {
    if (isPaused || !chartInstance.current) return;

    if (prevBytesRef.current === 0) {
      prevBytesRef.current = totalBytes;
      return;
    }

    const diff = Math.max(0, totalBytes - prevBytesRef.current);
    prevBytesRef.current = totalBytes;

    // Calculate KB/s based on 2-second tick estimation
    const speedKB = diff / 1024 / 2;
    setCurrentSpeed(speedKB);
    setPeakSpeed((prev) => Math.max(prev, speedKB));

    const hist = historyRef.current;
    hist.push(speedKB);
    hist.shift();

    chartInstance.current.data.datasets[0].data = [...hist];
    chartInstance.current.update("none");
  }, [totalBytes, isPaused]);

  return (
    <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-5 backdrop-blur-sm">
      {/* Metrics Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 mb-4 pb-4 border-b border-slate-800/60">
        <div>
          <h3 className="text-sm font-semibold text-slate-200 flex items-center gap-2">
            <Icon name="activity" className="w-4 h-4 text-indigo-400" />
            Băng Thông Lưu Lượng Realtime
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Giám sát tốc độ truyền tải trên toàn bộ các cổng proxy</p>
        </div>

        {/* 3 Metrics Cards */}
        <div className="flex items-center gap-5">
          <div className="text-right">
            <div className="text-xs text-slate-400">Tốc độ hiện tại</div>
            <div className="text-sm font-semibold text-emerald-400 font-mono">
              {currentSpeed.toFixed(1)} KB/s <span className="text-xs font-normal text-slate-400">({(currentSpeed / 1024).toFixed(2)} MB/s)</span>
            </div>
          </div>
          <div className="h-7 w-[1px] bg-slate-800"></div>
          <div className="text-right">
            <div className="text-xs text-slate-400">Tốc độ đỉnh</div>
            <div className="text-sm font-semibold text-indigo-400 font-mono">
              {peakSpeed.toFixed(1)} KB/s <span className="text-xs font-normal text-slate-400">({(peakSpeed / 1024).toFixed(2)} MB/s)</span>
            </div>
          </div>
          <div className="h-7 w-[1px] bg-slate-800"></div>
          <div className="text-right">
            <div className="text-xs text-slate-400">Tổng lưu lượng</div>
            <div className="text-sm font-semibold text-slate-100 font-mono">
              {formatBytes(totalBytes)}
            </div>
          </div>
        </div>
      </div>

      {/* Chart Canvas */}
      <div className="h-[220px] w-full relative">
        <canvas ref={canvasRef}></canvas>
      </div>
    </div>
  );
}

// ----------------------------------------------------------------------
// Login Screen Component
// ----------------------------------------------------------------------
function LoginView({ onLoginSuccess }) {
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg("");
    try {
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: username.trim(), password: password.trim() }),
      });
      const data = await res.json();
      if (res.ok) {
        onLoginSuccess();
      } else {
        setErrorMsg(data.error || "Tài khoản hoặc mật khẩu không chính xác!");
      }
    } catch {
      setErrorMsg("Không thể kết nối đến máy chủ. Vui lòng thử lại!");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-screen flex items-center justify-center p-4 bg-slate-950 text-slate-100 relative overflow-hidden">
      {/* Background subtle radial glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-indigo-600/10 rounded-full blur-3xl pointer-events-none"></div>

      <div className="relative z-10 bg-slate-900/90 border border-slate-800 rounded-2xl max-w-sm w-full p-8 shadow-2xl backdrop-blur-md">
        {/* Logo & Header */}
        <div className="flex flex-col items-center text-center mb-6">
          <div className="w-12 h-12 rounded-xl bg-indigo-600/20 border border-indigo-500/30 flex items-center justify-center text-indigo-400 mb-3 shadow-inner">
            <Icon name="server" className="w-6 h-6" />
          </div>
          <h2 className="text-lg font-bold text-slate-100">IPv6 Proxy Hub</h2>
          <p className="text-xs text-slate-400 mt-1">Đăng nhập trang quản trị máy chủ Proxy</p>
        </div>

        {errorMsg && (
          <div className="mb-4 p-3 rounded-lg bg-rose-950/50 border border-rose-500/40 text-rose-300 text-xs flex items-center gap-2 animate-fade-in">
            <Icon name="x" className="w-4 h-4 text-rose-400 shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-4 text-xs">
          <div>
            <label className="block text-slate-300 font-medium mb-1.5">Tài Khoản</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3.5 py-2.5 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-slate-300 font-medium mb-1.5">Mật Khẩu</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Nhập mật khẩu admin..."
              required
              autoFocus
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3.5 py-2.5 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-colors"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-2 py-2.5 px-4 rounded-lg font-semibold text-white bg-indigo-600 hover:bg-indigo-500 transition-colors shadow-md disabled:opacity-50 flex items-center justify-center gap-2 cursor-pointer"
          >
            {loading ? (
              <>
                <Icon name="refresh" className="w-4 h-4 animate-spin text-white" />
                <span>Đang xác thực...</span>
              </>
            ) : (
              <span>Đăng Nhập</span>
            )}
          </button>
        </form>

        <div className="mt-6 text-center text-[11px] text-slate-500">
          Mặc định: <code className="text-slate-400">admin</code> / <code className="text-slate-400">admin123</code>
        </div>
      </div>
    </div>
  );
}

// ----------------------------------------------------------------------
// Main Application Component
// ----------------------------------------------------------------------
function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(true);
  const [proxies, setProxies] = useState([]);
  const [stats, setStats] = useState(null);
  const [activeTab, setActiveTab] = useState("proxies"); // "proxies" | "stats" | "guide"

  // Filter States
  const [selectedFolder, setSelectedFolder] = useState("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [protoFilter, setProtoFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");

  // Polling State (User requested: Không cần cập nhật liên tục, có thể chọn giây hoặc tắt)
  const [refreshInterval, setRefreshInterval] = useState(30); // 0 = Tắt, 10, 30, 60
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(null);

  // Modals
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [bulkModalOpen, setBulkModalOpen] = useState(false);
  const [editProxy, setEditProxy] = useState(null);
  const [snippetProxy, setSnippetProxy] = useState(null);
  const [newFolderModalOpen, setNewFolderModalOpen] = useState(false);
  const [newFolderName, setNewFolderName] = useState("");
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Toast notifications
  const [toast, setToast] = useState(null);
  const showToast = (message, type = "success") => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3500);
  };

  // Fetch Data Callback
  const fetchData = useCallback(async (isManual = false) => {
    if (isManual) setIsRefreshing(true);
    try {
      const [resProxies, resStats] = await Promise.all([
        fetch("/api/proxies"),
        fetch("/api/stats"),
      ]);

      if (resProxies.status === 401 || resStats.status === 401) {
        setIsLoggedIn(false);
        return;
      }

      setIsLoggedIn(true);
      if (resProxies.ok) {
        const dataP = await resProxies.json();
        setProxies(dataP);
      }
      if (resStats.ok) {
        const dataS = await resStats.json();
        setStats(dataS);
      }
      setLastUpdated(new Date());
    } catch (err) {
      console.error("Lỗi cập nhật:", err);
    } finally {
      if (isManual) {
        setTimeout(() => setIsRefreshing(false), 300);
      }
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchData(true);
  }, [fetchData]);

  // Periodic Polling controlled by user preference
  useEffect(() => {
    if (refreshInterval <= 0) return; // 0 = Tắt tự động

    const timer = setInterval(() => {
      fetchData(false);
    }, refreshInterval * 1000);

    return () => clearInterval(timer);
  }, [refreshInterval, fetchData]);

  // Selection State for Table Multi-Row Actions
  const [selectedIds, setSelectedIds] = useState([]);

  // Persistent Custom Folders (so newly created folders NEVER disappear even if 0 proxies)
  const [customFolders, setCustomFolders] = useState(() => {
    try {
      const saved = localStorage.getItem("ipv6_proxy_folders");
      if (saved) {
        const parsed = JSON.parse(saved);
        if (Array.isArray(parsed) && parsed.length > 0) return parsed;
      }
    } catch (_) {}
    return ["Mặc định"];
  });

  // Keep all distinct folder names (custom + existing proxies)
  const allFolderNames = useMemo(() => {
    const set = new Set(["Mặc định", ...customFolders]);
    proxies.forEach((p) => {
      if (p.group && p.group.trim()) set.add(p.group.trim());
    });
    return Array.from(set);
  }, [customFolders, proxies]);

  // Distinct Folders calculation with exact counts
  const folders = useMemo(() => {
    const map = { all: proxies.length };
    allFolderNames.forEach((f) => {
      map[f] = 0;
    });
    proxies.forEach((p) => {
      const g = p.group && p.group.trim() !== "" ? p.group.trim() : "Mặc định";
      map[g] = (map[g] || 0) + 1;
    });
    return map;
  }, [allFolderNames, proxies]);

  // Create folder helper
  const handleCreateFolder = (name) => {
    const trimmed = name.trim();
    if (!trimmed) return;
    if (!customFolders.includes(trimmed)) {
      const updated = [...customFolders, trimmed];
      setCustomFolders(updated);
      try {
        localStorage.setItem("ipv6_proxy_folders", JSON.stringify(updated));
      } catch (_) {}
    }
    setSelectedFolder(trimmed);
    showToast(`Đã tạo thư mục: ${trimmed}`);
  };

  // Delete empty folder helper
  const handleDeleteFolder = (e, fName) => {
    e.stopPropagation();
    if (fName === "Mặc định" || fName === "all") return;
    const hasProxies = proxies.some((p) => (p.group || "Mặc định") === fName);
    if (hasProxies) {
      showToast("Không thể xóa thư mục đang có proxy bên trong!", "error");
      return;
    }
    const updated = customFolders.filter((f) => f !== fName);
    setCustomFolders(updated);
    try {
      localStorage.setItem("ipv6_proxy_folders", JSON.stringify(updated));
    } catch (_) {}
    if (selectedFolder === fName) setSelectedFolder("all");
    showToast(`Đã xóa thư mục: ${fName}`);
  };

  // Filtered Proxies calculation
  const filteredProxies = useMemo(() => {
    return proxies.filter((p) => {
      const pGroup = p.group && p.group.trim() !== "" ? p.group.trim() : "Mặc định";
      if (selectedFolder !== "all" && pGroup !== selectedFolder) {
        return false;
      }
      if (protoFilter !== "all") {
        if (protoFilter === "http" && p.proto !== "http") return false;
        if (protoFilter === "socks5" && p.proto !== "socks5") return false;
        if (protoFilter === "both" && p.proto !== "both") return false;
      }
      if (statusFilter === "active" && !p.enabled) return false;
      if (statusFilter === "disabled" && p.enabled) return false;

      if (searchQuery.trim() !== "") {
        const q = searchQuery.toLowerCase();
        const matchName = (p.name || "").toLowerCase().includes(q);
        const matchPort = String(p.port).includes(q);
        const matchUser = (p.username || "").toLowerCase().includes(q);
        const matchGroup = (p.group || "").toLowerCase().includes(q);
        return matchName || matchPort || matchUser || matchGroup;
      }
      return true;
    });
  }, [proxies, selectedFolder, protoFilter, statusFilter, searchQuery]);

  // Toggle Proxy Enabled
  const handleToggle = async (p, enabled) => {
    try {
      const res = await fetch(`/api/proxies/${p.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled }),
      });
      if (res.ok) {
        showToast(`Đã ${enabled ? "bật" : "tắt"} cổng :${p.port}`);
        fetchData(false);
      } else {
        showToast("Thao tác thất bại", "error");
      }
    } catch {
      showToast("Lỗi kết nối", "error");
    }
  };

  // Delete Proxy
  const handleDelete = async (p) => {
    if (!confirm(`Bạn có chắc muốn xóa cổng :${p.port} (${p.name || "Không tên"})?`)) return;
    try {
      const res = await fetch(`/api/proxies/${p.id}`, { method: "DELETE" });
      if (res.ok) {
        showToast(`Đã xóa cổng :${p.port}`);
        setSelectedIds((prev) => prev.filter((id) => id !== p.id));
        fetchData(false);
      } else {
        showToast("Không thể xóa", "error");
      }
    } catch {
      showToast("Lỗi kết nối", "error");
    }
  };

  // Quick Copy
  const copyToClipboard = (text, msg = "Đã sao chép!") => {
    navigator.clipboard.writeText(text);
    showToast(msg);
  };

  // Export File
  const handleExport = (fmt) => {
    window.open(`/api/export?format=${fmt}`, "_blank");
  };

  // Batch Operations
  const handleBatchChangeGroup = async (targetGroup) => {
    if (selectedIds.length === 0) return;
    try {
      const res = await fetch("/api/proxies/batch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "set_group", ids: selectedIds, group: targetGroup }),
      });
      const data = await res.json();
      if (res.ok) {
        showToast(data.message || `Đã chuyển ${selectedIds.length} proxy sang '${targetGroup}'`);
        setSelectedIds([]);
        fetchData(false);
      } else {
        showToast(data.error || "Thao tác thất bại", "error");
      }
    } catch {
      showToast("Lỗi kết nối", "error");
    }
  };

  const handleBatchSetEnabled = async (enabled) => {
    if (selectedIds.length === 0) return;
    try {
      const res = await fetch("/api/proxies/batch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "set_enabled", ids: selectedIds, enabled }),
      });
      const data = await res.json();
      if (res.ok) {
        showToast(data.message || `Đã ${enabled ? "bật" : "tắt"} ${selectedIds.length} proxy`);
        setSelectedIds([]);
        fetchData(false);
      } else {
        showToast(data.error || "Thao tác thất bại", "error");
      }
    } catch {
      showToast("Lỗi kết nối", "error");
    }
  };

  const handleBatchDelete = async () => {
    if (selectedIds.length === 0) return;
    if (!confirm(`Bạn có chắc muốn xóa vĩnh viễn ${selectedIds.length} proxy đã chọn?`)) return;
    try {
      const res = await fetch("/api/proxies/batch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "delete", ids: selectedIds }),
      });
      const data = await res.json();
      if (res.ok) {
        showToast(data.message || `Đã xóa ${selectedIds.length} proxy`);
        setSelectedIds([]);
        fetchData(false);
      } else {
        showToast(data.error || "Thao tác thất bại", "error");
      }
    } catch {
      showToast("Lỗi kết nối", "error");
    }
  };

  const handleBatchCopy = () => {
    if (selectedIds.length === 0) return;
    const host = stats?.public_ip || "103.199.11.9";
    const selectedProxies = proxies.filter((p) => selectedIds.includes(p.id));
    const lines = selectedProxies.map((p) =>
      p.username ? `${host}:${p.port}:${p.username}:${p.password}` : `${host}:${p.port}`
    );
    navigator.clipboard.writeText(lines.join("\n"));
    showToast(`Đã sao chép chuỗi kết nối ${lines.length} proxy!`);
  };

  const handleBatchExport = () => {
    if (selectedIds.length === 0) return;
    const host = stats?.public_ip || "103.199.11.9";
    const selectedProxies = proxies.filter((p) => selectedIds.includes(p.id));
    const lines = selectedProxies.map((p) =>
      p.username ? `${host}:${p.port}:${p.username}:${p.password}` : `${host}:${p.port}`
    );
    const blob = new Blob([lines.join("\n")], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `proxies_selected_${Date.now()}.txt`;
    a.click();
    URL.revokeObjectURL(url);
    showToast(`Đã tải xuống file cho ${lines.length} proxy!`);
  };

  if (!isLoggedIn) {
    return <LoginView onLoginSuccess={() => fetchData(true)} />;
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-slate-950 text-slate-100">
      {/* Toast Notification */}
      {toast && (
        <div className="fixed top-5 right-5 z-50 flex items-center gap-2.5 px-4 py-3 rounded-lg shadow-xl text-sm font-medium border bg-slate-900 border-slate-700 text-slate-100 animate-fade-in">
          <Icon
            name={toast.type === "error" ? "x" : "check"}
            className={`w-4 h-4 ${toast.type === "error" ? "text-rose-400" : "text-emerald-400"}`}
          />
          {toast.message}
        </div>
      )}

      {/* Mobile Backdrop */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm lg:hidden"
        ></div>
      )}

      {/* ------------------------------------------------------------------ */}
      {/* LEFT SIDEBAR                                                       */}
      {/* ------------------------------------------------------------------ */}
      <aside
        className={`fixed lg:static inset-y-0 left-0 z-50 w-64 bg-slate-900/90 border-r border-slate-800/80 flex flex-col transition-transform duration-300 backdrop-blur-md shrink-0 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        }`}
      >
        {/* Brand / Logo */}
        <div className="p-4 border-b border-slate-800/80 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-indigo-600/20 border border-indigo-500/30 flex items-center justify-center text-indigo-400 shrink-0">
              <Icon name="server" className="w-5 h-5" />
            </div>
            <div className="min-w-0">
              <h1 className="text-sm font-semibold text-slate-100 truncate">IPv6 Proxy Hub</h1>
              <span className="text-[11px] font-mono text-slate-400 block truncate">
                {stats?.prefix || "2001:.../64"}
              </span>
            </div>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden p-1.5 text-slate-400 hover:text-slate-200"
          >
            <Icon name="x" className="w-4 h-4" />
          </button>
        </div>

        {/* Navigation Menu */}
        <div className="p-3">
          <div className="text-[10px] font-semibold text-slate-500 uppercase tracking-wider px-3 mb-2">
            Điều Hướng Chính
          </div>
          <nav className="space-y-1">
            <button
              onClick={() => {
                setActiveTab("proxies");
                setSidebarOpen(false);
              }}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
                activeTab === "proxies"
                  ? "bg-indigo-600/15 text-indigo-300 border border-indigo-500/30"
                  : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Icon name="network" className="w-4 h-4" />
                <span>Quản Lý Proxy</span>
              </div>
              <span className="text-[10px] px-1.5 py-0.5 rounded-full font-mono bg-slate-800 text-slate-400">
                {proxies.length}
              </span>
            </button>

            <button
              onClick={() => {
                setActiveTab("stats");
                setSidebarOpen(false);
              }}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
                activeTab === "stats"
                  ? "bg-indigo-600/15 text-indigo-300 border border-indigo-500/30"
                  : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Icon name="activity" className="w-4 h-4" />
                <span>Thống Kê Băng Thông</span>
              </div>
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
            </button>

            <button
              onClick={() => {
                setActiveTab("guide");
                setSidebarOpen(false);
              }}
              className={`w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
                activeTab === "guide"
                  ? "bg-indigo-600/15 text-indigo-300 border border-indigo-500/30"
                  : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
              }`}
            >
              <Icon name="book" className="w-4 h-4" />
              <span>Hướng Dẫn Tích Hợp</span>
            </button>
          </nav>
        </div>

        {/* Folders List in Sidebar */}
        <div className="flex-1 p-3 overflow-y-auto border-t border-slate-800/60">
          <div className="flex items-center justify-between px-3 mb-2">
            <span className="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">
              Thư Mục / Nhóm
            </span>
            <button
              onClick={() => setNewFolderModalOpen(true)}
              title="Thêm thư mục mới"
              className="p-1 rounded text-slate-400 hover:text-indigo-300 hover:bg-slate-800/60 transition-colors"
            >
              <Icon name="folder-plus" className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="space-y-1">
            {Object.entries(folders).map(([fName, count]) => {
              const isSelected = selectedFolder === fName;
              return (
                <div
                  key={fName}
                  onClick={() => {
                    setSelectedFolder(fName);
                    setActiveTab("proxies");
                    setSidebarOpen(false);
                  }}
                  className={`w-full flex items-center justify-between px-3 py-1.5 rounded-lg text-xs font-medium transition-colors cursor-pointer group ${
                    isSelected && activeTab === "proxies"
                      ? "bg-slate-800 text-indigo-300 border border-indigo-500/40 shadow-sm"
                      : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center gap-2 truncate">
                    <Icon name="folder" className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span className="truncate">{fName === "all" ? "Tất Cả" : fName}</span>
                  </div>
                  <div className="flex items-center gap-1.5 shrink-0">
                    <span
                      className={`text-[10px] px-1.5 py-0.2 rounded-full font-mono ${
                        isSelected && activeTab === "proxies"
                          ? "bg-indigo-500/20 text-indigo-300"
                          : "bg-slate-800/80 text-slate-500"
                      }`}
                    >
                      {count}
                    </span>
                    {fName !== "all" && fName !== "Mặc định" && count === 0 && (
                      <button
                        onClick={(e) => handleDeleteFolder(e, fName)}
                        title="Xóa thư mục rỗng này"
                        className="opacity-0 group-hover:opacity-100 p-0.5 text-slate-500 hover:text-rose-400 transition-opacity"
                      >
                        <Icon name="trash" className="w-3 h-3" />
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Server Status Footer Widget */}
        <div className="p-3 border-t border-slate-800/80 bg-slate-950/40 space-y-2.5">
          <div className="bg-slate-900/60 border border-slate-800/80 rounded-lg p-2.5 text-[11px] space-y-1 font-mono">
            <div className="flex items-center justify-between text-slate-400">
              <span>VPS IP:</span>
              <span className="text-slate-200 font-medium">{stats?.public_ip || "103.199.11.9"}</span>
            </div>
            <div className="flex items-center justify-between text-slate-400">
              <span>Uptime:</span>
              <span className="text-slate-300">{stats?.uptime || "--"}</span>
            </div>
            <div className="flex items-center justify-between text-slate-400">
              <span>RAM:</span>
              <span className="text-slate-300">{stats?.ram_usage || "--"}</span>
            </div>
          </div>

          <div className="flex items-center justify-between px-2 pt-1 text-xs">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
              <span className="text-slate-300 font-medium">admin</span>
            </div>
            <button
              onClick={async () => {
                await fetch("/api/logout");
                setIsLoggedIn(false);
              }}
              title="Đăng xuất"
              className="text-slate-400 hover:text-rose-400 transition-colors flex items-center gap-1 cursor-pointer"
            >
              <Icon name="logout" className="w-3.5 h-3.5" />
              <span className="text-[11px]">Thoát</span>
            </button>
          </div>
        </div>
      </aside>

      {/* ------------------------------------------------------------------ */}
      {/* RIGHT MAIN WORKSPACE                                               */}
      {/* ------------------------------------------------------------------ */}
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {/* Top Header Bar */}
        <header className="h-16 border-b border-slate-800/80 bg-slate-900/50 backdrop-blur-md px-4 sm:px-6 flex items-center justify-between gap-4 shrink-0">
          {/* Left: Mobile Toggle & Page Title */}
          <div className="flex items-center gap-3">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden p-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800"
            >
              <Icon name="menu" className="w-5 h-5" />
            </button>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-semibold text-slate-100">
                  {activeTab === "proxies" && (
                    <span>
                      Quản Lý Proxy
                      {selectedFolder !== "all" && (
                        <span className="text-slate-400 font-normal ml-1.5">
                          / Thư mục: <strong className="text-indigo-300">{selectedFolder}</strong>
                        </span>
                      )}
                    </span>
                  )}
                  {activeTab === "stats" && "Thống Kê Băng Thông Realtime"}
                  {activeTab === "guide" && "Tài Liệu & Hướng Dẫn Tích Hợp"}
                </h2>
              </div>
              <div className="text-[11px] text-slate-400">
                {activeTab === "proxies" && `${filteredProxies.length} proxy hiển thị • Cập nhật lúc: ${lastUpdated ? lastUpdated.toLocaleTimeString() : "--"}`}
                {activeTab === "stats" && "Giám sát lưu lượng truyền tải và băng thông mạng"}
                {activeTab === "guide" && "Cấu hình kết nối cho Antidetect Browser và Tool Automation"}
              </div>
            </div>
          </div>

          {/* Right: Polling Interval & Manual Refresh */}
          <div className="flex items-center gap-2.5">
            {/* Polling Interval Select */}
            <div className="flex items-center gap-1.5 text-xs text-slate-400 bg-slate-900/80 border border-slate-800/80 rounded-lg px-2.5 py-1.5">
              <Icon name="clock" className="w-3.5 h-3.5 text-slate-400" />
              <span className="hidden sm:inline">Làm mới:</span>
              <select
                value={refreshInterval}
                onChange={(e) => setRefreshInterval(Number(e.target.value))}
                className="bg-transparent text-slate-200 text-xs font-medium focus:outline-none cursor-pointer"
              >
                <option value={0} className="bg-slate-900 text-slate-200">Tắt (Thủ công)</option>
                <option value={10} className="bg-slate-900 text-slate-200">10 giây</option>
                <option value={30} className="bg-slate-900 text-slate-200">30 giây</option>
                <option value={60} className="bg-slate-900 text-slate-200">60 giây</option>
              </select>
            </div>

            {/* Manual Refresh Button */}
            <button
              onClick={() => fetchData(true)}
              disabled={isRefreshing}
              title="Làm mới dữ liệu ngay"
              className="p-2 rounded-lg bg-slate-900/80 border border-slate-800/80 text-slate-300 hover:text-white hover:bg-slate-800 transition-colors disabled:opacity-50"
            >
              <Icon name="refresh" className={`w-4 h-4 ${isRefreshing ? "animate-spin text-indigo-400" : ""}`} />
            </button>
          </div>
        </header>

        {/* Scrollable Main Content */}
        <main className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
        {/* VIEW 1: PROXIES TABLE & FOLDER MANAGEMENT */}
        {activeTab === "proxies" && (
          <div className="flex flex-col gap-5">
            {/* Folder / Group Tabs Bar */}
            <div className="flex flex-wrap items-center justify-between gap-3 bg-slate-900/50 border border-slate-800/80 rounded-xl p-2.5 backdrop-blur-sm">
              <div className="flex flex-wrap items-center gap-1.5">
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider px-2 flex items-center gap-1.5">
                  <Icon name="folder" className="w-3.5 h-3.5 text-slate-400" />
                  Thư mục:
                </span>
                {Object.entries(folders).map(([fName, count]) => (
                  <button
                    key={fName}
                    onClick={() => setSelectedFolder(fName)}
                    className={`px-3 py-1 rounded-lg text-xs font-medium transition-all flex items-center gap-1.5 ${
                      selectedFolder === fName
                        ? "bg-slate-800 text-indigo-300 border border-indigo-500/40 shadow-sm"
                        : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40 border border-transparent"
                    }`}
                  >
                    <span>{fName === "all" ? "Tất Cả" : fName}</span>
                    <span
                      className={`text-[10px] px-1.5 py-0.2 rounded-full font-mono ${
                        selectedFolder === fName ? "bg-indigo-500/20 text-indigo-300" : "bg-slate-800 text-slate-500"
                      }`}
                    >
                      {count}
                    </span>
                  </button>
                ))}
              </div>

              {/* Add New Folder Button */}
              <button
                onClick={() => setNewFolderModalOpen(true)}
                className="text-xs text-slate-400 hover:text-indigo-300 flex items-center gap-1 px-2.5 py-1 rounded-md hover:bg-slate-800/50 transition-colors ml-auto"
              >
                <Icon name="folder-plus" className="w-3.5 h-3.5" />
                <span>+ Thêm thư mục</span>
              </button>
            </div>

            {/* Unified Toolbar (Single place for create/bulk actions - NO duplication) */}
            <div className="flex flex-wrap items-center justify-between gap-3 bg-slate-900/30 border border-slate-800/60 rounded-xl p-3">
              {/* Search & Filters */}
              <div className="flex flex-wrap items-center gap-2.5 flex-1 min-w-[280px]">
                {/* Search Input */}
                <div className="relative flex-1 max-w-xs">
                  <Icon name="search" className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Tìm theo port, ghi chú, user..."
                    className="w-full bg-slate-900/90 border border-slate-800 text-slate-200 text-xs rounded-lg pl-9 pr-3 py-2 placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                  />
                  {searchQuery && (
                    <button
                      onClick={() => setSearchQuery("")}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                    >
                      <Icon name="x" className="w-3.5 h-3.5" />
                    </button>
                  )}
                </div>

                {/* Protocol Filter */}
                <select
                  value={protoFilter}
                  onChange={(e) => setProtoFilter(e.target.value)}
                  className="bg-slate-900/90 border border-slate-800 text-slate-300 text-xs rounded-lg px-2.5 py-2 focus:outline-none focus:border-indigo-500 cursor-pointer"
                >
                  <option value="all">Tất cả giao thức</option>
                  <option value="both">HTTP & SOCKS5</option>
                  <option value="http">Chỉ HTTP(S)</option>
                  <option value="socks5">Chỉ SOCKS5</option>
                </select>

                {/* Status Filter */}
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="bg-slate-900/90 border border-slate-800 text-slate-300 text-xs rounded-lg px-2.5 py-2 focus:outline-none focus:border-indigo-500 cursor-pointer"
                >
                  <option value="all">Tất cả trạng thái</option>
                  <option value="active">Đang chạy</option>
                  <option value="disabled">Đã tắt</option>
                </select>
              </div>

              {/* Action Buttons (Unified) */}
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleExport("txt")}
                  className="px-3 py-2 rounded-lg text-xs font-medium text-slate-300 bg-slate-900 hover:bg-slate-800 border border-slate-800 transition-colors flex items-center gap-1.5"
                  title="Xuất file danh sách proxy"
                >
                  <Icon name="download" className="w-3.5 h-3.5" />
                  <span>Xuất file</span>
                </button>

                <button
                  onClick={() => setBulkModalOpen(true)}
                  className="px-3 py-2 rounded-lg text-xs font-medium text-slate-200 bg-slate-800 hover:bg-slate-700 border border-slate-700 transition-colors flex items-center gap-1.5"
                >
                  <Icon name="layers" className="w-3.5 h-3.5 text-indigo-400" />
                  <span>Tạo Hàng Loạt</span>
                </button>

                <button
                  onClick={() => setCreateModalOpen(true)}
                  className="px-3.5 py-2 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 transition-colors flex items-center gap-1.5 shadow-sm"
                >
                  <Icon name="plus" className="w-3.5 h-3.5" />
                  <span>+ Thêm Proxy</span>
                </button>
              </div>
            </div>

            {/* Batch Selection Action Bar (Appears when 1 or more proxies are selected) */}
            {selectedIds.length > 0 && (
              <div className="bg-indigo-950/40 border border-indigo-500/40 rounded-xl p-3 flex flex-wrap items-center justify-between gap-3 shadow-lg backdrop-blur-md animate-fade-in">
                <div className="flex items-center gap-3">
                  <span className="text-xs font-semibold text-indigo-300 bg-indigo-500/20 border border-indigo-500/30 px-2.5 py-1 rounded-lg flex items-center gap-1.5">
                    <Icon name="check" className="w-3.5 h-3.5 text-indigo-400" />
                    Đã chọn: <strong>{selectedIds.length}</strong> proxy
                  </span>
                  <button
                    onClick={() => setSelectedIds([])}
                    className="text-xs text-slate-400 hover:text-slate-200 underline cursor-pointer"
                  >
                    Bỏ chọn tất cả
                  </button>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  {/* Move to folder */}
                  <div className="flex items-center gap-1.5 bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1">
                    <Icon name="folder" className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                    <select
                      onChange={(e) => {
                        if (e.target.value) {
                          handleBatchChangeGroup(e.target.value);
                          e.target.value = "";
                        }
                      }}
                      defaultValue=""
                      className="bg-transparent text-xs text-slate-200 focus:outline-none cursor-pointer"
                    >
                      <option value="" disabled>Chuyển thư mục...</option>
                      {allFolderNames.map((f) => (
                        <option key={f} value={f} className="bg-slate-900 text-slate-200">{f}</option>
                      ))}
                    </select>
                  </div>

                  {/* Enable selected */}
                  <button
                    onClick={() => handleBatchSetEnabled(true)}
                    className="px-2.5 py-1.5 rounded-lg text-xs font-medium bg-emerald-950/60 hover:bg-emerald-900/70 border border-emerald-500/40 text-emerald-300 flex items-center gap-1 transition-colors"
                  >
                    <Icon name="check" className="w-3.5 h-3.5" />
                    <span>Bật</span>
                  </button>

                  {/* Disable selected */}
                  <button
                    onClick={() => handleBatchSetEnabled(false)}
                    className="px-2.5 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-300 flex items-center gap-1 transition-colors"
                  >
                    <Icon name="x" className="w-3.5 h-3.5" />
                    <span>Tắt</span>
                  </button>

                  {/* Copy selected */}
                  <button
                    onClick={handleBatchCopy}
                    className="px-2.5 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 flex items-center gap-1 transition-colors"
                  >
                    <Icon name="copy" className="w-3.5 h-3.5" />
                    <span>Sao chép chuỗi</span>
                  </button>

                  {/* Export selected */}
                  <button
                    onClick={handleBatchExport}
                    className="px-2.5 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 flex items-center gap-1 transition-colors"
                  >
                    <Icon name="download" className="w-3.5 h-3.5" />
                    <span>Xuất file</span>
                  </button>

                  {/* Delete selected */}
                  <button
                    onClick={handleBatchDelete}
                    className="px-2.5 py-1.5 rounded-lg text-xs font-medium bg-rose-950/60 hover:bg-rose-900/70 border border-rose-500/40 text-rose-300 flex items-center gap-1 transition-colors"
                  >
                    <Icon name="trash" className="w-3.5 h-3.5" />
                    <span>Xóa đã chọn</span>
                  </button>
                </div>
              </div>
            )}

            {/* Proxy Data Table */}
            <div className="bg-slate-900/40 border border-slate-800/80 rounded-xl overflow-hidden shadow-sm">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-900/80 text-slate-400 border-b border-slate-800/80 font-medium">
                    <tr>
                      <th className="py-3 px-3 w-10 text-center">
                        <input
                          type="checkbox"
                          checked={filteredProxies.length > 0 && filteredProxies.every((p) => selectedIds.includes(p.id))}
                          onChange={() => {
                            const allSelected = filteredProxies.length > 0 && filteredProxies.every((p) => selectedIds.includes(p.id));
                            if (allSelected) {
                              setSelectedIds([]);
                            } else {
                              const visibleIds = filteredProxies.map((p) => p.id);
                              setSelectedIds(Array.from(new Set([...selectedIds, ...visibleIds])));
                            }
                          }}
                          className="w-4 h-4 rounded text-indigo-600 bg-slate-800 border-slate-700 focus:ring-indigo-500 cursor-pointer"
                        />
                      </th>
                      <th className="py-3 px-3">Cổng (Port)</th>
                      <th className="py-3 px-3">Ghi Chú / Tên Khách</th>
                      <th className="py-3 px-3">Thư Mục</th>
                      <th className="py-3 px-3">Giao Thức</th>
                      <th className="py-3 px-3">Tài Khoản Xác Thực</th>
                      <th className="py-3 px-3">Chế Độ Xoay</th>
                      <th className="py-3 px-3">Dung Lượng</th>
                      <th className="py-3 px-3">Hạn Dùng</th>
                      <th className="py-3 px-3 text-center w-16">Bật/Tắt</th>
                      <th className="py-3 px-3 text-right">Thao Tác</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/50">
                    {filteredProxies.length === 0 ? (
                      <tr>
                        <td colSpan="11" className="py-12 text-center text-slate-500 text-xs">
                          Không tìm thấy cổng proxy nào phù hợp
                        </td>
                      </tr>
                    ) : (
                      filteredProxies.map((p, idx) => {
                        const isSelected = selectedIds.includes(p.id);
                        const maxText = p.max_bytes > 0 ? ` / ${formatBytes(p.max_bytes)}` : "";
                        const expDate = p.expires_at ? new Date(p.expires_at).toLocaleDateString("vi-VN") : "Vĩnh viễn";

                        return (
                          <tr key={p.id} className={`transition-colors ${isSelected ? "bg-indigo-950/30 hover:bg-indigo-950/50" : "hover:bg-slate-900/60"}`}>
                            <td className="py-3 px-3 text-center">
                              <input
                                type="checkbox"
                                checked={isSelected}
                                onChange={() => {
                                  setSelectedIds((prev) =>
                                    prev.includes(p.id) ? prev.filter((id) => id !== p.id) : [...prev, p.id]
                                  );
                                }}
                                className="w-4 h-4 rounded text-indigo-600 bg-slate-800 border-slate-700 focus:ring-indigo-500 cursor-pointer"
                              />
                            </td>

                            {/* Port */}
                            <td className="py-3 px-3 font-mono font-semibold text-slate-100">
                              :{p.port}
                            </td>

                            {/* Note / Customer Name */}
                            <td className="py-3 px-3 font-medium text-slate-200">
                              <span className="text-slate-200 font-medium">{p.name || "—"}</span>
                            </td>

                            {/* Folder */}
                            <td className="py-3 px-3">
                              <span className="px-2 py-0.5 rounded text-[11px] bg-slate-800 text-slate-400 border border-slate-700/60">
                                {p.group || "Mặc định"}
                              </span>
                            </td>

                            {/* Protocol */}
                            <td className="py-3 px-3">
                              {p.proto === "http" && (
                                <span className="px-2 py-0.5 rounded text-[10px] font-mono font-medium bg-blue-500/10 text-blue-300 border border-blue-500/20">
                                  HTTP(S)
                                </span>
                              )}
                              {p.proto === "socks5" && (
                                <span className="px-2 py-0.5 rounded text-[10px] font-mono font-medium bg-amber-500/10 text-amber-300 border border-amber-500/20">
                                  SOCKS5
                                </span>
                              )}
                              {(!p.proto || p.proto === "both") && (
                                <span className="px-2 py-0.5 rounded text-[10px] font-mono font-medium bg-emerald-500/10 text-emerald-300 border border-emerald-500/20">
                                  HTTP & SOCKS5
                                </span>
                              )}
                            </td>

                            {/* Credentials */}
                            <td className="py-3 px-3 font-mono text-[11px] text-slate-300">
                              {p.username ? (
                                <span title="Click để copy tài khoản:mật khẩu" className="cursor-pointer hover:underline" onClick={() => copyToClipboard(`${p.username}:${p.password}`)}>
                                  {p.username}:{p.password}
                                </span>
                              ) : (
                                <span className="text-slate-500 italic">Không mật khẩu</span>
                              )}
                            </td>

                            {/* Rotation Mode */}
                            <td className="py-3 px-3">
                              {p.rotation_type === "sticky" && (
                                <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                                  Sticky {p.sticky_sec || 10}s
                                </span>
                              )}
                              {p.rotation_type === "request" && (
                                <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                                  Mỗi Req 1 IP
                                </span>
                              )}
                              {p.rotation_type === "static" && (
                                <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-slate-800 text-slate-300">
                                  Cố định
                                </span>
                              )}
                            </td>

                            {/* Traffic */}
                            <td className="py-3 px-3 font-mono">
                              <span className="text-slate-200">{formatBytes(p.bytes_used)}</span>
                              <span className="text-slate-500 text-[11px]">{maxText}</span>
                            </td>

                            {/* Expiry */}
                            <td className="py-3 px-3 text-slate-400">
                              {expDate}
                            </td>

                            {/* Enable/Disable Toggle */}
                            <td className="py-3 px-3 text-center">
                              <input
                                type="checkbox"
                                checked={p.enabled}
                                onChange={(e) => handleToggle(p, e.target.checked)}
                                className="w-4 h-4 rounded text-indigo-600 bg-slate-800 border-slate-700 focus:ring-indigo-500 cursor-pointer"
                              />
                            </td>

                            {/* Actions */}
                            <td className="py-3 px-3 text-right">
                              <div className="flex items-center justify-end gap-1">
                                <button
                                  onClick={() => {
                                    const host = stats?.public_ip || "103.199.11.9";
                                    const str = p.username ? `${host}:${p.port}:${p.username}:${p.password}` : `${host}:${p.port}`;
                                    copyToClipboard(str, `Đã copy: ${str}`);
                                  }}
                                  title="Copy Host:Port:User:Pass"
                                  className="p-1.5 rounded text-slate-400 hover:text-slate-100 hover:bg-slate-800 transition-colors"
                                >
                                  <Icon name="copy" className="w-3.5 h-3.5" />
                                </button>
                                <button
                                  onClick={() => setSnippetProxy(p)}
                                  title="Xem code mẫu kết nối"
                                  className="p-1.5 rounded text-slate-400 hover:text-slate-100 hover:bg-slate-800 transition-colors"
                                >
                                  <Icon name="code" className="w-3.5 h-3.5" />
                                </button>
                                <button
                                  onClick={() => setEditProxy(p)}
                                  title="Chỉnh sửa proxy"
                                  className="p-1.5 rounded text-slate-400 hover:text-slate-100 hover:bg-slate-800 transition-colors"
                                >
                                  <Icon name="edit" className="w-3.5 h-3.5" />
                                </button>
                                <button
                                  onClick={() => handleDelete(p)}
                                  title="Xóa proxy"
                                  className="p-1.5 rounded text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition-colors"
                                >
                                  <Icon name="trash" className="w-3.5 h-3.5" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        );
                      })
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* VIEW 2: BANDWIDTH & SYSTEM ANALYTICS */}
        {activeTab === "stats" && (
          <div className="flex flex-col gap-6">
            {/* KPI Cards */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4">
                <div className="text-xs text-slate-400 font-medium">Cổng đang hoạt động</div>
                <div className="text-2xl font-bold text-slate-100 font-mono mt-1">
                  {stats?.active_proxies || 0} <span className="text-xs font-normal text-slate-400">/ {stats?.total_proxies || 0} tổng</span>
                </div>
              </div>
              <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4">
                <div className="text-xs text-slate-400 font-medium">Tổng lưu lượng đã truyền</div>
                <div className="text-2xl font-bold text-indigo-400 font-mono mt-1">
                  {formatBytes(stats?.total_bytes)}
                </div>
              </div>
              <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4">
                <div className="text-xs text-slate-400 font-medium">RAM Tiêu Thụ</div>
                <div className="text-2xl font-bold text-slate-100 font-mono mt-1">
                  {stats?.ram_usage || "0 MB"}
                </div>
              </div>
              <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4">
                <div className="text-xs text-slate-400 font-medium">Trạng thái Subnet Routing</div>
                <div className="text-sm font-semibold text-emerald-400 mt-2 flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  NDP Safe (Any-IP Active)
                </div>
              </div>
            </div>

            {/* Realtime Chart */}
            <BandwidthChart totalBytes={stats?.total_bytes || 0} isPaused={refreshInterval === 0} />
          </div>
        )}

        {/* VIEW 3: INTEGRATION GUIDE */}
        {activeTab === "guide" && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-5">
              <h3 className="text-sm font-semibold text-slate-200 mb-3 flex items-center gap-2">
                <Icon name="server" className="w-4 h-4 text-indigo-400" />
                Hướng Dẫn Sử Dụng Với Antidetect Browser
              </h3>
              <p className="text-xs text-slate-400 mb-4 leading-relaxed">
                Tương thích hoàn hảo với AdsPower, GoLogin, Hidemyacc, Dolphin{'{anty}'}, MoreLogin...
              </p>
              <div className="space-y-3 text-xs text-slate-300">
                <div className="bg-slate-950/60 p-3 rounded-lg border border-slate-800/60 font-mono space-y-1">
                  <div><strong>Host / IP:</strong> {stats?.public_ip || "103.199.11.9"}</div>
                  <div><strong>Port:</strong> (Cổng đã tạo, ví dụ: 10010)</div>
                  <div><strong>Protocol:</strong> HTTP hoặc SOCKS5 (tùy cấu hình cổng)</div>
                  <div><strong>User / Password:</strong> Điền thông tin đã đặt</div>
                </div>
              </div>
            </div>

            <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-5">
              <h3 className="text-sm font-semibold text-slate-200 mb-3 flex items-center gap-2">
                <Icon name="code" className="w-4 h-4 text-indigo-400" />
                Định Dạng Chuỗi Proxy Nhanh
              </h3>
              <p className="text-xs text-slate-400 mb-4 leading-relaxed">
                Copy định dạng chuẩn để dán trực tiếp vào các tool automation:
              </p>
              <div className="bg-slate-950/60 p-3 rounded-lg border border-slate-800/60 font-mono text-xs text-indigo-300 space-y-2">
                <div><code>{stats?.public_ip || "103.199.11.9"}:PORT:USERNAME:PASSWORD</code></div>
                <div><code>http://USERNAME:PASSWORD@{stats?.public_ip || "103.199.11.9"}:PORT</code></div>
                <div><code>socks5://USERNAME:PASSWORD@{stats?.public_ip || "103.199.11.9"}:PORT</code></div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>

      {/* ------------------------------------------------------------------ */}
      {/* MODAL: CREATE PROXY                                                */}
      {/* ------------------------------------------------------------------ */}
      {createModalOpen && (
        <CreateProxyModal
          folders={allFolderNames}
          defaultFolder={selectedFolder !== "all" ? selectedFolder : "Mặc định"}
          suggestPort={stats?.suggest_port || 10010}
          onClose={() => setCreateModalOpen(false)}
          onCreateFolder={handleCreateFolder}
          onSuccess={() => {
            setCreateModalOpen(false);
            showToast("Tạo cổng proxy thành công!");
            fetchData(true);
          }}
        />
      )}

      {/* ------------------------------------------------------------------ */}
      {/* MODAL: BULK GENERATE                                               */}
      {/* ------------------------------------------------------------------ */}
      {bulkModalOpen && (
        <BulkProxyModal
          folders={allFolderNames}
          defaultFolder={selectedFolder !== "all" ? selectedFolder : "Mặc định"}
          suggestPort={stats?.suggest_port || 10010}
          onClose={() => setBulkModalOpen(false)}
          onCreateFolder={handleCreateFolder}
          onSuccess={(count) => {
            setBulkModalOpen(false);
            showToast(`Đã tạo thành công ${count} cổng proxy!`);
            fetchData(true);
          }}
        />
      )}

      {/* ------------------------------------------------------------------ */}
      {/* MODAL: EDIT PROXY                                                  */}
      {/* ------------------------------------------------------------------ */}
      {editProxy && (
        <EditProxyModal
          proxy={editProxy}
          folders={allFolderNames}
          onClose={() => setEditProxy(null)}
          onCreateFolder={handleCreateFolder}
          onSuccess={() => {
            setEditProxy(null);
            showToast("Cập nhật proxy thành công!");
            fetchData(true);
          }}
        />
      )}

      {/* ------------------------------------------------------------------ */}
      {/* MODAL: CODE SNIPPET HELPER                                         */}
      {/* ------------------------------------------------------------------ */}
      {snippetProxy && (
        <SnippetModal
          proxy={snippetProxy}
          publicIP={stats?.public_ip || "103.199.11.9"}
          onClose={() => setSnippetProxy(null)}
          onCopy={copyToClipboard}
        />
      )}

      {/* ------------------------------------------------------------------ */}
      {/* MODAL: ADD FOLDER                                                  */}
      {/* ------------------------------------------------------------------ */}
      {newFolderModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-sm w-full p-5 shadow-2xl">
            <h3 className="text-sm font-semibold text-slate-100 mb-3 flex items-center gap-2">
              <Icon name="folder-plus" className="w-4 h-4 text-indigo-400" />
              Thêm Thư Mục / Nhóm Mới
            </h3>
            <input
              type="text"
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              placeholder="VD: Nuôi Shopee, Khách Tuấn Anh..."
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 mb-4"
              autoFocus
            />
            <div className="flex justify-end gap-2">
              <button
                onClick={() => {
                  setNewFolderName("");
                  setNewFolderModalOpen(false);
                }}
                className="px-3 py-1.5 text-xs text-slate-400 hover:text-slate-200"
              >
                Hủy
              </button>
              <button
                onClick={() => {
                  if (newFolderName.trim()) {
                    handleCreateFolder(newFolderName.trim());
                    setNewFolderName("");
                    setNewFolderModalOpen(false);
                  }
                }}
                className="px-4 py-1.5 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg"
              >
                Tạo Thư Mục
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ----------------------------------------------------------------------
// Modal: Create Proxy (Single)
// ----------------------------------------------------------------------
function CreateProxyModal({ folders, defaultFolder = "Mặc định", suggestPort, onClose, onSuccess, onCreateFolder }) {
  const [port, setPort] = useState(suggestPort);
  const [name, setName] = useState("");
  const [group, setGroup] = useState(folders.includes(defaultFolder) ? defaultFolder : (folders[0] || "Mặc định"));
  const [customGroup, setCustomGroup] = useState("");
  const [proto, setProto] = useState("both");
  const [rotType, setRotType] = useState("request");
  const [stickySec, setStickySec] = useState(10);
  const [user, setUser] = useState("");
  const [pass, setPass] = useState("");
  const [maxGB, setMaxGB] = useState(0);
  const [days, setDays] = useState(30);

  const genPass = () => {
    const chars = "abcdefghjkmnpqrstuvwxyz23456789ABCDEFGHJKLMNPQRSTUVWXYZ";
    let res = "";
    for (let i = 0; i < 8; i++) res += chars[Math.floor(Math.random() * chars.length)];
    setPass(res);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    let finalGroup = group;
    if (group === "__custom__") {
      finalGroup = customGroup.trim() || "Mặc định";
      if (onCreateFolder && customGroup.trim()) {
        onCreateFolder(customGroup.trim());
      }
    }

    const payload = {
      port: Number(port),
      name: name.trim() || `Proxy :${port}`,
      group: finalGroup,
      proto,
      rotation_type: rotType,
      sticky_sec: Number(stickySec),
      username: user.trim(),
      password: pass.trim(),
      max_gb: Number(maxGB),
      expire_days: Number(days),
    };

    try {
      const res = await fetch("/api/proxies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (res.ok) {
        onSuccess();
      } else {
        alert(data.error || "Có lỗi xảy ra khi tạo proxy");
      }
    } catch {
      alert("Lỗi kết nối tới máy chủ");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-lg w-full p-6 shadow-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between pb-3 border-b border-slate-800">
          <h3 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
            <Icon name="plus" className="w-4 h-4 text-indigo-400" />
            Thêm Cổng Proxy Mới
          </h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-200">
            <Icon name="x" className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4 text-xs">
          {/* Note / Customer Name */}
          <div>
            <label className="block text-slate-300 font-medium mb-1">Ghi Chú / Tên Khách Hàng (Note)</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="VD: Nuôi nick FB #1, Khách Tuấn Anh..."
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500"
              autoFocus
            />
          </div>

          {/* Folder / Grouping */}
          <div>
            <label className="block text-slate-300 font-medium mb-1">Thư Mục / Nhóm Phân Loại</label>
            <div className="flex gap-2">
              <select
                value={group}
                onChange={(e) => {
                  setGroup(e.target.value);
                  setCustomGroup("");
                }}
                className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500 flex-1 cursor-pointer"
              >
                {folders.map((f) => (
                  <option key={f} value={f}>{f}</option>
                ))}
                <option value="__custom__">+ Nhập nhóm mới...</option>
              </select>
              {group === "__custom__" && (
                <input
                  type="text"
                  value={customGroup}
                  onChange={(e) => setCustomGroup(e.target.value)}
                  placeholder="Tên nhóm mới..."
                  className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 flex-1 focus:outline-none focus:border-indigo-500"
                />
              )}
            </div>
          </div>

          {/* Port & Protocol */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Cổng (Port) *</label>
              <input
                type="number"
                value={port}
                onChange={(e) => setPort(e.target.value)}
                min="1024"
                max="65535"
                required
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Giao Thức</label>
              <select
                value={proto}
                onChange={(e) => setProto(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
              >
                <option value="both">HTTP & SOCKS5 (Đa năng)</option>
                <option value="http">Chỉ HTTP / HTTPS</option>
                <option value="socks5">Chỉ SOCKS5</option>
              </select>
            </div>
          </div>

          {/* Rotation Mode */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Chế Độ Xoay IP</label>
              <select
                value={rotType}
                onChange={(e) => setRotType(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
              >
                <option value="request">Mỗi Request 1 IP Mới</option>
                <option value="sticky">Sticky Theo Thời Gian</option>
                <option value="static">IP Tĩnh Cố Định</option>
              </select>
            </div>
            {rotType === "sticky" && (
              <div>
                <label className="block text-slate-300 font-medium mb-1">Giữ IP (Số giây)</label>
                <input
                  type="number"
                  value={stickySec}
                  onChange={(e) => setStickySec(e.target.value)}
                  min="1"
                  max="86400"
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                />
              </div>
            )}
          </div>

          {/* Credentials */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Username (Để trống = No Auth)</label>
              <input
                type="text"
                value={user}
                onChange={(e) => setUser(e.target.value)}
                placeholder="user..."
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Password</label>
              <div className="flex gap-1.5">
                <input
                  type="text"
                  value={pass}
                  onChange={(e) => setPass(e.target.value)}
                  placeholder="pass..."
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
                />
                <button
                  type="button"
                  onClick={genPass}
                  title="Random mật khẩu"
                  className="px-2.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs"
                >
                  Random
                </button>
              </div>
            </div>
          </div>

          {/* Limits */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Giới Hạn Dung Lượng (GB)</label>
              <input
                type="number"
                value={maxGB}
                onChange={(e) => setMaxGB(e.target.value)}
                min="0"
                step="0.5"
                placeholder="0 = Không giới hạn"
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Hạn Sử Dụng (Số ngày)</label>
              <input
                type="number"
                value={days}
                onChange={(e) => setDays(e.target.value)}
                min="0"
                placeholder="0 = Vĩnh viễn"
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          {/* Buttons */}
          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200"
            >
              Hủy
            </button>
            <button
              type="submit"
              className="px-5 py-2 rounded-lg font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-sm"
            >
              Tạo Cổng Proxy
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ----------------------------------------------------------------------
// Modal: Bulk Create Proxies
// ----------------------------------------------------------------------
function BulkProxyModal({ folders, defaultFolder = "Mặc định", suggestPort, onClose, onSuccess, onCreateFolder }) {
  const [startPort, setStartPort] = useState(suggestPort);
  const [count, setCount] = useState(10);
  const [namePrefix, setNamePrefix] = useState("Khách");
  const [group, setGroup] = useState(folders.includes(defaultFolder) ? defaultFolder : (folders[0] || "Mặc định"));
  const [customGroup, setCustomGroup] = useState("");
  const [proto, setProto] = useState("both");
  const [rotType, setRotType] = useState("request");
  const [stickySec, setStickySec] = useState(10);
  const [autoUserPass, setAutoUserPass] = useState(true);
  const [user, setUser] = useState("proxy");
  const [pass, setPass] = useState("");
  const [maxGB, setMaxGB] = useState(0);
  const [days, setDays] = useState(30);

  const handleSubmit = async (e) => {
    e.preventDefault();
    let finalGroup = group;
    if (group === "__custom__") {
      finalGroup = customGroup.trim() || "Mặc định";
      if (onCreateFolder && customGroup.trim()) {
        onCreateFolder(customGroup.trim());
      }
    }

    const payload = {
      start_port: Number(startPort),
      count: Number(count),
      name_prefix: namePrefix.trim() || "Proxy",
      group: finalGroup,
      proto,
      rotation_type: rotType,
      sticky_sec: Number(stickySec),
      auto_user_pass: autoUserPass,
      username: autoUserPass ? "" : user.trim(),
      password: autoUserPass ? "" : pass.trim(),
      max_gb: Number(maxGB),
      expire_days: Number(days),
    };

    try {
      const res = await fetch("/api/proxies/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (res.ok) {
        onSuccess(data.count || count);
      } else {
        alert(data.error || "Có lỗi xảy ra khi tạo hàng loạt");
      }
    } catch {
      alert("Lỗi kết nối máy chủ");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-lg w-full p-6 shadow-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between pb-3 border-b border-slate-800">
          <h3 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
            <Icon name="layers" className="w-4 h-4 text-indigo-400" />
            Tạo Cổng Proxy Hàng Loạt
          </h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-200">
            <Icon name="x" className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4 text-xs">
          {/* Port and Count */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Cổng bắt đầu *</label>
              <input
                type="number"
                value={startPort}
                onChange={(e) => setStartPort(e.target.value)}
                min="1024"
                max="65500"
                required
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Số lượng cổng *</label>
              <input
                type="number"
                value={count}
                onChange={(e) => setCount(e.target.value)}
                min="1"
                max="500"
                required
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          {/* Name Prefix and Group */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Tiền tố tên ghi chú</label>
              <input
                type="text"
                value={namePrefix}
                onChange={(e) => setNamePrefix(e.target.value)}
                placeholder="VD: Nuôi Shopee, Khách VIP..."
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Thư mục / Nhóm</label>
              <div className="flex gap-2">
                <select
                  value={group}
                  onChange={(e) => {
                    setGroup(e.target.value);
                    if (e.target.value !== "__custom__") setCustomGroup("");
                  }}
                  className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500 flex-1 cursor-pointer"
                >
                  {folders.map((f) => (
                    <option key={f} value={f}>{f}</option>
                  ))}
                  <option value="__custom__">+ Nhập nhóm mới...</option>
                </select>
                {group === "__custom__" && (
                  <input
                    type="text"
                    value={customGroup}
                    onChange={(e) => setCustomGroup(e.target.value)}
                    placeholder="Tên nhóm mới..."
                    className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 flex-1 focus:outline-none focus:border-indigo-500"
                  />
                )}
              </div>
            </div>
          </div>

          {/* Protocol & Rotation */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Giao Thức</label>
              <select
                value={proto}
                onChange={(e) => setProto(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
              >
                <option value="both">HTTP & SOCKS5 (Đa năng)</option>
                <option value="http">Chỉ HTTP(S)</option>
                <option value="socks5">Chỉ SOCKS5</option>
              </select>
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Chế Độ Xoay IP</label>
              <select
                value={rotType}
                onChange={(e) => setRotType(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
              >
                <option value="request">Mỗi Request 1 IP Mới</option>
                <option value="sticky">Sticky Theo Thời Gian</option>
                <option value="static">Cố định</option>
              </select>
            </div>
          </div>

          {/* Auth options */}
          <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800 space-y-2">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={autoUserPass}
                onChange={(e) => setAutoUserPass(e.target.checked)}
                className="rounded text-indigo-600 bg-slate-900 border-slate-700"
              />
              <span className="text-slate-300">Tự động sinh User/Pass ngẫu nhiên riêng cho từng cổng</span>
            </label>
            {!autoUserPass && (
              <div className="grid grid-cols-2 gap-2 pt-2">
                <input
                  type="text"
                  value={user}
                  onChange={(e) => setUser(e.target.value)}
                  placeholder="Username chung"
                  className="bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200"
                />
                <input
                  type="text"
                  value={pass}
                  onChange={(e) => setPass(e.target.value)}
                  placeholder="Password chung"
                  className="bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200"
                />
              </div>
            )}
          </div>

          {/* Buttons */}
          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200"
            >
              Hủy
            </button>
            <button
              type="submit"
              className="px-5 py-2 rounded-lg font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-sm"
            >
              Tiến Hành Tạo {count} Cổng
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ----------------------------------------------------------------------
// Modal: Edit Proxy
// ----------------------------------------------------------------------
function EditProxyModal({ proxy, folders, onClose, onSuccess, onCreateFolder }) {
  const [name, setName] = useState(proxy.name || "");
  const [group, setGroup] = useState(proxy.group || "Mặc định");
  const [customGroup, setCustomGroup] = useState("");
  const [proto, setProto] = useState(proxy.proto || "both");
  const [rotType, setRotType] = useState(proxy.rotation_type || "request");
  const [stickySec, setStickySec] = useState(proxy.sticky_sec || 10);
  const [user, setUser] = useState(proxy.username || "");
  const [pass, setPass] = useState(proxy.password || "");
  const [maxGB, setMaxGB] = useState(proxy.max_bytes ? (proxy.max_bytes / 1024 / 1024 / 1024).toFixed(1) : 0);

  const handleSubmit = async (e) => {
    e.preventDefault();
    let finalGroup = group;
    if (group === "__custom__") {
      finalGroup = customGroup.trim() || "Mặc định";
      if (onCreateFolder && customGroup.trim()) {
        onCreateFolder(customGroup.trim());
      }
    }

    const payload = {
      name: name.trim(),
      group: finalGroup,
      proto,
      rotation_type: rotType,
      sticky_sec: Number(stickySec),
      username: user.trim(),
      password: pass.trim(),
      max_gb: Number(maxGB),
    };

    try {
      const res = await fetch(`/api/proxies/${proxy.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        onSuccess();
      } else {
        alert("Lỗi khi cập nhật proxy");
      }
    } catch {
      alert("Lỗi kết nối máy chủ");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-lg w-full p-6 shadow-2xl">
        <div className="flex items-center justify-between pb-3 border-b border-slate-800">
          <h3 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
            <Icon name="edit" className="w-4 h-4 text-indigo-400" />
            Chỉnh Sửa Proxy :{proxy.port}
          </h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-200">
            <Icon name="x" className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4 text-xs">
          <div>
            <label className="block text-slate-300 font-medium mb-1">Ghi Chú / Tên Khách Hàng (Note)</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-slate-300 font-medium mb-1">Thư Mục / Nhóm</label>
            <div className="flex gap-2">
              <select
                value={group}
                onChange={(e) => {
                  setGroup(e.target.value);
                  setCustomGroup("");
                }}
                className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 flex-1"
              >
                {folders.map((f) => (
                  <option key={f} value={f}>{f}</option>
                ))}
                <option value="__custom__">+ Đổi sang nhóm mới...</option>
              </select>
              {group === "__custom__" && (
                <input
                  type="text"
                  value={customGroup}
                  onChange={(e) => setCustomGroup(e.target.value)}
                  placeholder="Tên nhóm mới..."
                  className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 flex-1"
                />
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Giao Thức</label>
              <select
                value={proto}
                onChange={(e) => setProto(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200"
              >
                <option value="both">HTTP & SOCKS5 (Đa năng)</option>
                <option value="http">Chỉ HTTP / HTTPS</option>
                <option value="socks5">Chỉ SOCKS5</option>
              </select>
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Chế Độ Xoay IP</label>
              <select
                value={rotType}
                onChange={(e) => setRotType(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200"
              >
                <option value="request">Mỗi Request 1 IP</option>
                <option value="sticky">Sticky Theo Thời Gian</option>
                <option value="static">Cố định</option>
              </select>
            </div>
          </div>

          {rotType === "sticky" && (
            <div>
              <label className="block text-slate-300 font-medium mb-1">Thời gian giữ IP (Giây)</label>
              <input
                type="number"
                value={stickySec}
                onChange={(e) => setStickySec(e.target.value)}
                min="1"
                max="86400"
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200"
              />
            </div>
          )}

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Username</label>
              <input
                type="text"
                value={user}
                onChange={(e) => setUser(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">Password</label>
              <input
                type="text"
                value={pass}
                onChange={(e) => setPass(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-200"
              />
            </div>
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200"
            >
              Hủy
            </button>
            <button
              type="submit"
              className="px-5 py-2 rounded-lg font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-sm"
            >
              Lưu Thay Đổi
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ----------------------------------------------------------------------
// Modal: Connection Snippet Helper
// ----------------------------------------------------------------------
function SnippetModal({ proxy, publicIP, onClose, onCopy }) {
  const p = proxy;
  const host = publicIP;
  const authPart = p.username ? `${p.username}:${p.password}@` : "";
  const plainString = p.username ? `${host}:${p.port}:${p.username}:${p.password}` : `${host}:${p.port}`;
  const httpUrl = `http://${authPart}${host}:${p.port}`;
  const socks5Url = `socks5://${authPart}${host}:${p.port}`;

  const pythonSnippet = `import requests

proxies = {
    "http": "${httpUrl}",
    "https": "${httpUrl}"
}

r = requests.get("https://api64.ipify.org?format=json", proxies=proxies, timeout=10)
print("Public IPv6:", r.json()["ip"])`;

  const curlSnippet = `curl -x ${httpUrl} https://api64.ipify.org`;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-xl w-full p-6 shadow-2xl text-xs">
        <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-4">
          <h3 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
            <Icon name="code" className="w-4 h-4 text-indigo-400" />
            Thông Tin Kết Nối Cổng :{p.port} ({p.name || "Không tên"})
          </h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-200">
            <Icon name="x" className="w-4 h-4" />
          </button>
        </div>

        <div className="space-y-4">
          {/* Antidetect String */}
          <div>
            <div className="flex justify-between items-center mb-1 text-slate-300">
              <span className="font-semibold">Định dạng Antidetect (Host:Port:User:Pass):</span>
              <button
                onClick={() => onCopy(plainString, "Đã copy chuỗi Antidetect!")}
                className="text-indigo-400 hover:underline flex items-center gap-1"
              >
                <Icon name="copy" className="w-3 h-3" /> Copy
              </button>
            </div>
            <div className="bg-slate-950 p-2.5 rounded-lg border border-slate-800 font-mono text-slate-200 select-all">
              {plainString}
            </div>
          </div>

          {/* cURL */}
          <div>
            <div className="flex justify-between items-center mb-1 text-slate-300">
              <span className="font-semibold">Lệnh cURL:</span>
              <button
                onClick={() => onCopy(curlSnippet, "Đã copy lệnh cURL!")}
                className="text-indigo-400 hover:underline flex items-center gap-1"
              >
                <Icon name="copy" className="w-3 h-3" /> Copy
              </button>
            </div>
            <pre className="bg-slate-950 p-2.5 rounded-lg border border-slate-800 font-mono text-slate-300 overflow-x-auto">
              {curlSnippet}
            </pre>
          </div>

          {/* Python */}
          <div>
            <div className="flex justify-between items-center mb-1 text-slate-300">
              <span className="font-semibold">Code mẫu Python:</span>
              <button
                onClick={() => onCopy(pythonSnippet, "Đã copy code Python!")}
                className="text-indigo-400 hover:underline flex items-center gap-1"
              >
                <Icon name="copy" className="w-3 h-3" /> Copy
              </button>
            </div>
            <pre className="bg-slate-950 p-2.5 rounded-lg border border-slate-800 font-mono text-slate-300 overflow-x-auto text-[11px] leading-relaxed">
              {pythonSnippet}
            </pre>
          </div>
        </div>

        <div className="flex justify-end pt-4 mt-4 border-t border-slate-800">
          <button
            onClick={onClose}
            className="px-4 py-1.5 rounded-lg text-slate-300 bg-slate-800 hover:bg-slate-700"
          >
            Đóng
          </button>
        </div>
      </div>
    </div>
  );
}

// ----------------------------------------------------------------------
// Mount React App
// ----------------------------------------------------------------------
const rootElement = document.getElementById("root");
if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(<App />);
}
