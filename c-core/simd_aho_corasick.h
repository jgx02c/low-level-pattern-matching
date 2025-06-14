#ifndef SIMD_AHO_CORASICK_H
#define SIMD_AHO_CORASICK_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>
#include <stdatomic.h>

#ifdef __cplusplus
extern "C" {
#endif

// SIMD configuration
#ifdef __x86_64__
#include <immintrin.h>
#define SIMD_VECTOR_SIZE 64  // AVX-512
#define SIMD_ALIGNMENT 64
#elif defined(__aarch64__)
#include <arm_neon.h>
#define SIMD_VECTOR_SIZE 16  // NEON
#define SIMD_ALIGNMENT 16
#else
#define SIMD_VECTOR_SIZE 8   // Fallback
#define SIMD_ALIGNMENT 8
#endif

// Performance-critical constants
#define AC_ALPHABET_SIZE 256
#define AC_MAX_PATTERNS 2000000
#define AC_MAX_STATES 10000000
#define AC_CACHE_LINE_SIZE 64
#define AC_PREFETCH_DISTANCE 3

// Match result structure (cache-aligned)
typedef struct __attribute__((aligned(16))) {
    uint64_t offset;        // Byte offset in input text
    uint32_t length;        // Length of matched pattern
    uint32_t pattern_id;    // ID of matched pattern
    uint32_t confidence;    // Match confidence (0-100)
    uint32_t _padding;      // Align to 16 bytes
} simd_match_t;

// SIMD-optimized state structure (cache-line aligned)
typedef struct __attribute__((aligned(AC_CACHE_LINE_SIZE))) {
    // Hot path: state transitions (first cache line)
    uint32_t next[AC_ALPHABET_SIZE];    // 1024 bytes = 16 cache lines
    
    // Cold path: failure and output info (separate cache lines)
    uint32_t failure;                   // Failure link
    uint16_t output_count;              // Number of patterns ending here
    uint16_t _padding1;
    uint32_t output_offset;             // Offset into output array
    uint32_t _padding2[3];              // Pad to cache line boundary
} simd_ac_state_t;

// Compact output storage
typedef struct {
    uint32_t pattern_id;
    uint32_t pattern_length;
} ac_output_t;

// SIMD automaton structure
typedef struct {
    // Core DFA data (cache-aligned)
    simd_ac_state_t* states;            // State array
    ac_output_t* outputs;               // Pattern outputs
    uint32_t state_count;               // Number of states
    uint32_t pattern_count;             // Number of patterns
    uint32_t output_count;              // Total outputs
    
    // SIMD optimization data
    void* simd_lookup_table;            // Vectorized lookup tables
    void* simd_transition_cache;        // Hot state cache
    
    // Performance tracking
    atomic_uint_fast64_t searches;
    atomic_uint_fast64_t matches;
    atomic_uint_fast64_t simd_ops;
    atomic_uint_fast64_t cache_hits;
    
    // CPU features
    bool avx512_available;
    bool avx2_available;
    bool neon_available;
    bool initialized;
} simd_ac_automaton_t;

// Core API functions
simd_ac_automaton_t* simd_ac_create(void);
void simd_ac_destroy(simd_ac_automaton_t* ac);

// Pattern management
int simd_ac_add_pattern(simd_ac_automaton_t* ac, const char* pattern, uint32_t pattern_id);
int simd_ac_load_from_file(simd_ac_automaton_t* ac, const char* filename);
int simd_ac_build(simd_ac_automaton_t* ac);

// Ultra-fast search functions
int simd_ac_search(simd_ac_automaton_t* ac, const char* text, size_t text_len,
                   simd_match_t* matches, size_t max_matches);

// SIMD-specific search variants
int simd_ac_search_avx512(simd_ac_automaton_t* ac, const char* text, size_t text_len,
                          simd_match_t* matches, size_t max_matches);
int simd_ac_search_avx2(simd_ac_automaton_t* ac, const char* text, size_t text_len,
                        simd_match_t* matches, size_t max_matches);
int simd_ac_search_neon(simd_ac_automaton_t* ac, const char* text, size_t text_len,
                        simd_match_t* matches, size_t max_matches);

// Performance optimization
void simd_ac_prefetch_states(simd_ac_automaton_t* ac, uint32_t* state_sequence, size_t count);
void simd_ac_warm_cache(simd_ac_automaton_t* ac);

// Statistics and diagnostics
typedef struct {
    uint64_t total_searches;
    uint64_t total_matches;
    uint64_t simd_operations;
    uint64_t cache_hits;
    uint64_t avg_search_time_ns;
    double simd_utilization;
    const char* simd_variant;
} simd_ac_stats_t;

void simd_ac_get_stats(simd_ac_automaton_t* ac, simd_ac_stats_t* stats);
void simd_ac_reset_stats(simd_ac_automaton_t* ac);

// CPU feature detection
bool simd_ac_detect_avx512(void);
bool simd_ac_detect_avx2(void);
bool simd_ac_detect_neon(void);
const char* simd_ac_get_cpu_info(void);

// Memory management
void* simd_ac_aligned_alloc(size_t alignment, size_t size);
void simd_ac_aligned_free(void* ptr);

// Timing utilities
uint64_t simd_ac_get_cycles(void);
uint64_t simd_ac_get_time_ns(void);

#ifdef __cplusplus
}
#endif

#endif // SIMD_AHO_CORASICK_H 