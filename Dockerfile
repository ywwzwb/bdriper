FROM ubuntu:24.04 AS base
RUN apt-get update && apt-get install -y \
    ffmpeg \
    x264 \
    x265 \
    mkvtoolnix \
    vainfo \
    intel-media-va-driver \
    mesa-va-drivers \
    ca-certificates \
    curl \
    locales \
    && rm -rf /var/lib/apt/lists/* \
    && locale-gen en_US.UTF-8

FROM golang:1.24-bookworm AS go-builder
WORKDIR /src
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bdriper ./cmd/server/

FROM node:20 AS web-builder
COPY web /src
WORKDIR /src
RUN npm ci && npm run build

FROM base AS runtime
COPY --from=go-builder /bdriper /usr/local/bin/bdriper
COPY --from=web-builder /src/dist /app/web/dist
COPY presets/ /app/presets/
COPY docs/help/ /app/docs/help/
EXPOSE 8080
ENV DATA_DIR=/data
ENV INPUT_DIR=/input
ENV OUTPUT_DIR=/output
ENV PRESETS_DIR=/app/presets
ENV HELP_DIR=/app/docs/help
ENV WEB_DIST=/app/web/dist
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
RUN mkdir -p /tmp/bdriper
ENTRYPOINT ["bdriper"]
