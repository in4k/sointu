%define BPM 100

%include "../src/sointu_header.inc"

BEGIN_PATTERNS
    PATTERN 64, 0, 68, 0, 32, 0, 0, 0,  75, 0, 78, 0,   0, 0, 0, 0,
END_PATTERNS

BEGIN_TRACKS
    TRACK   VOICES(1),0
END_TRACKS

BEGIN_PATCH
    BEGIN_INSTRUMENT VOICES(1) ; Instrument0
        SU_ENVELOPE MONO,ATTAC(64),DECAY(64),SUSTAIN(64),RELEASE(72),GAIN(128)
        SU_OSCILLAT MONO,TRANSPOSE(64),DETUNE(64),PHASE(0),COLOR(128),SHAPE(64),GAIN(128),FLAGS(TRISAW)
        SU_MULP     MONO
        SU_FILTER   MONO,FREQUENCY(32),RESONANCE(64),FLAGS(HIGHPASS)
        SU_PAN      MONO,PANNING(64)
        SU_OUT      STEREO,GAIN(128)
    END_INSTRUMENT
END_PATCH

%include "../src/sointu_footer.inc"
