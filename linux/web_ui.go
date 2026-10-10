package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed web/*
var embeddedWebFS embed.FS

// getFileSystem returns an http.FileSystem. If the local "web" directory exists on disk,
// it serves live files directly from disk (so users can edit HTML/CSS directly without recompiling).
// Otherwise it falls back to the embedded files in the binary.
func getFileSystem() http.FileSystem {
	if fi, err := os.Stat("web"); err == nil && fi.IsDir() {
		log.Println("📁 Đang phục vụ theme/giao diện trực tiếp từ thư mục đĩa: ./web")
		return http.Dir("web")
	}
	sub, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		log.Printf("⚠️ Lỗi nạp embedded web filesystem: %v", err)
		return http.Dir("web")
	}
	return http.FS(sub)
}
