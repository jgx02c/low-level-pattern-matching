#ifndef MATCHER_H
#define MATCHER_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>
#include <stdatomic.h>

#ifdef __cplusplus
extern "C" {
#endif

// Match result structure
typedef struct {
    uint64_t offset;        // Byte offset in input text
    uint64_t length;        // Length of matched pattern
    uint32_t pattern_id;    // ID of matched pattern
    uint32_t confidence;    // Match confidence (0-100)
} match_result_t;

// Performance statistics (atomic for lock-free access)
typedef struct {
    atomic_uint_fast64_t total_searches;
    atomic_uint_fast64_t total_matches;
    atomic_uint_fast64_t cache_hits;
    atomic_uint_fast64_t cache_misses;
    atomic_uint_fast64_t simd_operations;
    atomic_uint_fast64_t fallback_operations;
} perf_stats_t;

// Matcher state structure
typedef struct {
    void* pattern_buffer;           // Pre-compiled SIMD patterns
    size_t pattern_buffer_size;     // Buffer size in bytes
    uint32_t pattern_count;         // Number of loaded patterns
    perf_stats_t stats;             // Performance counters
    bool avx512_available;          // CPU feature detection
    bool initialized;               // Initialization status
} matcher_state_t;

// Initialize the matcher with legal hearsay patterns
int matcher_init(matcher_state_t* state);

// Cleanup matcher resources
void matcher_cleanup(matcher_state_t* state);

// Main pattern search function (calls assembly SIMD core)
int search_patterns(
    matcher_state_t* state,
    const char* text,
    size_t text_len,
    match_result_t* results,
    size_t max_results
);

// Fast single pattern search
int search_single_pattern(
    const char* text,
    size_t text_len,
    const char* pattern,
    size_t pattern_len,
    match_result_t* result
);

// Performance and diagnostics
void get_performance_stats(matcher_state_t* state, perf_stats_t* stats);
void reset_performance_stats(matcher_state_t* state);

// CPU feature detection
bool detect_avx512_support(void);
bool detect_avx2_support(void);
const char* get_cpu_features(void);

// Assembly function declarations (implemented in simd_match.s)
extern uint64_t simd_search_patterns(
    const char* text,
    size_t text_len,
    match_result_t* results
);

extern uint64_t simd_search_single(
    const char* text,
    size_t text_len,
    const char* pattern,
    size_t pattern_len
);

extern uint64_t get_pattern_count(void);

// Pattern compilation and loading
int load_legal_patterns(matcher_state_t* state, const char* patterns_file);
int compile_pattern_to_simd(const char* pattern, void* simd_buffer);

// Memory management utilities
void* aligned_alloc_64(size_t size);
void aligned_free(void* ptr);

// Timing utilities for performance measurement
uint64_t get_timestamp_ns(void);
uint64_t get_cpu_cycles(void);

#ifdef __cplusplus
}
#endif

#endif // MATCHER_H 