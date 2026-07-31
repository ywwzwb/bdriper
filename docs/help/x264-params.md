# x264 Encoder Parameters

## Basic Settings
- **crf** — Constant Rate Factor (0-51, default 23). Lower = better quality, larger file. Anime BD: 16-18.
- **preset** — Speed/quality tradeoff. ultrafast to placebo. Recommended: slower or veryslow.
- **tune** — Source type optimization: film, animation, grain, stillimage.

## Frame Types
- **keyint** — Max GOP size (IDR frame interval). Default 250. Recommended: 600 for 24fps.
- **bframes** — Max consecutive B-frames. Anime: 8-12, Film: 4-8.
- **ref** — Reference frames. Higher = better compression, slower decode. Recommended: 6-13.

## Rate Control
- **qcomp** — Temporal rate control (0-1). Default 0.6. Higher = more consistent quality. 0.7-0.8 with mbtree on.
- **aq-mode** — Adaptive quantization mode. 1=default, 2=auto-variance, 3=auto-variance + bias (best for anime).
- **aq-strength** — AQ strength (0-3). Anime: 0.6-1.0, recommend 0.8.
- **mbtree** — Macroblock tree rate control. On for most use. Off for very high bitrate (crf < 16).

## Motion Estimation
- **me** — Motion estimation algorithm. dia < hex < umh < esa < tesa. Recommended: umh or tesa.
- **subme** — Subpixel refinement (0-11). Recommended: 10.
- **me_range** — Search range. 1080p: 24-32.

## Psycho-visual
- **psy-rd** — Psycho-visual optimization. Retains texture detail. Anime: 0.4-1.0.
- **psy-trellis** — Fine detail retention. With mbtree: 0.1-0.15. Without: 0.

## Other
- **deblock** — Deblocking filter (alpha:beta). High quality: -1:-1. Default: 0:0.
- **colormatrix** — Color matrix. BD: bt709.
- **threads** — Thread count. Keep ≤ 16.
