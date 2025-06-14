#include "aho_corasick.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <ctype.h>

#ifdef __x86_64__
#include <immintrin.h>
#include <cpuid.h>
#endif

#ifdef __aarch64__
#include <arm_neon.h>
#endif

// Global performance statistics
static ac_stats_t g_stats = {0};

// CPU feature flags
static bool g_avx512_available = false;
static bool g_avx2_available = false;
static bool g_neon_available = false;
static bool g_features_detected = false;

// Forward declarations
static void detect_cpu_features(void);
static int build_goto_function(ac_automaton_t* ac);
static int build_failure_function(ac_automaton_t* ac);
static int build_output_function(ac_automaton_t* ac);
static uint64_t get_time_ns(void);

// Create a new Aho-Corasick automaton
ac_automaton_t* ac_create(void) {
    if (!g_features_detected) {
        detect_cpu_features();
        g_features_detected = true;
    }
    
    ac_automaton_t* ac = (ac_automaton_t*)calloc(1, sizeof(ac_automaton_t));
    if (!ac) return NULL;
    
    // Allocate aligned memory for states (64-byte alignment for SIMD)
    ac->states = (ac_state_t*)ac_aligned_alloc(64, AC_MAX_STATES * sizeof(ac_state_t));
    if (!ac->states) {
        free(ac);
        return NULL;
    }
    
    // Initialize all states to zero
    memset(ac->states, 0, AC_MAX_STATES * sizeof(ac_state_t));
    
    // Allocate pattern arrays
    ac->patterns = (char**)calloc(AC_MAX_PATTERNS, sizeof(char*));
    ac->pattern_lengths = (uint32_t*)calloc(AC_MAX_PATTERNS, sizeof(uint32_t));
    
    if (!ac->patterns || !ac->pattern_lengths) {
        ac_destroy(ac);
        return NULL;
    }
    
    ac->state_count = 1; // Start with root state
    ac->pattern_count = 0;
    ac->simd_enabled = g_avx512_available || g_avx2_available || g_neon_available;
    ac->initialized = false;
    
    printf("🚀 Aho-Corasick automaton created (SIMD: %s)\n", 
           ac->simd_enabled ? "ENABLED" : "DISABLED");
    
    return ac;
}

// Destroy automaton and free memory
void ac_destroy(ac_automaton_t* ac) {
    if (!ac) return;
    
    if (ac->states) {
        ac_aligned_free(ac->states);
    }
    
    if (ac->patterns) {
        for (uint32_t i = 0; i < ac->pattern_count; i++) {
            free(ac->patterns[i]);
        }
        free(ac->patterns);
    }
    
    if (ac->pattern_lengths) {
        free(ac->pattern_lengths);
    }
    
    free(ac);
}

// Add a single pattern to the automaton
int ac_add_pattern(ac_automaton_t* ac, const char* pattern, uint32_t pattern_id) {
    if (!ac || !pattern || ac->pattern_count >= AC_MAX_PATTERNS) {
        return -1;
    }
    
    size_t len = strlen(pattern);
    if (len == 0) return -1;
    
    // Store pattern (convert to lowercase for case-insensitive matching)
    ac->patterns[ac->pattern_count] = (char*)malloc(len + 1);
    if (!ac->patterns[ac->pattern_count]) return -1;
    
    for (size_t i = 0; i < len; i++) {
        ac->patterns[ac->pattern_count][i] = tolower(pattern[i]);
    }
    ac->patterns[ac->pattern_count][len] = '\0';
    ac->pattern_lengths[ac->pattern_count] = len;
    
    ac->pattern_count++;
    ac->initialized = false; // Need to rebuild
    
    return 0;
}

// Load patterns from file
int ac_load_patterns_from_file(ac_automaton_t* ac, const char* filename) {
    if (!ac || !filename) return -1;
    
    FILE* file = fopen(filename, "r");
    if (!file) return -1;
    
    char line[1024];
    uint32_t pattern_id = 0;
    uint32_t loaded = 0;
    
    while (fgets(line, sizeof(line), file) && ac->pattern_count < AC_MAX_PATTERNS) {
        // Remove newline and trim whitespace
        char* end = line + strlen(line) - 1;
        while (end > line && (*end == '\n' || *end == '\r' || *end == ' ')) {
            *end-- = '\0';
        }
        
        // Skip empty lines and comments
        if (line[0] == '\0' || line[0] == '#') continue;
        
        if (ac_add_pattern(ac, line, pattern_id++) == 0) {
            loaded++;
            
            // Progress indicator for large files
            if (loaded % 100000 == 0) {
                printf("📖 Loaded %u patterns...\n", loaded);
            }
        }
    }
    
    fclose(file);
    printf("📚 Loaded %u patterns from %s\n", loaded, filename);
    return 0;
}

// Load patterns from array
int ac_load_patterns_from_array(ac_automaton_t* ac, const char** patterns, uint32_t count) {
    if (!ac || !patterns) return -1;
    
    for (uint32_t i = 0; i < count && ac->pattern_count < AC_MAX_PATTERNS; i++) {
        if (ac_add_pattern(ac, patterns[i], i) != 0) {
            return -1;
        }
    }
    
    return 0;
}

// Build the Aho-Corasick automaton (DFA construction)
int ac_build(ac_automaton_t* ac) {
    if (!ac || ac->pattern_count == 0) return -1;
    
    printf("🏗️  Building Aho-Corasick automaton with %u patterns...\n", ac->pattern_count);
    uint64_t start_time = get_time_ns();
    
    // Build goto function (trie construction)
    if (build_goto_function(ac) != 0) return -1;
    
    // Build failure function
    if (build_failure_function(ac) != 0) return -1;
    
    // Build output function
    if (build_output_function(ac) != 0) return -1;
    
    ac->initialized = true;
    
    uint64_t build_time = get_time_ns() - start_time;
    printf("✅ Automaton built: %u states, %.2fms\n", 
           ac->state_count, build_time / 1000000.0);
    
    return 0;
}

// Build goto function (construct trie)
static int build_goto_function(ac_automaton_t* ac) {
    // Reset state count to 1 (root)
    ac->state_count = 1;
    
    for (uint32_t i = 0; i < ac->pattern_count; i++) {
        const char* pattern = ac->patterns[i];
        uint32_t state = 0; // Start at root
        
        // Follow existing path or create new states
        for (uint32_t j = 0; j < ac->pattern_lengths[i]; j++) {
            unsigned char c = (unsigned char)pattern[j];
            
            if (ac->states[state].next[c] == 0) {
                // Create new state
                if (ac->state_count >= AC_MAX_STATES) return -1;
                
                ac->states[state].next[c] = ac->state_count;
                memset(&ac->states[ac->state_count], 0, sizeof(ac_state_t));
                ac->state_count++;
            }
            
            state = ac->states[state].next[c];
        }
        
        // Mark this state as accepting this pattern
        if (ac->states[state].output_count < 8) {
            ac->states[state].output[ac->states[state].output_count++] = i;
        }
    }
    
    return 0;
}

// Build failure function (KMP-like failure links)
static int build_failure_function(ac_automaton_t* ac) {
    // Use dynamic allocation for queue to avoid stack overflow
    uint32_t* queue = (uint32_t*)malloc(ac->state_count * sizeof(uint32_t));
    if (!queue) return -1;
    
    uint32_t front = 0, rear = 0;
    
    // Initialize failure links for depth-1 states
    for (int c = 0; c < AC_ALPHABET_SIZE; c++) {
        uint32_t state = ac->states[0].next[c];
        if (state != 0) {
            ac->states[state].failure = 0;
            if (rear < ac->state_count) {
                queue[rear++] = state;
            }
        }
    }
    
    // Build failure links for deeper states
    while (front < rear) {
        uint32_t r = queue[front++];
        
        for (int c = 0; c < AC_ALPHABET_SIZE; c++) {
            uint32_t u = ac->states[r].next[c];
            if (u == 0) continue;
            
            if (rear < ac->state_count) {
                queue[rear++] = u;
            }
            
            uint32_t state = ac->states[r].failure;
            while (state != 0 && ac->states[state].next[c] == 0) {
                state = ac->states[state].failure;
            }
            
            ac->states[u].failure = ac->states[state].next[c];
        }
    }
    
    free(queue);
    return 0;
}

// Build output function (collect all patterns ending at each state)
static int build_output_function(ac_automaton_t* ac) {
    for (uint32_t i = 1; i < ac->state_count; i++) {
        uint32_t failure_state = ac->states[i].failure;
        
        // Copy outputs from failure state
        for (uint32_t j = 0; j < ac->states[failure_state].output_count; j++) {
            if (ac->states[i].output_count < 8) {
                ac->states[i].output[ac->states[i].output_count++] = 
                    ac->states[failure_state].output[j];
            }
        }
    }
    
    return 0;
}

// Main search function
int ac_search(ac_automaton_t* ac, const char* text, size_t text_len, 
              ac_match_t* matches, size_t max_matches) {
    if (!ac || !ac->initialized || !text || !matches) return -1;
    
    uint64_t start_time = get_time_ns();
    
    int result;
    if (ac->simd_enabled && text_len > 64) {
        result = ac_search_simd(ac, text, text_len, matches, max_matches);
        g_stats.simd_operations++;
    } else {
        // Fallback to standard search
        uint32_t state = 0;
        size_t match_count = 0;
        
        for (size_t i = 0; i < text_len && match_count < max_matches; i++) {
            unsigned char c = tolower(text[i]);
            
            // Follow failure links until we find a transition
            while (state != 0 && ac->states[state].next[c] == 0) {
                state = ac->states[state].failure;
            }
            
            state = ac->states[state].next[c];
            
            // Check for matches at this state
            for (uint32_t j = 0; j < ac->states[state].output_count && match_count < max_matches; j++) {
                uint32_t pattern_id = ac->states[state].output[j];
                uint32_t pattern_len = ac->pattern_lengths[pattern_id];
                
                matches[match_count].offset = i - pattern_len + 1;
                matches[match_count].length = pattern_len;
                matches[match_count].pattern_id = pattern_id;
                matches[match_count].confidence = 95;
                match_count++;
            }
        }
        
        result = match_count;
        g_stats.fallback_operations++;
    }
    
    uint64_t search_time = get_time_ns() - start_time;
    g_stats.total_searches++;
    g_stats.total_matches += result;
    g_stats.total_bytes_processed += text_len;
    g_stats.avg_search_time_ns = (g_stats.avg_search_time_ns * (g_stats.total_searches - 1) + search_time) / g_stats.total_searches;
    
    return result;
}

// SIMD-accelerated search
int ac_search_simd(ac_automaton_t* ac, const char* text, size_t text_len,
                   ac_match_t* matches, size_t max_matches) {
    // For now, fall back to standard search
    // TODO: Implement SIMD-optimized state transitions
    return ac_search(ac, text, text_len, matches, max_matches);
}

// CPU feature detection
static void detect_cpu_features(void) {
#ifdef __x86_64__
    unsigned int eax, ebx, ecx, edx;
    
    // Check for AVX-512
    if (__get_cpuid_max(0, NULL) >= 7) {
        __cpuid_count(7, 0, eax, ebx, ecx, edx);
        g_avx512_available = (ebx & (1 << 16)) != 0; // AVX-512F
        g_avx2_available = (ebx & (1 << 5)) != 0;    // AVX2
    }
#endif

#ifdef __aarch64__
    // ARM64 NEON is always available
    g_neon_available = true;
#endif

    printf("🔍 CPU Features: AVX-512=%s, AVX2=%s, NEON=%s\n",
           g_avx512_available ? "YES" : "NO",
           g_avx2_available ? "YES" : "NO", 
           g_neon_available ? "YES" : "NO");
}

// Performance statistics
void ac_get_stats(ac_automaton_t* ac, ac_stats_t* stats) {
    if (stats) {
        *stats = g_stats;
    }
}

void ac_reset_stats(ac_automaton_t* ac) {
    memset(&g_stats, 0, sizeof(g_stats));
}

bool ac_has_simd_support(void) {
    return g_avx512_available || g_avx2_available || g_neon_available;
}

const char* ac_get_simd_info(void) {
    static char info[256];
    snprintf(info, sizeof(info), "AVX-512: %s, AVX2: %s, NEON: %s",
             g_avx512_available ? "YES" : "NO",
             g_avx2_available ? "YES" : "NO",
             g_neon_available ? "YES" : "NO");
    return info;
}

// Memory management
void* ac_aligned_alloc(size_t alignment, size_t size) {
    void* ptr;
    if (posix_memalign(&ptr, alignment, size) != 0) {
        return NULL;
    }
    return ptr;
}

void ac_aligned_free(void* ptr) {
    free(ptr);
}

// CPU feature detection functions
bool ac_detect_avx512(void) { return g_avx512_available; }
bool ac_detect_avx2(void) { return g_avx2_available; }
bool ac_detect_neon(void) { return g_neon_available; }

// Timing utility
static uint64_t get_time_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec * 1000000000ULL + ts.tv_nsec;
} 