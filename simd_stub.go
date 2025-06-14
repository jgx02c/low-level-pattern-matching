//go:build !cgo
// +build !cgo

package main

import (
	"fmt"
	"os"
)

// mainSIMD stub for Pure Go builds (SIMD requires CGO)
func mainSIMD() {
	fmt.Println("❌ SIMD mode not available in Pure Go build")
	fmt.Println("💡 SIMD requires CGO. Build with: make legal-nlp-simd")
	fmt.Println("🔵 Using Pure Go mode instead...")
	os.Exit(1)
}
