//go:build linux

package main

import (
	"log"
	"syscall"
)

func raiseFDLimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err == nil {
		rLimit.Cur = 65535
		rLimit.Max = 65535
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			log.Printf("⚠️ Không thể nâng ulimit -n lên 65535: %v", err)
		} else {
			log.Printf("🚀 Đã tự động nâng giới hạn File Descriptors (ulimit -n) lên %d sockets", rLimit.Cur)
		}
	}
}
