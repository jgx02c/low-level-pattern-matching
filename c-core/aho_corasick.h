#ifndef AHO_CORASICK_H
#define AHO_CORASICK_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// Maximum number of patterns and states
#define AC_MAX_PATTERNS 100000
#define AC_MAX_STATES 200000
#define AC_ALPHABET_SIZE 256

// Match result structure
typedef struct {
    uint64_t offset;        // Byte offset in input text
    uint64_t length;        // Length of matched pattern
    uint32_t pattern_id;    // ID of matched pattern
    uint32_t confidence;    // Match confidence (0-100)
} ac_match_t;

// Aho-Corasick state structure (optimized for cache efficiency)
typedef struct {
    uint32_t next[AC_ALPHABET_SIZE];  // Next state transitions
    uint32_t failure;                 // Failure link
    uint32_t output_count;            // Number of patterns ending here
    uint32_t output[8];               // Pattern IDs (up to 8 per state)
} ac_state_t;

// Main Aho-Corasick automaton structure
typedef struct {
    ac_state_t* states;               // State array (aligned for SIMD)
    uint32_t state_count;             // Number of states
    uint32_t pattern_count;           // Number of patterns
    char** patterns;                  // Pattern strings
    uint32_t* pattern_lengths;        // Pattern lengths
    bool simd_enabled;                // SIMD acceleration available
    bool initialized;                 // Automaton built
} ac_automaton_t;

// Performance statistics
typedef struct {
    uint64_t total_searches;
    uint64_t total_matches;
    uint64_t total_bytes_processed;
    uint64_t simd_operations;
    uint64_t fallback_operations;
    double avg_search_time_ns;
} ac_stats_t;

// Core API functions
ac_automaton_t* ac_create(void);
void ac_destroy(ac_automaton_t* ac);

// Pattern management
int ac_add_pattern(ac_automaton_t* ac, const char* pattern, uint32_t pattern_id);
int ac_load_patterns_from_file(ac_automaton_t* ac, const char* filename);
int ac_load_patterns_from_array(ac_automaton_t* ac, const char** patterns, uint32_t count);

// Automaton building
int ac_build(ac_automaton_t* ac);

// Search functions
int ac_search(ac_automaton_t* ac, const char* text, size_t text_len, 
              ac_match_t* matches, size_t max_matches);

// SIMD-accelerated search (if available)
int ac_search_simd(ac_automaton_t* ac, const char* text, size_t text_len,
                   ac_match_t* matches, size_t max_matches);

// Performance and diagnostics
void ac_get_stats(ac_automaton_t* ac, ac_stats_t* stats);
void ac_reset_stats(ac_automaton_t* ac);
bool ac_has_simd_support(void);
const char* ac_get_simd_info(void);

// Memory management utilities
void* ac_aligned_alloc(size_t alignment, size_t size);
void ac_aligned_free(void* ptr);

// CPU feature detection
bool ac_detect_avx512(void);
bool ac_detect_avx2(void);
bool ac_detect_neon(void);

#ifdef __cplusplus
}
#endif

#endif // AHO_CORASICK_H 