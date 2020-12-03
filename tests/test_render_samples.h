#ifndef SU_RENDER_H
#define SU_RENDER_H

#define SU_MAX_SAMPLES     105840
#define SU_BUFFER_LENGTH   (SU_MAX_SAMPLES*2)

#define SU_SAMPLE_RATE     44100
#define SU_BPM             100
#define SU_PATTERN_SIZE    16
#define SU_MAX_PATTERNS    1
#define SU_TOTAL_ROWS      (SU_MAX_PATTERNS*SU_PATTERN_SIZE)
#define SU_SAMPLES_PER_ROW (SU_SAMPLE_RATE*4*60/(BPM*16))

#include <stdint.h>
#if UINTPTR_MAX == 0xffffffff
    #if defined(__clang__) || defined(__GNUC__)
        #define SU_CALLCONV __attribute__ ((stdcall))
    #elif defined(_WIN32)
        #define SU_CALLCONV __stdcall
    #endif
#else
    #define SU_CALLCONV
#endif

typedef float SUsample;
#define SU_SAMPLE_RANGE 1.0f

#ifdef __cplusplus
extern "C" {
#endif

void SU_CALLCONV su_render_song(SUsample *buffer);

#ifdef __cplusplus
}
#endif

#endif
