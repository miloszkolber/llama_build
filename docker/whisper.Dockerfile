# syntax=docker/dockerfile:1
FROM golang:1.27-bookworm AS helper
WORKDIR /src
COPY go.mod ./
COPY cmd/probe/ cmd/probe/
COPY cmd/whisper/ cmd/whisper/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /probe ./cmd/probe \
    && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /whisper-entrypoint ./cmd/whisper

FROM ubuntu:26.04 AS build
ARG SOURCE_SHA
ARG BUILD_DATE
RUN apt-get update && apt-get install -y --no-install-recommends \
    git cmake ninja-build build-essential python3 ca-certificates \
    libvulkan-dev mesa-vulkan-drivers glslc glslang-tools spirv-headers \
    libavcodec-dev libavformat-dev libavutil-dev libswresample-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
RUN git init -q && git remote add origin https://github.com/ggml-org/whisper.cpp.git \
    && git fetch --depth 1 origin "$SOURCE_SHA" && git checkout -q FETCH_HEAD \
    && test "$(git rev-parse HEAD)" = "$SOURCE_SHA"
RUN cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release \
      -DGGML_VULKAN=ON -DGGML_NATIVE=OFF -DWHISPER_BUILD_TESTS=OFF \
      -DWHISPER_BUILD_IS_DEV=OFF -DWHISPER_COMMON_FFMPEG=ON \
    && cmake --build build --target whisper-server --parallel 2
COPY scripts/collect-rootfs.py /collect-rootfs.py
RUN python3 /collect-rootfs.py build/bin whisper-server /rootfs vulkan

FROM scratch
ARG SOURCE_SHA
LABEL org.opencontainers.image.revision="$SOURCE_SHA"
COPY --from=build /rootfs/ /
COPY --from=helper /probe /app/probe
COPY --from=helper /whisper-entrypoint /app/whisper-entrypoint
ENV HOME=/home/runtime LD_LIBRARY_PATH=/app
USER 1000:1000
WORKDIR /app
ENTRYPOINT ["/app/whisper-entrypoint"]
HEALTHCHECK CMD ["/app/probe", "whisper"]
