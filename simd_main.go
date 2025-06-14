//go:build cgo
// +build cgo

package main

/*
#cgo CFLAGS: -O3 -std=c11 -Ic-core
#cgo darwin,amd64 CFLAGS: -mavx512f -mavx2 -msse4.2
#cgo linux,amd64 CFLAGS: -mavx512f -mavx2 -msse4.2
#cgo darwin,arm64 CFLAGS: -mcpu=apple-m1
#cgo linux,arm64 CFLAGS: -mcpu=native
#cgo LDFLAGS: -lm
#include "c-core/simd_aho_corasick.h"
#include "c-core/simd_aho_corasick.c"
#include <stdlib.h>
*/
import "C"

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"
)

// SIMDMatcher provides ultra-fast SIMD-accelerated pattern matching
type SIMDMatcher struct {
	automaton   *C.simd_ac_automaton_t
	patterns    []string
	cache       *Cache
	initialized bool
}

// NewSIMDMatcher creates a SIMD-accelerated pattern matcher
func NewSIMDMatcher(patternsFile string) (*SIMDMatcher, error) {
	matcher := &SIMDMatcher{
		cache: NewCache(10000),
	}

	// Create SIMD automaton
	matcher.automaton = C.simd_ac_create()
	if matcher.automaton == nil {
		return nil, fmt.Errorf("failed to create SIMD automaton")
	}

	// Load patterns
	if patternsFile != "" {
		// Load directly from file in C for maximum performance
		cFilename := C.CString(patternsFile)
		defer C.free(unsafe.Pointer(cFilename))

		result := C.simd_ac_load_from_file(matcher.automaton, cFilename)
		if result != 0 {
			C.simd_ac_destroy(matcher.automaton)
			return nil, fmt.Errorf("failed to load patterns from file")
		}

		// Load patterns into Go for pattern name lookup
		patterns, err := loadPatternsFromFile(patternsFile)
		if err == nil {
			matcher.patterns = patterns
		}
	} else {
		// Load default patterns
		matcher.patterns = LegalPatterns
		if err := matcher.loadDefaultPatterns(); err != nil {
			C.simd_ac_destroy(matcher.automaton)
			return nil, err
		}
	}

	matcher.initialized = true
	fmt.Printf("✅ SIMD matcher ready with %d patterns\n", len(matcher.patterns))

	return matcher, nil
}

// loadDefaultPatterns loads the default legal patterns into the SIMD core
func (m *SIMDMatcher) loadDefaultPatterns() error {
	for i, pattern := range m.patterns {
		cPattern := C.CString(pattern)
		result := C.simd_ac_add_pattern(m.automaton, cPattern, C.uint32_t(i))
		C.free(unsafe.Pointer(cPattern))

		if result != 0 {
			return fmt.Errorf("failed to add pattern: %s", pattern)
		}
	}

	// Build the automaton
	result := C.simd_ac_build(m.automaton)
	if result != 0 {
		return fmt.Errorf("failed to build SIMD automaton")
	}

	return nil
}

// Search performs ultra-fast SIMD pattern matching
func (m *SIMDMatcher) Search(text string) ([]MatchResult, time.Duration, error) {
	if !m.initialized {
		return nil, 0, fmt.Errorf("matcher not initialized")
	}

	// Check cache first
	if results, duration, found := m.cache.Get(text); found {
		return results, duration, nil
	}

	start := time.Now()

	// Convert Go string to C string
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// Allocate results buffer
	maxResults := 1000
	cMatches := make([]C.simd_match_t, maxResults)

	// Call SIMD search
	matchCount := C.simd_ac_search(
		m.automaton,
		cText,
		C.size_t(len(text)),
		&cMatches[0],
		C.size_t(maxResults),
	)

	elapsed := time.Since(start)

	if matchCount < 0 {
		return nil, elapsed, fmt.Errorf("SIMD search failed")
	}

	// Convert C results to Go results
	results := make([]MatchResult, matchCount)
	for i := 0; i < int(matchCount); i++ {
		cMatch := cMatches[i]

		// Get pattern text
		patternText := ""
		if int(cMatch.pattern_id) < len(m.patterns) {
			patternText = m.patterns[cMatch.pattern_id]
		} else {
			// Extract from original text
			start := int(cMatch.offset)
			end := start + int(cMatch.length)
			if end <= len(text) {
				patternText = text[start:end]
			}
		}

		results[i] = MatchResult{
			Offset:     uint64(cMatch.offset),
			Length:     uint64(cMatch.length),
			PatternID:  uint32(cMatch.pattern_id),
			Confidence: uint32(cMatch.confidence),
			Text:       patternText,
		}
	}

	// Cache the results
	m.cache.Put(text, results, elapsed)

	return results, elapsed, nil
}

// GetSIMDStats returns SIMD performance statistics
func (m *SIMDMatcher) GetSIMDStats() map[string]interface{} {
	if !m.initialized {
		return map[string]interface{}{}
	}

	var cStats C.simd_ac_stats_t
	C.simd_ac_get_stats(m.automaton, &cStats)

	return map[string]interface{}{
		"total_searches":     uint64(cStats.total_searches),
		"total_matches":      uint64(cStats.total_matches),
		"simd_operations":    uint64(cStats.simd_operations),
		"cache_hits":         uint64(cStats.cache_hits),
		"avg_search_time_ns": uint64(cStats.avg_search_time_ns),
		"simd_utilization":   float64(cStats.simd_utilization),
		"simd_variant":       C.GoString(cStats.simd_variant),
		"cpu_info":           C.GoString(C.simd_ac_get_cpu_info()),
	}
}

// GetPatternName returns the pattern name for an ID
func (m *SIMDMatcher) GetPatternName(patternID uint32) string {
	if int(patternID) < len(m.patterns) {
		return m.patterns[patternID]
	}
	return fmt.Sprintf("unknown-%d", patternID)
}

// GetCacheStats returns cache performance statistics
func (m *SIMDMatcher) GetCacheStats() CacheStats {
	return m.cache.GetStats()
}

// Cleanup releases SIMD resources
func (m *SIMDMatcher) Cleanup() {
	if m.automaton != nil {
		C.simd_ac_destroy(m.automaton)
		m.automaton = nil
		m.initialized = false
	}
}

// displaySIMDStats shows SIMD performance statistics
func displaySIMDStats(matcher *SIMDMatcher, totalSearches, totalMatches int64, totalTime time.Duration) {
	cacheStats := matcher.GetCacheStats()
	simdStats := matcher.GetSIMDStats()

	fmt.Printf("\n📊 Performance Statistics:\n")
	fmt.Printf("   Total Searches: %d\n", totalSearches)
	fmt.Printf("   Total Matches: %d\n", totalMatches)
	fmt.Printf("   Total Time: %v\n", totalTime)
	if totalSearches > 0 {
		fmt.Printf("   Avg Time/Search: %v\n", totalTime/time.Duration(totalSearches))
		fmt.Printf("   Searches/Second: %.0f\n", float64(totalSearches)/totalTime.Seconds())
	}

	fmt.Printf("\n🗄️  Cache Statistics:\n")
	fmt.Printf("   Cache Hits: %d\n", cacheStats.Hits)
	fmt.Printf("   Cache Misses: %d\n", cacheStats.Misses)
	fmt.Printf("   Hit Ratio: %.1f%%\n", matcher.cache.HitRatio())
	fmt.Printf("   Cached Entries: %d\n", cacheStats.TotalEntries)

	if len(simdStats) > 0 {
		fmt.Printf("\n⚡ SIMD Core Statistics:\n")
		fmt.Printf("   SIMD Variant: %v\n", simdStats["simd_variant"])
		fmt.Printf("   CPU Info: %v\n", simdStats["cpu_info"])
		fmt.Printf("   Core Searches: %v\n", simdStats["total_searches"])
		fmt.Printf("   Core Matches: %v\n", simdStats["total_matches"])
		fmt.Printf("   SIMD Operations: %v\n", simdStats["simd_operations"])
		fmt.Printf("   SIMD Utilization: %.1f%%\n", simdStats["simd_utilization"])
		if avgTime, ok := simdStats["avg_search_time_ns"].(uint64); ok && avgTime > 0 {
			fmt.Printf("   Avg Core Time: %dns\n", avgTime)
		}
	}
}

// runSIMDBenchmark performs SIMD performance testing
func runSIMDBenchmark(matcher *SIMDMatcher) {
	fmt.Println("🚀 Running SIMD Aho-Corasick benchmark...")

	testTexts := []string{
		"he said the defendant was guilty",
		"according to the witness testimony, the case was clear",
		"she told me that it happened yesterday during the meeting",
		"the contract was signed without any issues whatsoever",
		"reportedly there were serious problems with the case",
		"i heard from multiple sources about this incident",
		"this is clean legal text with no hearsay indicators",
		"witnesses claim that the events unfolded differently",
		"testimony indicates a pattern of misconduct over time",
		"didn't you say something different during your deposition",
		"plaintiff claims damages in excess of one million dollars",
		"defendant stated under oath that the allegations were false",
		"court records show a pattern of similar complaints",
		"evidence suggests that the incident occurred as described",
		"witness testified that they saw the defendant at the scene",
	}

	iterations := 10000
	start := time.Now()
	totalMatches := 0

	for i := 0; i < iterations; i++ {
		for _, text := range testTexts {
			results, _, err := matcher.Search(text)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			totalMatches += len(results)
		}
	}

	elapsed := time.Since(start)
	totalSearches := iterations * len(testTexts)

	fmt.Printf("\n🏁 SIMD Benchmark Results:\n")
	fmt.Printf("   Iterations: %d\n", iterations)
	fmt.Printf("   Test Texts: %d\n", len(testTexts))
	fmt.Printf("   Total Searches: %d\n", totalSearches)
	fmt.Printf("   Total Matches: %d\n", totalMatches)
	fmt.Printf("   Total Time: %v\n", elapsed)
	fmt.Printf("   Avg Time/Search: %v\n", elapsed/time.Duration(totalSearches))
	fmt.Printf("   Searches/Second: %.0f\n", float64(totalSearches)/elapsed.Seconds())
	fmt.Printf("   Cache Hit Ratio: %.1f%%\n", matcher.cache.HitRatio())

	// Show detailed SIMD stats
	simdStats := matcher.GetSIMDStats()
	if len(simdStats) > 0 {
		fmt.Printf("\n⚡ SIMD Core Performance:\n")
		fmt.Printf("   SIMD Variant: %v\n", simdStats["simd_variant"])
		fmt.Printf("   SIMD Operations: %v\n", simdStats["simd_operations"])
		fmt.Printf("   SIMD Utilization: %.1f%%\n", simdStats["simd_utilization"])
	}
}

func mainSIMD() {
	fmt.Println("🏛️  Legal NLP Pipeline - SIMD Ultra-Fast Hearsay Detection")
	fmt.Println("⚡ SIMD Aho-Corasick + AVX-512/NEON Implementation with Nanosecond Response Times")

	// Parse command line arguments
	var patternsFile string
	var mode string = "interactive"

	for i, arg := range os.Args[1:] {
		switch arg {
		case "--patterns", "-p":
			if i+1 < len(os.Args)-1 {
				patternsFile = os.Args[i+2]
			}
		case "--benchmark", "-b":
			mode = "benchmark"
		case "--test", "-t":
			mode = "test"
		case "--simd":
			// SIMD mode flag (this function is SIMD by default)
		case "--help", "-h":
			fmt.Println("\nUsage:")
			fmt.Println("  legal-nlp-simd --simd [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  --patterns, -p FILE    Load patterns from file")
			fmt.Println("  --benchmark, -b        Run SIMD benchmark")
			fmt.Println("  --test, -t             Run SIMD test cases")
			fmt.Println("  --help, -h             Show this help")
			fmt.Println("\nFeatures:")
			fmt.Println("  • SIMD-accelerated Aho-Corasick DFA")
			fmt.Println("  • AVX-512, AVX2, NEON support")
			fmt.Println("  • Cache-optimized state transitions")
			fmt.Println("  • Sub-microsecond search times")
			return
		}
	}

	// Initialize SIMD matcher
	matcher, err := NewSIMDMatcher(patternsFile)
	if err != nil {
		fmt.Printf("❌ Failed to initialize SIMD matcher: %v\n", err)
		return
	}
	defer matcher.Cleanup()

	fmt.Printf("📚 SIMD matcher loaded with %d patterns\n", len(matcher.patterns))

	// Performance tracking
	var totalSearches, totalMatches int64
	var totalTime time.Duration

	// Handle different modes
	switch mode {
	case "benchmark":
		runSIMDBenchmark(matcher)
		return
	case "test":
		// Run test cases
		testCases := []string{
			"he said the defendant was guilty",
			"according to witnesses, the meeting was productive",
			"clean legal text with no hearsay",
			"she told me about the contract terms",
			"plaintiff claims damages in the amount of fifty thousand dollars",
			"witness testified that the events occurred as described",
		}

		fmt.Println("\n🧪 Running SIMD test cases...")
		for _, testCase := range testCases {
			fmt.Printf("\nInput: \"%s\"\n", testCase)
			results, duration, err := matcher.Search(testCase)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			formatResults(testCase, results, duration, nil)
			totalSearches++
			totalMatches += int64(len(results))
			totalTime += duration
		}

		displaySIMDStats(matcher, totalSearches, totalMatches, totalTime)
		return
	}

	// Interactive mode
	fmt.Println("\n💬 SIMD Interactive Mode - Type legal text and press Enter")
	fmt.Println("📝 Commands: 'stats' (show stats), 'clear' (clear cache), 'quit' (exit)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Handle commands
		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			fmt.Println("👋 Goodbye!")
			return
		case "stats", "s":
			displaySIMDStats(matcher, totalSearches, totalMatches, totalTime)
			continue
		case "clear", "c":
			matcher.cache.Clear()
			totalSearches = 0
			totalMatches = 0
			totalTime = 0
			fmt.Println("🗑️  Cache and stats cleared")
			continue
		case "help", "h":
			fmt.Println("Commands:")
			fmt.Println("  stats/s  - Show SIMD performance statistics")
			fmt.Println("  clear/c  - Clear cache and reset stats")
			fmt.Println("  quit/q   - Exit the program")
			continue
		}

		// Perform SIMD search
		results, duration, err := matcher.Search(input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Update stats
		totalSearches++
		totalMatches += int64(len(results))
		totalTime += duration

		// Display results
		formatResults(input, results, duration, nil)

		// Show quick stats
		cacheStats := matcher.GetCacheStats()
		cached := ""
		if cacheStats.Hits > 0 {
			cached = fmt.Sprintf(" | Cache: %.0f%% hit", matcher.cache.HitRatio())
		}
		fmt.Printf("📊 Searches: %d | Matches: %d%s | SIMD: ON\n\n", totalSearches, totalMatches, cached)
	}
}
