# Legal NLP Pipeline - Ultra-Fast Hearsay Detection
# Makefile for building both Pure Go and SIMD-accelerated versions

# Build configuration
GO_VERSION := $(shell go version | cut -d' ' -f3)
ARCH := $(shell uname -m)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

# Optimization flags
CFLAGS_BASE := -O3 -std=c11 -Wall -Wextra
CFLAGS_X86 := -mavx512f -mavx2 -msse4.2 -march=native
CFLAGS_ARM := -mcpu=native -mtune=native

# Detect architecture and set appropriate flags
ifeq ($(ARCH),x86_64)
    CFLAGS_ARCH := $(CFLAGS_X86)
    SIMD_SUPPORT := AVX-512/AVX2
else ifeq ($(ARCH),arm64)
    CFLAGS_ARCH := $(CFLAGS_ARM)
    SIMD_SUPPORT := NEON
else
    CFLAGS_ARCH := 
    SIMD_SUPPORT := Scalar
endif

# Go build flags
GO_BUILD_FLAGS := -ldflags="-s -w" -trimpath

.PHONY: all clean test benchmark simd-test simd-benchmark help

# Default target
all: legal-nlp legal-nlp-simd

# Pure Go version (no C dependencies)
legal-nlp:
	@echo "🏗️  Building Pure Go Legal NLP Pipeline..."
	@echo "   Go Version: $(GO_VERSION)"
	@echo "   Architecture: $(ARCH)"
	@echo "   OS: $(OS)"
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o legal-nlp main.go cache.go simd_stub.go
	@echo "✅ Pure Go build complete: legal-nlp"

# SIMD-accelerated version (with C core)
legal-nlp-simd:
	@echo "🚀 Building SIMD-accelerated Legal NLP Pipeline..."
	@echo "   Go Version: $(GO_VERSION)"
	@echo "   Architecture: $(ARCH)"
	@echo "   SIMD Support: $(SIMD_SUPPORT)"
	@echo "   C Flags: $(CFLAGS_BASE) $(CFLAGS_ARCH)"
	CGO_ENABLED=1 CGO_CFLAGS="$(CFLAGS_BASE) $(CFLAGS_ARCH)" \
	go build $(GO_BUILD_FLAGS) -o legal-nlp-simd main.go simd_main.go cache.go
	@echo "✅ SIMD build complete: legal-nlp-simd"

# Test targets
test: legal-nlp
	@echo "🧪 Running Pure Go tests..."
	./legal-nlp --test

simd-test: legal-nlp-simd
	@echo "🧪 Running SIMD tests..."
	./legal-nlp-simd --simd --test

# Benchmark targets
benchmark: legal-nlp
	@echo "🏃 Running Pure Go benchmark..."
	./legal-nlp --benchmark

simd-benchmark: legal-nlp-simd
	@echo "🏃 Running SIMD benchmark..."
	./legal-nlp-simd --simd --benchmark

# Performance comparison
compare: legal-nlp legal-nlp-simd
	@echo "⚡ Performance Comparison: Pure Go vs SIMD"
	@echo ""
	@echo "🔵 Pure Go DFA Performance:"
	@./legal-nlp --benchmark | grep -E "(Avg Time|Searches/Second)"
	@echo ""
	@echo "🚀 SIMD Accelerated Performance:"
	@./legal-nlp-simd --simd --benchmark | grep -E "(Avg Time|Searches/Second)"

# Generate large pattern files for testing
patterns-1m:
	@echo "📝 Generating 1M patterns for testing..."
	go run generate_patterns.go

# Interactive demos
demo: legal-nlp
	@echo "💬 Starting Pure Go interactive demo..."
	./legal-nlp

simd-demo: legal-nlp-simd
	@echo "💬 Starting SIMD interactive demo..."
	./legal-nlp-simd --simd

# Development targets
dev-build:
	@echo "🔧 Development build (with debug info)..."
	go build -race -o legal-nlp-dev main.go cache.go

# Check CPU features
cpu-info:
	@echo "🖥️  CPU Feature Detection:"
	@echo "   Architecture: $(ARCH)"
	@echo "   SIMD Support: $(SIMD_SUPPORT)"
	@if [ "$(ARCH)" = "x86_64" ]; then \
		echo "   Checking x86_64 features..."; \
		grep -E "(avx|sse)" /proc/cpuinfo | head -5 || echo "   /proc/cpuinfo not available"; \
	elif [ "$(ARCH)" = "arm64" ]; then \
		echo "   ARM64 NEON support: Available"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f legal-nlp legal-nlp-simd legal-nlp-dev
	rm -f patterns_1000000.txt
	@echo "✅ Clean complete"

# Install dependencies
deps:
	@echo "📦 Installing Go dependencies..."
	go mod tidy
	@echo "✅ Dependencies installed"

# Help target
help:
	@echo "🏛️  Legal NLP Pipeline - Build System"
	@echo ""
	@echo "📋 Available targets:"
	@echo "   all              - Build both Pure Go and SIMD versions"
	@echo "   legal-nlp        - Build Pure Go version only"
	@echo "   legal-nlp-simd   - Build SIMD-accelerated version"
	@echo ""
	@echo "🧪 Testing:"
	@echo "   test             - Run Pure Go tests"
	@echo "   simd-test        - Run SIMD tests"
	@echo "   benchmark        - Run Pure Go benchmark"
	@echo "   simd-benchmark   - Run SIMD benchmark"
	@echo "   compare          - Compare Pure Go vs SIMD performance"
	@echo ""
	@echo "💬 Interactive:"
	@echo "   demo             - Start Pure Go interactive demo"
	@echo "   simd-demo        - Start SIMD interactive demo"
	@echo ""
	@echo "🔧 Development:"
	@echo "   dev-build        - Development build with race detection"
	@echo "   patterns-1m      - Generate 1M test patterns"
	@echo "   cpu-info         - Show CPU feature information"
	@echo "   deps             - Install Go dependencies"
	@echo "   clean            - Clean build artifacts"
	@echo "   help             - Show this help"
	@echo ""
	@echo "🚀 Quick start:"
	@echo "   make all && make compare"

# Version info
version:
	@echo "🏛️  Legal NLP Pipeline v2.0"
	@echo "   Pure Go: Aho-Corasick DFA"
	@echo "   SIMD: AVX-512/NEON accelerated"
	@echo "   Build system: $(OS)/$(ARCH)"
	@echo "   Go: $(GO_VERSION)" 