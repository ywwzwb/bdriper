# x265 Encoder Parameters

## Basic Settings
- **crf** — Constant Rate Factor (default 28). High quality: 15-18.
- **preset** — Speed/quality. ultrafast to placebo. Recommended: slow or slower.

## Frame & Block
- **ctu** — Max coding unit size. 1080p: 32 (recommended, not 64).
- **qg-size** — QP group size. Lower = better quality + bitrate. Recommended: 8.
- **bframes** — Max consecutive B-frames. Recommended: 6-10.
- **ref** — Reference frames. Recommended: 4-5 (less important than x264).

## Rate Control
- **qcomp** — Temporal rate control. Recommended: 0.65.
- **aq-mode** — AQ mode. 1=high quality, 2=efficient, 3=dark scenes. Recommend: 1 or 2.
- **aq-strength** — AQ strength. Mode 1: 0.8, Mode 2: 0.9.
- **no-sao** — Disable SAO (Sample Adaptive Offset). Recommended ON for high quality (SAO causes blurring).

## Motion & Analysis
- **me** — Motion estimation. 0=dia, 1=hex, 2=umh, 3=star. Recommended: 3 (star).
- **subme** — Subpixel refinement. Recommended: 5.
- **merange** — Search range. Recommended: 38 for 1080p.

## Psycho-visual
- **psy-rd** — Detail/sharpness retention (default 2.0). Lower for low bitrate: 1.5.
- **psy-rdoq** — Fine detail in RDOQ. Recommended: 1.0.
- **rdoq-level** — RDOQ level. 0=off, 1=partial, 2=full. Recommended: 2 (auto-enabled slow+).

## Chroma
- **cbqpoffs** — Cb chroma QP offset. Recommended: -2.
- **crqpoffs** — Cr chroma QP offset. Recommended: -2.

## Other
- **deblock** — Deblocking filter. High quality: -1:-1.
- **keyint** — Max GOP size. Recommended: 360.
- **pbratio** — P/B frame quality ratio. Anime: 1.2.
- **no-open-gop** — Disable open GOP. Recommended ON.
