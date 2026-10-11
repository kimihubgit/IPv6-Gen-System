package store

import (
	"sync"
	"sync/atomic"
	"time"
)

// RotationPolicy defines how IPv6 addresses are rotated
type RotationPolicy string

const (
	RotationPerRequest RotationPolicy = "request" // Mỗi request 1 IPv6 mới
	RotationSticky     RotationPolicy = "sticky"  // Giữ nguyên IPv6 trong X giây (5s/10s/30s...)
	RotationStatic     RotationPolicy = "static"  // Cố định 1 IPv6 duy nhất
)

// ProxyAccount represents a proxy allocated on a dedicated port
type ProxyAccount struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`          // Tên ghi chú / Khách hàng
	Group        string         `json:"group"`         // Nhóm / Thư mục phân loại
	Port         int            `json:"port"`          // Cổng kết nối (1024 - 65535)
	Username     string         `json:"username"`      // Tài khoản auth
	Password     string         `json:"password"`      // Mật khẩu auth
	Proto        string         `json:"proto"`         // "both" (HTTP+SOCKS5 auto-detect)
	MaxBytes     int64          `json:"max_bytes"`     // Hạn mức dung lượng (0 = không giới hạn)
	BytesUsed    int64          `json:"bytes_used"`    // Tổng lưu lượng đã dùng (bytes) = BytesUp + BytesDown
	BytesUp      int64          `json:"bytes_up"`      // Lưu lượng Upload (tải lên)
	BytesDown    int64          `json:"bytes_down"`    // Lưu lượng Download (tải về)
	ExpiresAt    *time.Time     `json:"expires_at"`    // Hạn dùng (nil = vĩnh viễn)
	RotationType RotationPolicy `json:"rotation_type"` // "request" | "sticky" | "static"
	StickySec    int            `json:"sticky_sec"`    // Số giây giữ IP
	StaticIP     string         `json:"static_ip"`     // IPv6 tĩnh nếu chọn static
	Enabled      bool           `json:"enabled"`       // Trạng thái bật/tắt
	CreatedAt    time.Time      `json:"created_at"`
	LastActive   time.Time      `json:"last_active"`

	mu sync.RWMutex `json:"-"`
}

// IsExpired checks if the account has expired
func (p *ProxyAccount) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}

// IsTrafficExceeded checks if quota has been exceeded
func (p *ProxyAccount) IsTrafficExceeded() bool {
	if p.MaxBytes <= 0 {
		return false
	}
	return atomic.LoadInt64(&p.BytesUsed) >= p.MaxBytes
}

// CanAccess returns whether proxy can accept connections
func (p *ProxyAccount) CanAccess() (bool, string) {
	if !p.Enabled {
		return false, "Proxy đã bị tạm dừng"
	}
	if p.IsExpired() {
		return false, "Proxy đã hết hạn sử dụng"
	}
	if p.IsTrafficExceeded() {
		return false, "Proxy đã dùng hết dung lượng (GB)"
	}
	return true, ""
}

// AddUp atomically increments uploaded bandwidth and total bandwidth
func (p *ProxyAccount) AddUp(bytes int64) {
	if bytes > 0 {
		atomic.AddInt64(&p.BytesUp, bytes)
		atomic.AddInt64(&p.BytesUsed, bytes)
	}
}

// AddDown atomically increments downloaded bandwidth and total bandwidth
func (p *ProxyAccount) AddDown(bytes int64) {
	if bytes > 0 {
		atomic.AddInt64(&p.BytesDown, bytes)
		atomic.AddInt64(&p.BytesUsed, bytes)
	}
}

// AddTraffic atomically increments used bandwidth
func (p *ProxyAccount) AddTraffic(bytes int64) {
	if bytes > 0 {
		atomic.AddInt64(&p.BytesUsed, bytes)
	}
}
