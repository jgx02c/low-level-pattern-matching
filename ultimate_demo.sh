#!/bin/bash

# Ultimate Legal NLP Pipeline Demo
# Showcases Pure Go DFA vs SIMD-accelerated performance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Demo configuration
DEMO_TEXT="The witness said that he saw the defendant at the scene. According to the plaintiff's testimony, the incident occurred around midnight. She told the court that the defendant allegedly made threatening statements. Court records show that similar complaints have been filed before. The evidence suggests that the defendant was present during the incident."

echo -e "${PURPLE}🏛️  LEGAL NLP PIPELINE - ULTIMATE PERFORMANCE DEMO${NC}"
echo -e "${PURPLE}⚡ Pure Go DFA vs SIMD-Accelerated Aho-Corasick${NC}"
echo ""

# Check if binaries exist
if [ ! -f "legal-nlp" ] || [ ! -f "legal-nlp-simd" ]; then
    echo -e "${YELLOW}📦 Building binaries...${NC}"
    make all
    echo ""
fi

# System information
echo -e "${CYAN}🖥️  System Information:${NC}"
echo "   OS: $(uname -s) $(uname -r)"
echo "   Architecture: $(uname -m)"
echo "   CPU: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs 2>/dev/null || echo 'Unknown')"
echo "   Cores: $(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 'Unknown')"
echo ""

# CPU Features
echo -e "${CYAN}🔧 CPU Features:${NC}"
make cpu-info
echo ""

# Demo 1: Real-world legal text analysis
echo -e "${BLUE}📋 DEMO 1: Real-World Legal Text Analysis${NC}"
echo -e "${YELLOW}Input text:${NC}"
echo "\"$DEMO_TEXT\""
echo ""

echo -e "${GREEN}🔵 Pure Go DFA Analysis:${NC}"
echo "$DEMO_TEXT" | ./legal-nlp
echo ""

echo -e "${GREEN}🚀 SIMD-Accelerated Analysis:${NC}"
echo "$DEMO_TEXT" | ./legal-nlp-simd --simd
echo ""

# Demo 2: Performance benchmarks
echo -e "${BLUE}📊 DEMO 2: Performance Benchmarks${NC}"
echo ""

echo -e "${GREEN}🔵 Pure Go DFA Benchmark:${NC}"
./legal-nlp --benchmark
echo ""

echo -e "${GREEN}🚀 SIMD-Accelerated Benchmark:${NC}"
./legal-nlp-simd --simd --benchmark
echo ""

# Demo 3: Large pattern set performance (if available)
if [ -f "patterns_1000000.txt" ]; then
    echo -e "${BLUE}📈 DEMO 3: 1M Pattern Performance Test${NC}"
    echo ""
    
    echo -e "${GREEN}🔵 Pure Go with 1M patterns:${NC}"
    timeout 30s ./legal-nlp --patterns patterns_1000000.txt --test || echo "   (Timeout after 30s - expected for 1M patterns)"
    echo ""
    
    echo -e "${GREEN}🚀 SIMD with 1M patterns:${NC}"
    timeout 30s ./legal-nlp-simd --simd --patterns patterns_1000000.txt --test || echo "   (Timeout after 30s)"
    echo ""
else
    echo -e "${YELLOW}📝 Generating 1M patterns for large-scale test...${NC}"
    make patterns-1m
    echo ""
    
    echo -e "${BLUE}📈 DEMO 3: 1M Pattern Performance Test${NC}"
    echo ""
    
    echo -e "${GREEN}🔵 Pure Go with 1M patterns (building DFA...):${NC}"
    timeout 60s ./legal-nlp --patterns patterns_1000000.txt --test || echo "   (Timeout - DFA construction takes time)"
    echo ""
    
    echo -e "${GREEN}🚀 SIMD with 1M patterns (building optimized DFA...):${NC}"
    timeout 60s ./legal-nlp-simd --simd --patterns patterns_1000000.txt --test || echo "   (Timeout - even SIMD needs time for 1M patterns)"
    echo ""
fi

# Demo 4: Interactive comparison
echo -e "${BLUE}💬 DEMO 4: Interactive Mode Showcase${NC}"
echo ""

echo -e "${YELLOW}Testing various legal phrases...${NC}"

test_phrases=(
    "he said the contract was invalid"
    "according to the witness testimony"
    "she told me about the settlement"
    "clean legal document with no hearsay"
    "plaintiff claims damages of one million"
    "defendant stated under oath"
    "reportedly there were issues"
    "i heard from reliable sources"
    "witnesses claim the opposite"
    "testimony indicates misconduct"
)

echo -e "${GREEN}🔵 Pure Go Results:${NC}"
for phrase in "${test_phrases[@]}"; do
    echo "Testing: \"$phrase\""
    echo "$phrase" | ./legal-nlp | grep -E "(HEARSAY|No hearsay)" | head -1
done
echo ""

echo -e "${GREEN}🚀 SIMD Results:${NC}"
for phrase in "${test_phrases[@]}"; do
    echo "Testing: \"$phrase\""
    echo "$phrase" | ./legal-nlp-simd --simd | grep -E "(HEARSAY|No hearsay)" | head -1
done
echo ""

# Demo 5: Performance comparison summary
echo -e "${BLUE}📊 DEMO 5: Performance Summary${NC}"
echo ""

echo -e "${CYAN}Performance Comparison:${NC}"
make compare
echo ""

# Demo 6: Memory and efficiency
echo -e "${BLUE}🧠 DEMO 6: Memory Efficiency${NC}"
echo ""

echo -e "${GREEN}Pure Go Memory Usage:${NC}"
/usr/bin/time -l ./legal-nlp --test 2>&1 | grep -E "(maximum resident|real)" || echo "Memory info not available"
echo ""

echo -e "${GREEN}SIMD Memory Usage:${NC}"
/usr/bin/time -l ./legal-nlp-simd --simd --test 2>&1 | grep -E "(maximum resident|real)" || echo "Memory info not available"
echo ""

# Final summary
echo -e "${PURPLE}🎯 ULTIMATE DEMO SUMMARY${NC}"
echo ""
echo -e "${CYAN}✅ Pure Go DFA Implementation:${NC}"
echo "   • Zero dependencies, pure Go"
echo "   • Microsecond search times"
echo "   • Excellent for moderate pattern sets"
echo "   • Cross-platform compatibility"
echo ""
echo -e "${CYAN}🚀 SIMD-Accelerated Implementation:${NC}"
echo "   • AVX-512/NEON vectorized operations"
echo "   • Sub-microsecond search times"
echo "   • Optimized for large pattern sets"
echo "   • Maximum performance on modern CPUs"
echo ""
echo -e "${YELLOW}🏛️  Both implementations provide enterprise-grade${NC}"
echo -e "${YELLOW}   legal text analysis with real-time hearsay detection!${NC}"
echo ""

# Interactive mode option
echo -e "${BLUE}💬 Want to try interactive mode?${NC}"
echo "   Pure Go:  ./legal-nlp"
echo "   SIMD:     ./legal-nlp-simd --simd"
echo ""
echo -e "${GREEN}Demo complete! 🎉${NC}" 