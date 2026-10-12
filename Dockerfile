# syntax=docker/dockerfile:1.7
# llama_edge all-in-one: llama.cpp (SYCL) + whisper.cpp (SYCL) + gate.
# Design and open risks: docs/architecture.md.
# Every upstream input is a pinned ARG. CI fail-fasts on the deb URLs
# before the long toolchain stages run.

ARG ONEAPI_VERSION=2026.1.1-devel-ubuntu24.04
ARG LLAMA_SHA=e937b17a2
# Pin to a verified SHA after the first green whisper SYCL build.
ARG WHISPER_REF=master
ARG LEVEL_ZERO_VERSION=1.34.0
ARG COMPUTE_RUNTIME_VERSION=26.35.39758.10
ARG COMPUTE_RUNTIME_FULL=26.35.39758.10-0
ARG IGC_VERSION=v2.41.5
ARG IGC_VERSION_FULL=2_2.41.5+22716
ARG IGDGMM_VERSION=22.10.0
# Local gate image. Publish it to a registry once, then override with
#   --build-context gate-image=ghcr.io/<owner>/gate:latest
# in CI. See docs/architecture.md R4.
ARG GATE_IMAGE=core/gate

# Pinned driver debs, downloaded once and shared by later stages.
FROM docker.io/ubuntu:24.04 AS debs
ARG LEVEL_ZERO_VERSION COMPUTE_RUNTIME_VERSION COMPUTE_RUNTIME_FULL
ARG IGC_VERSION IGC_VERSION_FULL IGDGMM_VERSION
WORKDIR /debs
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists \
    && wget -q "https://github.com/oneapi-src/level-zero/releases/download/v${LEVEL_ZERO_VERSION}/libze1_${LEVEL_ZERO_VERSION}+u24.04_amd64.deb" \
    && wget -q "https://github.com/oneapi-src/level-zero/releases/download/v${LEVEL_ZERO_VERSION}/libze-dev_${LEVEL_ZERO_VERSION}+u24.04_amd64.deb" \
    && wget -q "https://github.com/intel/intel-graphics-compiler/releases/download/${IGC_VERSION}/intel-igc-core-${IGC_VERSION_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/intel-graphics-compiler/releases/download/${IGC_VERSION}/intel-igc-opencl-${IGC_VERSION_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/intel-ocloc_${COMPUTE_RUNTIME_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/intel-opencl-icd_${COMPUTE_RUNTIME_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/libigdgmm12_${IGDGMM_VERSION}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/libze-intel-gpu1_${COMPUTE_RUNTIME_FULL}_amd64.deb"

# Toolchain plus pinned driver stack for both builds.
FROM docker.io/intel/oneapi-toolkit:${ONEAPI_VERSION} AS toolchain
COPY --from=debs /debs/*.deb /tmp/debs/
RUN apt-get update \
    && apt-get install -y git libssl-dev \
    && dpkg --install /tmp/debs/*.deb \
    && rm -rf /tmp/debs /var/lib/apt/lists

FROM toolchain AS llama-build
ARG LLAMA_SHA
WORKDIR /src/llama.cpp
RUN git init -q . && git remote add origin https://github.com/ggml-org/llama.cpp.git \
    && git fetch -q --depth 1 origin "${LLAMA_SHA}" && git checkout -q FETCH_HEAD
RUN cmake -B build -DGGML_NATIVE=OFF -DGGML_SYCL=ON \
        -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx \
        -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON \
        -DGGML_SYCL_F16=ON -DLLAMA_BUILD_TESTS=OFF \
    && cmake --build build --config Release -j$(nproc) --target llama-server
RUN mkdir -p /out/lib && cp build/bin/llama-server /out/ \
    && find build -name "*.so*" -exec cp -P {} /out/lib \;

FROM toolchain AS whisper-build
ARG WHISPER_REF
WORKDIR /src/whisper.cpp
RUN git init -q . && git remote add origin https://github.com/ggml-org/whisper.cpp.git \
    && git fetch -q --depth 1 origin "${WHISPER_REF}" && git checkout -q FETCH_HEAD
RUN cmake -B build -DWHISPER_SYCL=ON -DWHISPER_SYCL_F16=ON \
        -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx \
    && cmake --build build --config Release -j$(nproc) --target whisper-server
RUN mkdir -p /out/lib && cp build/bin/whisper-server /out/ \
    && find build -name "*.so*" -exec cp -P {} /out/lib \;

# Static gate binary from the existing local image. No source duplication.
FROM ${GATE_IMAGE} AS gate-src

# Debian stage: install the driver debs, collect exactly the runtime libs
# the two servers load, plus ffmpeg for whisper audio conversion.
FROM docker.io/ubuntu:24.04 AS rootfs
COPY --from=debs /debs/*.deb /tmp/debs/
COPY --from=llama-build /out/ /pkg/llama/
COPY --from=whisper-build /out/ /pkg/whisper/
RUN apt-get update \
    && apt-get install -y --no-install-recommends libgomp1 ffmpeg binutils \
    && dpkg --install /tmp/debs/*.deb \
    && rm -rf /tmp/debs /var/lib/apt/lists \
    && mkdir -p /rootfs/app /rootfs/app/lib \
    && cp /pkg/llama/llama-server /pkg/whisper/whisper-server /rootfs/app/ \
    && cp -P /pkg/llama/lib/*.so* /pkg/whisper/lib/*.so* /rootfs/app/lib/ 2>/dev/null; true \
    && for bin in /rootfs/app/llama-server /rootfs/app/whisper-server; do \
           ldd "$bin" | awk '{print $3}' | grep -E '^/' ; \
       done | sort -u | while read -r lib; do \
           mkdir -p "/rootfs$(dirname "$lib")" && cp -P "$lib" "/rootfs$lib"; \
       done \
    && for pattern in 'libze*' 'libig*' 'libocloc*' 'libsycl*' 'libur_*' 'libsvml*' 'libirng*' 'libimf*'; do \
           find /usr/lib/x86_64-linux-gnu /opt/intel -name "$pattern" \
               -exec sh -c 'mkdir -p "/rootfs$(dirname "$1")" && cp -P "$1" "/rootfs$1"' _ {} \; 2>/dev/null; \
       done; true

# Final stage: distroless C++ runtime. No shell, no package manager.
FROM gcr.io/distroless/cc-debian12:nonroot AS runtime
COPY --from=rootfs /rootfs/ /
COPY --from=gate-src /app/gate /app/gate
ENV LD_LIBRARY_PATH=/app/lib
WORKDIR /app
# Phase 1 roles override the entrypoint per service (see compose/).
# The gate role keeps its self healthcheck; server roles run without one
# because the gate performs HTTP readiness waits itself.
ENTRYPOINT ["/app/gate"]
