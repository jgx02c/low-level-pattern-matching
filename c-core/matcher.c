#include "matcher.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/mman.h>
#include <immintrin.h>
#include <cpuid.h>
#include <ctype.h>

// Forward declarations
static uint64_t fallback_search(
    const char* text,
    size_t text_len,
    match_result_t* results,
    size_t max_results
);

// Global matcher state (shared across FFI calls)
static matcher_state_t g_matcher = {0};

// Legal hearsay patterns (hardcoded for demo)
static const char* legal_patterns[] = {
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
    "as stated by"
};
static const size_t num_legal_patterns = sizeof(legal_patterns) / sizeof(legal_patterns[0]);

// Initialize the matcher with legal hearsay patterns
int matcher_init(matcher_state_t* state) {
    if (state->initialized) {
        return 0; // Already initialized
    }
    
    // Detect CPU features
    state->avx512_available = detect_avx512_support();
    
    // Allocate 64-byte aligned buffer for SIMD patterns
    state->pattern_buffer_size = num_legal_patterns * 64; // 64 bytes per pattern
    state->pattern_buffer = aligned_alloc_64(state->pattern_buffer_size);
    if (!state->pattern_buffer) {
        return -1;
    }
    
    // Zero out the buffer
    memset(state->pattern_buffer, 0, state->pattern_buffer_size);
    
    // Compile patterns to SIMD format
    char* pattern_ptr = (char*)state->pattern_buffer;
    for (size_t i = 0; i < num_legal_patterns; i++) {
        compile_pattern_to_simd(legal_patterns[i], pattern_ptr);
        pattern_ptr += 64; // Next 64-byte aligned slot
    }
    
    state->pattern_count = num_legal_patterns;
    
    // Initialize atomic counters
    atomic_store(&state->stats.total_searches, 0);
    atomic_store(&state->stats.total_matches, 0);
    atomic_store(&state->stats.cache_hits, 0);
    atomic_store(&state->stats.cache_misses, 0);
    atomic_store(&state->stats.simd_operations, 0);
    atomic_store(&state->stats.fallback_operations, 0);
    
    state->initialized = true;
    
    printf("🚀 Matcher initialized: %zu patterns, AVX-512: %s\n", 
           num_legal_patterns, state->avx512_available ? "YES" : "NO");
    
    return 0;
}

// Cleanup matcher resources
void matcher_cleanup(matcher_state_t* state) {
    if (state->pattern_buffer) {
        aligned_free(state->pattern_buffer);
        state->pattern_buffer = NULL;
    }
    state->initialized = false;
}

// Main pattern search function
int search_patterns(
    matcher_state_t* state,
    const char* text,
    size_t text_len,
    match_result_t* results,
    size_t max_results
) {
    if (!state->initialized) {
        return -1;
    }
    
    // Increment search counter atomically
    atomic_fetch_add(&state->stats.total_searches, 1);
    
    uint64_t start_cycles = get_cpu_cycles();
    
    // Call assembly SIMD core
    uint64_t match_count;
    if (state->avx512_available) {
        atomic_fetch_add(&state->stats.simd_operations, 1);
        match_count = simd_search_patterns(text, text_len, results);
    } else {
        atomic_fetch_add(&state->stats.fallback_operations, 1);
        // Fallback to simple string search
        match_count = fallback_search(text, text_len, results, max_results);
    }
    
    uint64_t end_cycles = get_cpu_cycles();
    
    // Update match counter
    atomic_fetch_add(&state->stats.total_matches, match_count);
    
    return (int)match_count;
}

// Fallback search for non-AVX512 systems
static uint64_t fallback_search(
    const char* text,
    size_t text_len,
    match_result_t* results,
    size_t max_results
) {
    uint64_t match_count = 0;
    
    for (size_t i = 0; i < num_legal_patterns && match_count < max_results; i++) {
        const char* pattern = legal_patterns[i];
        size_t pattern_len = strlen(pattern);
        
        // Simple Boyer-Moore-like search
        for (size_t j = 0; j <= text_len - pattern_len; j++) {
            if (strncasecmp(&text[j], pattern, pattern_len) == 0) {
                results[match_count].offset = j;
                results[match_count].length = pattern_len;
                results[match_count].pattern_id = i;
                results[match_count].confidence = 95; // Fixed confidence for demo
                match_count++;
                
                if (match_count >= max_results) break;
            }
        }
    }
    
    return match_count;
}

// Fast single pattern search
int search_single_pattern(
    const char* text,
    size_t text_len,
    const char* pattern,
    size_t pattern_len,
    match_result_t* result
) {
    // Simple implementation for demo
    char* found = strstr(text, pattern);
    if (found) {
        result->offset = found - text;
        result->length = pattern_len;
        result->pattern_id = 0;
        result->confidence = 90;
        return 1;
    }
    return 0;
}

// CPU feature detection
bool detect_avx512_support(void) {
    unsigned int eax, ebx, ecx, edx;
    
    // Check if CPUID leaf 7 is supported
    __cpuid(0, eax, ebx, ecx, edx);
    if (eax < 7) return false;
    
    // Check AVX-512F support (bit 16 of EBX)
    __cpuid_count(7, 0, eax, ebx, ecx, edx);
    return (ebx & (1 << 16)) != 0;
}

bool detect_avx2_support(void) {
    unsigned int eax, ebx, ecx, edx;
    __cpuid_count(7, 0, eax, ebx, ecx, edx);
    return (ebx & (1 << 5)) != 0;
}

const char* get_cpu_features(void) {
    static char features[256];
    features[0] = '\0';
    
    if (detect_avx512_support()) {
        strcat(features, "AVX-512 ");
    }
    if (detect_avx2_support()) {
        strcat(features, "AVX2 ");
    }
    
    // Check SSE support
    unsigned int eax, ebx, ecx, edx;
    __cpuid(1, eax, ebx, ecx, edx);
    if (edx & (1 << 25)) strcat(features, "SSE ");
    if (edx & (1 << 26)) strcat(features, "SSE2 ");
    if (ecx & (1 << 0)) strcat(features, "SSE3 ");
    
    if (strlen(features) == 0) {
        strcpy(features, "Basic x86-64");
    }
    
    return features;
}

// Performance statistics
void get_performance_stats(matcher_state_t* state, perf_stats_t* stats) {
    stats->total_searches = atomic_load(&state->stats.total_searches);
    stats->total_matches = atomic_load(&state->stats.total_matches);
    stats->cache_hits = atomic_load(&state->stats.cache_hits);
    stats->cache_misses = atomic_load(&state->stats.cache_misses);
    stats->simd_operations = atomic_load(&state->stats.simd_operations);
    stats->fallback_operations = atomic_load(&state->stats.fallback_operations);
}

void reset_performance_stats(matcher_state_t* state) {
    atomic_store(&state->stats.total_searches, 0);
    atomic_store(&state->stats.total_matches, 0);
    atomic_store(&state->stats.cache_hits, 0);
    atomic_store(&state->stats.cache_misses, 0);
    atomic_store(&state->stats.simd_operations, 0);
    atomic_store(&state->stats.fallback_operations, 0);
}

// Pattern compilation to SIMD format
int compile_pattern_to_simd(const char* pattern, void* simd_buffer) {
    // Simple implementation: copy pattern to 64-byte aligned buffer
    size_t pattern_len = strlen(pattern);
    memset(simd_buffer, 0, 64);
    
    // Convert to lowercase for case-insensitive matching
    char* buf = (char*)simd_buffer;
    for (size_t i = 0; i < pattern_len && i < 63; i++) {
        buf[i] = tolower(pattern[i]);
    }
    
    return 0;
}

// Memory management utilities
void* aligned_alloc_64(size_t size) {
    void* ptr;
    if (posix_memalign(&ptr, 64, size) != 0) {
        return NULL;
    }
    return ptr;
}

void aligned_free(void* ptr) {
    free(ptr);
}

// Timing utilities
uint64_t get_timestamp_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec * 1000000000ULL + ts.tv_nsec;
}

uint64_t get_cpu_cycles(void) {
    unsigned int lo, hi;
    __asm__ __volatile__("rdtsc" : "=a" (lo), "=d" (hi));
    return ((uint64_t)hi << 32) | lo;
}

// Global initialization function for Go FFI
int global_matcher_init(void) {
    return matcher_init(&g_matcher);
}

// Global search function for Go FFI
int global_search_patterns(
    const char* text,
    size_t text_len,
    match_result_t* results,
    size_t max_results
) {
    return search_patterns(&g_matcher, text, text_len, results, max_results);
}

// Global cleanup for Go FFI
void global_matcher_cleanup(void) {
    matcher_cleanup(&g_matcher);
}

// Global stats for Go FFI
void global_get_stats(perf_stats_t* stats) {
    get_performance_stats(&g_matcher, stats);
} 