package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// MatchResult represents a detected hearsay pattern (pure Go version)
type MatchResult struct {
	Offset     uint64
	Length     uint64
	PatternID  uint32
	Confidence uint32
	Text       string
}

// PureMatcher provides fast Go-based pattern matching
type PureMatcher struct {
	patterns []string
	cache    *Cache
}

// Legal hearsay patterns for demo
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
}

// NewPureMatcher creates a pure Go matcher
func NewPureMatcher() *PureMatcher {
	return &PureMatcher{
		patterns: LegalPatterns,
		cache:    NewCache(1000), // Cache up to 1000 results
	}
}

// Search performs fast pattern matching using Go
func (m *PureMatcher) Search(text string) ([]MatchResult, time.Duration, error) {
	// Check cache first
	if results, duration, found := m.cache.Get(text); found {
		return results, duration, nil
	}

	start := time.Now()
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

	elapsed := time.Since(start)

	// Cache the results
	m.cache.Put(text, results, elapsed)

	return results, elapsed, nil
}

// GetPatternName returns the pattern name for an ID
func (m *PureMatcher) GetPatternName(patternID uint32) string {
	if int(patternID) < len(m.patterns) {
		return m.patterns[patternID]
	}
	return fmt.Sprintf("unknown-%d", patternID)
}

// GetCacheStats returns cache performance statistics
func (m *PureMatcher) GetCacheStats() CacheStats {
	return m.cache.GetStats()
}

// formatResults formats search results for display
func formatResults(text string, results []MatchResult, duration time.Duration, matcher *PureMatcher) {
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
func displayStats(matcher *PureMatcher, totalSearches, totalMatches int64, totalTime time.Duration) {
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
func runBenchmark(matcher *PureMatcher) {
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
	}

	iterations := 1000
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
	fmt.Println("⚡ Pure Go Implementation with Microsecond Response Times")

	// Initialize matcher
	matcher := NewPureMatcher()
	fmt.Printf("📚 Loaded %d legal hearsay patterns\n", len(LegalPatterns))

	// Performance tracking
	var totalSearches, totalMatches int64
	var totalTime time.Duration

	// Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--benchmark", "-b":
			runBenchmark(matcher)
			return
		case "--test", "-t":
			// Run test cases
			testCases := []string{
				"he said the defendant was guilty",
				"according to witnesses, the meeting was productive",
				"clean legal text with no hearsay",
				"she told me about the contract terms",
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
		case "--help", "-h":
			fmt.Println("\nUsage:")
			fmt.Println("  legal-nlp-simd                Interactive mode")
			fmt.Println("  legal-nlp-simd --benchmark     Run performance benchmark")
			fmt.Println("  legal-nlp-simd --test          Run test cases")
			fmt.Println("  legal-nlp-simd --help          Show this help")
			return
		}
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
		fmt.Printf("📊 Searches: %d | Matches: %d%s\n\n", totalSearches, totalMatches, cached)
	}
}
