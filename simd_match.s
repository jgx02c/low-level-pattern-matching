; Ultra-Low Latency SIMD Pattern Matching
; AVX-512 optimized assembly for legal hearsay detection
; Target: Sub-50μs pattern matching performance

section .data
    align 64
    ; Pre-compiled legal hearsay patterns (64-byte aligned for AVX-512)
    pattern_he_said:    db "he said", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    pattern_she_told:   db "she told", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    pattern_i_heard:    db "i heard", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    pattern_according:  db "according to", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    
    ; Pattern metadata
    pattern_count: dq 4
    pattern_lengths: dq 7, 8, 7, 12  ; lengths of each pattern

section .text
    global simd_search_patterns
    global simd_search_single
    global get_pattern_count

; Main SIMD pattern search function
; RDI = input text pointer
; RSI = text length  
; RDX = results buffer (output)
; Returns: number of matches found
simd_search_patterns:
    push rbp
    mov rbp, rsp
    push rbx
    push r12
    push r13
    push r14
    push r15
    
    ; Initialize registers
    mov r12, rdi        ; input text
    mov r13, rsi        ; text length
    mov r14, rdx        ; results buffer
    xor r15, r15        ; match counter
    
    ; Check if we have AVX-512 support
    mov eax, 7
    xor ecx, ecx
    cpuid
    test ebx, 0x10000   ; Check AVX-512F bit
    jz fallback_sse     ; Fall back to SSE if no AVX-512
    
    ; AVX-512 main loop
avx512_loop:
    cmp r13, 64
    jl final_bytes      ; Less than 64 bytes remaining
    
    ; Load 64 bytes of text into ZMM0
    vmovdqu8 zmm0, [r12]
    
    ; Convert to lowercase for case-insensitive matching
    call to_lowercase_avx512
    
    ; Search for each pattern
    call search_he_said_avx512
    call search_she_told_avx512  
    call search_i_heard_avx512
    call search_according_avx512
    
    ; Advance pointer and continue
    add r12, 64
    sub r13, 64
    jmp avx512_loop

search_he_said_avx512:
    ; Load pattern into ZMM1
    vmovdqu8 zmm1, [pattern_he_said]
    
    ; Compare with sliding window technique
    mov rcx, 58         ; 64 - 7 + 1 (pattern length)
search_he_said_loop:
    ; Extract 7 bytes starting at offset RCX
    vpshufb zmm2, zmm0, zmm3  ; zmm3 contains shuffle mask for offset
    vpcmpeqb k1, zmm2, zmm1
    kortestq k1, k1
    jnz he_said_found
    
    dec rcx
    jnz search_he_said_loop
    ret

he_said_found:
    ; Store match result
    mov rax, r12
    sub rax, 64
    add rax, rcx
    mov [r14 + r15*16], rax     ; offset
    mov [r14 + r15*16 + 8], 7   ; length
    inc r15
    ret

search_she_told_avx512:
    ; Similar pattern for "she told" - 8 bytes
    vmovdqu8 zmm1, [pattern_she_told]
    mov rcx, 57         ; 64 - 8 + 1
search_she_told_loop:
    vpshufb zmm2, zmm0, zmm3
    vpcmpeqb k1, zmm2, zmm1
    kortestq k1, k1
    jnz she_told_found
    dec rcx
    jnz search_she_told_loop
    ret

she_told_found:
    mov rax, r12
    sub rax, 64
    add rax, rcx
    mov [r14 + r15*16], rax
    mov [r14 + r15*16 + 8], 8
    inc r15
    ret

search_i_heard_avx512:
    ; Similar pattern for "i heard" - 7 bytes
    vmovdqu8 zmm1, [pattern_i_heard]
    mov rcx, 58
search_i_heard_loop:
    vpshufb zmm2, zmm0, zmm3
    vpcmpeqb k1, zmm2, zmm1
    kortestq k1, k1
    jnz i_heard_found
    dec rcx
    jnz search_i_heard_loop
    ret

i_heard_found:
    mov rax, r12
    sub rax, 64
    add rax, rcx
    mov [r14 + r15*16], rax
    mov [r14 + r15*16 + 8], 7
    inc r15
    ret

search_according_avx512:
    ; Similar pattern for "according to" - 12 bytes
    vmovdqu8 zmm1, [pattern_according]
    mov rcx, 53         ; 64 - 12 + 1
search_according_loop:
    vpshufb zmm2, zmm0, zmm3
    vpcmpeqb k1, zmm2, zmm1
    kortestq k1, k1
    jnz according_found
    dec rcx
    jnz search_according_loop
    ret

according_found:
    mov rax, r12
    sub rax, 64
    add rax, rcx
    mov [r14 + r15*16], rax
    mov [r14 + r15*16 + 8], 12
    inc r15
    ret

to_lowercase_avx512:
    ; Convert ZMM0 to lowercase
    vpcmpgtb k1, zmm0, [uppercase_a_vec]
    vpcmpltb k2, zmm0, [uppercase_z_vec]
    kandq k1, k1, k2
    vpaddb zmm0{k1}, zmm0, [case_diff_vec]
    ret

fallback_sse:
    ; SSE2 fallback for older CPUs
    ; Simplified linear search
sse_loop:
    cmp r13, 0
    je done
    
    ; Simple byte-by-byte comparison
    mov al, [r12]
    cmp al, 'h'
    je check_he_said_sse
    cmp al, 's'  
    je check_she_told_sse
    cmp al, 'i'
    je check_i_heard_sse
    cmp al, 'a'
    je check_according_sse
    
    inc r12
    dec r13
    jmp sse_loop

check_he_said_sse:
    ; Simple string comparison for "he said"
    cmp r13, 7
    jl sse_next
    lea rsi, [pattern_he_said]
    mov rcx, 7
    repe cmpsb
    je he_said_found_sse
    sub r12, 6  ; Reset pointer
    jmp sse_next

he_said_found_sse:
    sub r12, 7
    mov [r14 + r15*16], r12
    mov [r14 + r15*16 + 8], 7
    inc r15
    add r12, 7
    sub r13, 7
    jmp sse_loop

check_she_told_sse:
    ; Similar for other patterns...
    ; (Simplified for demo)
    jmp sse_next

check_i_heard_sse:
    jmp sse_next

check_according_sse:
    jmp sse_next

sse_next:
    inc r12
    dec r13
    jmp sse_loop

final_bytes:
    ; Handle remaining bytes < 64
    cmp r13, 0
    je done
    
    ; Use SSE for final bytes
    jmp sse_loop

done:
    mov rax, r15        ; return match count
    
    pop r15
    pop r14
    pop r13
    pop r12
    pop rbx
    pop rbp
    ret

; Single pattern search function (for testing)
simd_search_single:
    push rbp
    mov rbp, rsp
    
    ; Simple implementation for testing
    xor rax, rax
    
    pop rbp
    ret

; Get number of compiled patterns
get_pattern_count:
    mov rax, [pattern_count]
    ret

section .data
    align 64
    uppercase_a_vec: times 64 db 'A'
    uppercase_z_vec: times 64 db 'Z'
    case_diff_vec:   times 64 db 32    ; 'a' - 'A' = 32 