# Ultra-Low Latency Legal NLP Pipeline
# Go + C + Assembly/SIMD Build System

CC = gcc
CFLAGS = -mavx512f -O3 -march=native -fPIC -Wall
ASM = nasm
ASMFLAGS = -f elf64

# Targets
BINARY = legal-nlp-simd
LIB = libmatcher.so

# Source files
C_SOURCES = matcher.c
ASM_SOURCES = simd_match.s
GO_SOURCES = main.go cache.go types.go

# Object files
C_OBJECTS = $(C_SOURCES:.c=.o)
ASM_OBJECTS = $(ASM_SOURCES:.s=.o)

.PHONY: all clean test benchmark

all: $(BINARY)

# Build shared library from C and Assembly
$(LIB): $(C_OBJECTS) $(ASM_OBJECTS)
	$(CC) -shared -o $@ $^ $(CFLAGS)

# Compile C source
%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

# Assemble ASM source  
%.o: %.s
	$(ASM) $(ASMFLAGS) $< -o $@

# Build Go binary with CGO
$(BINARY): $(LIB) $(GO_SOURCES)
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BINARY) .

# Generate legal patterns
patterns:
	@echo "🏛️  Generating legal hearsay patterns..."
	go run patterns/generate.go

# Run performance tests
test: $(BINARY)
	@echo "🚀 Running performance tests..."
	./$(BINARY) --test

# Benchmark assembly performance
benchmark: $(BINARY)
	@echo "⚡ Benchmarking SIMD performance..."
	./$(BINARY) --benchmark

# Check CPU features
check-cpu:
	@echo "🔍 Checking CPU SIMD support..."
	@lscpu | grep -E "(avx|sse)" || echo "❌ No advanced SIMD support detected"

clean:
	rm -f *.o $(LIB) $(BINARY)
	
install-deps:
	@echo "📦 Installing dependencies..."
	@command -v nasm >/dev/null 2>&1 || { echo "Installing nasm..."; brew install nasm 2>/dev/null || sudo apt-get install nasm -y; }

setup: install-deps check-cpu patterns
	@echo "✅ Setup complete! Run 'make' to build." 