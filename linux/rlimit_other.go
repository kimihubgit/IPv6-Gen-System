//go:build !linux

package main

func raiseFDLimit() {
	// No-op on non-Linux platforms
}
