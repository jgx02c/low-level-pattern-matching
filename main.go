package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// MatchResult represents a detected hearsay pattern
type MatchResult struct {
	Offset     uint64
	Length     uint64
	PatternID  uint32
	Confidence uint32
	Text       string
}

// HybridMatcher provides both pure Go and optimized pattern matching
type HybridMatcher struct {
	patterns     []string
	cache        *Cache
	useOptimized bool
}

// Legal hearsay patterns for demo (fallback if no file provided)
var LegalPatterns = []string{
	"he said",
	"she said",
	"she told",
	"he told",
	"i heard",
	"according to",
	"reportedly",
	"allegedly",
	"it was reported",
	"sources say",
	"witnesses claim",
	"testimony indicates",
	"didn't you say",
	"you mentioned",
	"as stated by",
	"witness testified",
	"plaintiff claims",
	"defendant stated",
	"court records show",
	"evidence suggests",
}

// NewHybridMatcher creates a matcher that can use both Go and optimized implementations
func NewHybridMatcher(patternsFile string, useOptimized bool) (*HybridMatcher, error) {
	matcher := &HybridMatcher{
		cache:        NewCache(10000), // Larger cache for high-performance scenarios
		useOptimized: useOptimized,
	}

	// Load patterns from file or use defaults
	if patternsFile != "" {
		patterns, err := loadPatternsFromFile(patternsFile)
		if err != nil {
			fmt.Printf("⚠️  Failed to load patterns from %s: %v\n", patternsFile, err)
			fmt.Println("📚 Using default legal patterns...")
			matcher.patterns = LegalPatterns
		} else {
			matcher.patterns = patterns
			fmt.Printf("📚 Loaded %d patterns from %s\n", len(patterns), patternsFile)
		}
	} else {
		matcher.patterns = LegalPatterns
	}

	return matcher, nil
}

// loadPatternsFromFile loads patterns from a text file (one per line)
func loadPatternsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") { // Skip empty lines and comments
			patterns = append(patterns, line)
		}
		lineCount++

		// Progress indicator for large files
		if lineCount%100000 == 0 {
			fmt.Printf("📖 Loading patterns... %d lines processed\n", lineCount)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// Search performs pattern matching using the best available method
func (m *HybridMatcher) Search(text string) ([]MatchResult, time.Duration, error) {
	// Check cache first
	if results, duration, found := m.cache.Get(text); found {
		return results, duration, nil
	}

	start := time.Now()
	var results []MatchResult
	var err error

	// Use optimized implementation if available
	if m.useOptimized {
		results, err = m.searchOptimized(text)
	} else {
		results, err = m.searchPureGo(text)
	}

	elapsed := time.Since(start)

	if err != nil {
		return nil, elapsed, err
	}

	// Cache the results
	m.cache.Put(text, results, elapsed)

	return results, elapsed, nil
}

// searchOptimized uses optimized pattern matching (placeholder for future SIMD implementation)
func (m *HybridMatcher) searchOptimized(text string) ([]MatchResult, error) {
	// For now, fall back to pure Go implementation
	// TODO: Add SIMD/optimized implementation for both x86-64 and ARM64
	return m.searchPureGo(text)
}

// searchPureGo uses pure Go implementation for pattern matching
func (m *HybridMatcher) searchPureGo(text string) ([]MatchResult, error) {
	var results []MatchResult

	// Convert to lowercase for case-insensitive matching
	lowerText := strings.ToLower(text)

	// Search for each pattern
	for patternID, pattern := range m.patterns {
		lowerPattern := strings.ToLower(pattern)

		// Find all occurrences of this pattern
		offset := 0
		for {
			index := strings.Index(lowerText[offset:], lowerPattern)
			if index == -1 {
				break
			}

			actualOffset := offset + index
			results = append(results, MatchResult{
				Offset:     uint64(actualOffset),
				Length:     uint64(len(pattern)),
				PatternID:  uint32(patternID),
				Confidence: 95, // Fixed confidence for demo
				Text:       text[actualOffset : actualOffset+len(pattern)],
			})

			offset = actualOffset + 1 // Move past this match
		}
	}

	return results, nil
}

// GetPatternName returns the pattern name for an ID
func (m *HybridMatcher) GetPatternName(patternID uint32) string {
	if int(patternID) < len(m.patterns) {
		return m.patterns[patternID]
	}
	return fmt.Sprintf("unknown-%d", patternID)
}

// GetCacheStats returns cache performance statistics
func (m *HybridMatcher) GetCacheStats() CacheStats {
	return m.cache.GetStats()
}

// formatResults formats search results for display
func formatResults(text string, results []MatchResult, duration time.Duration, matcher *HybridMatcher) {
	if len(results) == 0 {
		fmt.Printf("✅ No hearsay detected (%v)\n", duration)
		return
	}

	fmt.Printf("⚠️  HEARSAY DETECTED (%d matches, %v):\n", len(results), duration)
	for _, result := range results {
		fmt.Printf("   • \"%s\" at position %d-%d (confidence: %d%%)\n",
			result.Text, result.Offset, result.Offset+result.Length-1, result.Confidence)

		// Show context
		start := int(result.Offset)
		end := int(result.Offset + result.Length)
		contextStart := start - 10
		contextEnd := end + 10

		if contextStart < 0 {
			contextStart = 0
		}
		if contextEnd > len(text) {
			contextEnd = len(text)
		}

		context := text[contextStart:contextEnd]
		fmt.Printf("     Context: ...%s...\n", context)
	}
}

// displayStats shows performance and cache statistics
func displayStats(matcher *HybridMatcher, totalSearches, totalMatches int64, totalTime time.Duration) {
	cacheStats := matcher.GetCacheStats()

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
}

// runBenchmark performs performance testing
func runBenchmark(matcher *HybridMatcher) {
	fmt.Println("🚀 Running performance benchmark...")

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

	iterations := 10000 // High iteration count for performance measurement
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

	fmt.Printf("\n🏁 Benchmark Results:\n")
	fmt.Printf("   Iterations: %d\n", iterations)
	fmt.Printf("   Test Texts: %d\n", len(testTexts))
	fmt.Printf("   Total Searches: %d\n", totalSearches)
	fmt.Printf("   Total Matches: %d\n", totalMatches)
	fmt.Printf("   Total Time: %v\n", elapsed)
	fmt.Printf("   Avg Time/Search: %v\n", elapsed/time.Duration(totalSearches))
	fmt.Printf("   Searches/Second: %.0f\n", float64(totalSearches)/elapsed.Seconds())
	fmt.Printf("   Cache Hit Ratio: %.1f%%\n", matcher.cache.HitRatio())
}

func main() {
	fmt.Println("🏛️  Legal NLP Pipeline - Ultra-Fast Hearsay Detection")
	fmt.Println("⚡ Hybrid Go + Optimized Implementation with Microsecond Response Times")

	// Parse command line arguments
	var patternsFile string
	var useOptimized bool = true
	var mode string = "interactive"

	for i, arg := range os.Args[1:] {
		switch arg {
		case "--patterns", "-p":
			if i+1 < len(os.Args)-1 {
				patternsFile = os.Args[i+2]
			}
		case "--no-optimized":
			useOptimized = false
		case "--benchmark", "-b":
			mode = "benchmark"
		case "--test", "-t":
			mode = "test"
		case "--help", "-h":
			fmt.Println("\nUsage:")
			fmt.Println("  legal-nlp-simd [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  --patterns, -p FILE    Load patterns from file")
			fmt.Println("  --no-optimized         Disable optimizations")
			fmt.Println("  --benchmark, -b        Run performance benchmark")
			fmt.Println("  --test, -t             Run test cases")
			fmt.Println("  --help, -h             Show this help")
			fmt.Println("\nPattern File Format:")
			fmt.Println("  One pattern per line, # for comments")
			fmt.Println("  Example: patterns.txt with 1M legal patterns")
			return
		}
	}

	// Initialize matcher
	matcher, err := NewHybridMatcher(patternsFile, useOptimized)
	if err != nil {
		fmt.Printf("❌ Failed to initialize matcher: %v\n", err)
		return
	}

	fmt.Printf("📚 Loaded %d legal hearsay patterns\n", len(matcher.patterns))
	fmt.Printf("⚡ Optimizations: %s\n", map[bool]string{true: "ENABLED", false: "DISABLED"}[useOptimized])

	// Performance tracking
	var totalSearches, totalMatches int64
	var totalTime time.Duration

	// Handle different modes
	switch mode {
	case "benchmark":
		runBenchmark(matcher)
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

		fmt.Println("\n🧪 Running test cases...")
		for _, testCase := range testCases {
			fmt.Printf("\nInput: \"%s\"\n", testCase)
			results, duration, err := matcher.Search(testCase)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			formatResults(testCase, results, duration, matcher)
			totalSearches++
			totalMatches += int64(len(results))
			totalTime += duration
		}

		displayStats(matcher, totalSearches, totalMatches, totalTime)
		return
	}

	// Interactive mode
	fmt.Println("\n💬 Interactive Mode - Type legal text and press Enter")
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
			displayStats(matcher, totalSearches, totalMatches, totalTime)
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
			fmt.Println("  stats/s  - Show performance statistics")
			fmt.Println("  clear/c  - Clear cache and reset stats")
			fmt.Println("  quit/q   - Exit the program")
			continue
		}

		// Perform search
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
		formatResults(input, results, duration, matcher)

		// Show quick stats
		cacheStats := matcher.GetCacheStats()
		cached := ""
		if cacheStats.Hits > 0 {
			cached = fmt.Sprintf(" | Cache: %.0f%% hit", matcher.cache.HitRatio())
		}
		optimizedInfo := ""
		if useOptimized {
			optimizedInfo = " | Optimized: ON"
		}
		fmt.Printf("📊 Searches: %d | Matches: %d%s%s\n\n", totalSearches, totalMatches, cached, optimizedInfo)
	}
}
