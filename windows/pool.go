package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// RollingPool dynamically manages a bounded pool of IPv6 addresses on a network interface.
// It keeps the OS routing table small (e.g. 200 IPs) while continuously rotating through
// thousands of addresses over time without freezing or lagging the OS.
type RollingPool struct {
	ifaceName      string
	ipNet          *net.IPNet
	poolSize       int           // Max IPs maintained concurrently on interface (e.g. 200)
	rotateBatch    int           // Number of IPs swapped out each interval (e.g. 30)
	rotateInterval time.Duration // Interval between rotations (e.g. 30s)
	storeType      string        // "active" (RAM-only)
	skipAsSource   bool          // true: prevents Windows from using pool IPs for ordinary apps
	prefixLen      int           // 64

	mu           sync.RWMutex
	activeIPs    []net.IP
	retiringIPs  []net.IP
	counter      uint64
	totalRotated uint64

	stopChan chan struct{}
	stopped  bool
	wg       sync.WaitGroup
}

// NewRollingPool creates a new dynamic rolling pool configuration
func NewRollingPool(ifaceName string, ipNet *net.IPNet, poolSize int, rotateBatch int, rotateInterval time.Duration) *RollingPool {
	if poolSize <= 0 {
		poolSize = 200
	}
	if rotateBatch <= 0 {
		rotateBatch = 30
	}
	if rotateBatch > poolSize {
		rotateBatch = poolSize
	}
	if rotateInterval <= 0 {
		rotateInterval = 30 * time.Second
	}

	return &RollingPool{
		ifaceName:      ifaceName,
		ipNet:          ipNet,
		poolSize:       poolSize,
		rotateBatch:    rotateBatch,
		rotateInterval: rotateInterval,
		storeType:      "active",
		skipAsSource:   true,
		prefixLen:      64,
		stopChan:       make(chan struct{}),
	}
}

// Start fills the initial pool and launches the background rolling routine
func (p *RollingPool) Start() error {
	fmt.Printf("\n⚡ Đang nạp Dynamic Pool ban đầu (%d IPs) vào card '%s'...\n", p.poolSize, p.ifaceName)
	ips, err := GenerateIPv6List(p.ipNet, p.poolSize, ModeRandom)
	if err != nil {
		return fmt.Errorf("lỗi sinh IPv6 ban đầu: %w", err)
	}

	success, errs := BatchAddIPv6(p.ifaceName, ips, p.prefixLen, p.storeType, p.skipAsSource, nil)
	if success == 0 && len(errs) > 0 {
		return fmt.Errorf("không thể gán IP vào card mạng: %v", errs[0])
	}

	p.mu.Lock()
	p.activeIPs = ips[:success]
	p.totalRotated = uint64(success)
	p.mu.Unlock()

	fmt.Printf("✅ Đã gán thành công %d IPv6 vào card mạng (Chế độ RAM 'active')!\n", success)
	fmt.Printf("🔄 Cơ chế Cuốn Chiếu: Mỗi %v sẽ tự động xoay %d IP mới (Giữ card mạng luôn nhẹ ~%d IP)\n\n",
		p.rotateInterval, p.rotateBatch, p.poolSize)

	p.wg.Add(1)
	go p.loop()

	return nil
}

// loop runs the periodic rotation ticker
func (p *RollingPool) loop() {
	defer p.wg.Done()
	ticker := time.NewTicker(p.rotateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.rotate()
		}
	}
}

// rotate performs a graceful retirement and adds a fresh batch of IPs
func (p *RollingPool) rotate() {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return
	}

	// 1. Physically delete IPs whose grace period has expired
	toDeleteList := make([]string, len(p.retiringIPs))
	for i, ip := range p.retiringIPs {
		toDeleteList[i] = ip.String()
	}
	p.retiringIPs = nil
	p.mu.Unlock()

	if len(toDeleteList) > 0 {
		BatchDeleteIPv6(p.ifaceName, toDeleteList, p.storeType, nil)
	}

	// 2. Generate a new batch of IPs
	newIPs, err := GenerateIPv6List(p.ipNet, p.rotateBatch, ModeRandom)
	if err != nil {
		log.Printf("[⚠️ Dynamic Pool] Lỗi sinh IP mới: %v", err)
		return
	}

	added, _ := BatchAddIPv6(p.ifaceName, newIPs, p.prefixLen, p.storeType, p.skipAsSource, nil)
	if added == 0 {
		log.Printf("[⚠️ Dynamic Pool] Không gán được IP mới vào card mạng")
		return
	}

	validNewIPs := newIPs[:added]

	// 3. Move oldest batch to retiring list (giving active connections grace time to complete)
	p.mu.Lock()
	defer p.mu.Unlock()

	retireCount := added
	if retireCount > len(p.activeIPs) {
		retireCount = len(p.activeIPs)
	}

	p.retiringIPs = append(p.retiringIPs, p.activeIPs[:retireCount]...)
	p.activeIPs = append(p.activeIPs[retireCount:], validNewIPs...)
	p.totalRotated += uint64(added)

	log.Printf("[🔄 Dynamic Pool] Đã xoay +%d IP mới | Đang duy trì: %d IPs trên card | Tổng IP đã xoay: %d",
		added, len(p.activeIPs), p.totalRotated)
}

// PickIPv6 picks an IPv6 using round-robin from current active pool
func (p *RollingPool) PickIPv6() net.IP {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.activeIPs) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&p.counter, 1) % uint64(len(p.activeIPs))
	return p.activeIPs[idx]
}

// RandomIPv6 picks a random IPv6 from current active pool
func (p *RollingPool) RandomIPv6() net.IP {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.activeIPs) == 0 {
		return nil
	}
	return p.activeIPs[rand.Intn(len(p.activeIPs))]
}

// Stop terminates the rolling worker and removes all pool IPs from the interface
func (p *RollingPool) Stop() {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return
	}
	p.stopped = true
	close(p.stopChan)

	var allToDelete []string
	for _, ip := range p.activeIPs {
		allToDelete = append(allToDelete, ip.String())
	}
	for _, ip := range p.retiringIPs {
		allToDelete = append(allToDelete, ip.String())
	}
	p.activeIPs = nil
	p.retiringIPs = nil
	p.mu.Unlock()

	p.wg.Wait()

	if len(allToDelete) > 0 {
		fmt.Printf("\n🧹 Đang gỡ bỏ toàn bộ %d IP của Dynamic Pool khỏi card '%s'...\n", len(allToDelete), p.ifaceName)
		BatchDeleteIPv6(p.ifaceName, allToDelete, p.storeType, nil)
		fmt.Printf("✅ Đã dọn dẹp sạch card mạng!\n")
	}
}
