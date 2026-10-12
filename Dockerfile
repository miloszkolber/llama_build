# syntax=docker/dockerfile:1.7
ARG ONEAPI_VERSION=2026.1.1-devel-ubuntu24.04
ARG WHISPER_BACKEND=vulkan
FROM docker.io/golang:1.25-bookworm AS probe-build
WORKDIR /src
COPY probe/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/probe .
WORKDIR /launcher
COPY launcher/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/whisper-entrypoint .

FROM docker.io/intel/oneapi-toolkit:${ONEAPI_VERSION} AS toolchain
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG LEVEL_ZERO_VERSION=1.34.0
ARG COMPUTE_RUNTIME_VERSION=26.35.39758.10
ARG COMPUTE_RUNTIME_FULL=26.35.39758.10-0
ARG IGC_VERSION=2.41.5
ARG IGC_VERSION_FULL=2_2.41.5+22716
ARG IGDGMM_VERSION=22.10.0
RUN apt-get update && apt-get install -y --no-install-recommends \
    git cmake ninja-build python3 wget ca-certificates libssl-dev \
    libavcodec-dev libavformat-dev libavutil-dev libswresample-dev
WORKDIR /tmp/debs
RUN wget -q "https://github.com/oneapi-src/level-zero/releases/download/v${LEVEL_ZERO_VERSION}/libze1_${LEVEL_ZERO_VERSION}+u24.04_amd64.deb" \
    && wget -q "https://github.com/oneapi-src/level-zero/releases/download/v${LEVEL_ZERO_VERSION}/libze-dev_${LEVEL_ZERO_VERSION}+u24.04_amd64.deb" \
    && wget -q "https://github.com/intel/intel-graphics-compiler/releases/download/v${IGC_VERSION}/intel-igc-core-${IGC_VERSION_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/intel-graphics-compiler/releases/download/v${IGC_VERSION}/intel-igc-opencl-${IGC_VERSION_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/intel-ocloc_${COMPUTE_RUNTIME_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/intel-opencl-icd_${COMPUTE_RUNTIME_FULL}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/libigdgmm12_${IGDGMM_VERSION}_amd64.deb" \
    && wget -q "https://github.com/intel/compute-runtime/releases/download/${COMPUTE_RUNTIME_VERSION}/libze-intel-gpu1_${COMPUTE_RUNTIME_FULL}_amd64.deb" \
    && apt-get -o Dpkg::Options::="--force-overwrite" install -y ./*.deb \
    && rm -rf /tmp/debs /var/lib/apt/lists/*
COPY scripts/collect-rootfs.py /collect-rootfs.py

FROM toolchain AS llama-build
ARG LLAMA_SHA=cbe28bb3cb0758f0c566a9a3561766821aeb0748
WORKDIR /src
RUN git init -q && git remote add origin https://github.com/ggml-org/llama.cpp.git \
    && git fetch --depth 1 origin "${LLAMA_SHA}" && git checkout -q FETCH_HEAD \
    && test "$(git rev-parse HEAD)" = "${LLAMA_SHA}"
RUN source /opt/intel/oneapi/setvars.sh \
    && cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx \
      -DGGML_SYCL=ON -DGGML_SYCL_F16=ON -DGGML_NATIVE=OFF \
      -DLLAMA_BUILD_TESTS=OFF -DLLAMA_CURL=OFF \
    && cmake --build build --target llama-server --parallel 2 \
    && python3 /collect-rootfs.py build/bin llama-server /rootfs

FROM toolchain AS whisper-sycl-build
# CI overrides this with the resolved release SHA. Local builds resolve the tag.
ARG WHISPER_REF=v1.9.5
WORKDIR /src
RUN git init -q && git remote add origin https://github.com/ggml-org/whisper.cpp.git \
    && git fetch --depth 1 origin "${WHISPER_REF}" && git checkout -q FETCH_HEAD
RUN source /opt/intel/oneapi/setvars.sh \
    && cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx \
      -DGGML_SYCL=ON -DGGML_SYCL_F16=ON -DGGML_NATIVE=OFF \
      -DWHISPER_BUILD_TESTS=OFF -DWHISPER_BUILD_IS_DEV=OFF \
      -DWHISPER_COMMON_FFMPEG=ON \
    && cmake --build build --target whisper-server --parallel 2 \
    && python3 /collect-rootfs.py build/bin whisper-server /rootfs \
    && mkdir -p /rootfs/app/public && cp -a examples/server/public/. /rootfs/app/public/

FROM docker.io/ubuntu:24.04 AS whisper-vulkan-build
ARG WHISPER_REF=v1.9.5
RUN apt-get update && apt-get install -y --no-install-recommends \
    git cmake ninja-build build-essential python3 ca-certificates \
    libvulkan-dev mesa-vulkan-drivers glslc glslang-tools \
    libavcodec-dev libavformat-dev libavutil-dev libswresample-dev \
    && rm -rf /var/lib/apt/lists/*
COPY scripts/collect-rootfs.py /collect-rootfs.py
WORKDIR /src
RUN git init -q && git remote add origin https://github.com/ggml-org/whisper.cpp.git \
    && git fetch --depth 1 origin "${WHISPER_REF}" && git checkout -q FETCH_HEAD
RUN cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release \
      -DGGML_VULKAN=ON -DGGML_NATIVE=OFF -DWHISPER_BUILD_TESTS=OFF \
      -DWHISPER_BUILD_IS_DEV=OFF -DWHISPER_COMMON_FFMPEG=ON \
    && cmake --build build --target whisper-server --parallel 2 \
    && python3 /collect-rootfs.py build/bin whisper-server /rootfs vulkan \
    && mkdir -p /rootfs/app/public && cp -a examples/server/public/. /rootfs/app/public/

FROM whisper-${WHISPER_BACKEND}-build AS whisper-selected

FROM scratch AS llama_runtime
ARG LLAMA_SHA=cbe28bb3cb0758f0c566a9a3561766821aeb0748
LABEL org.opencontainers.image.source="https://github.com/ggml-org/llama.cpp" \
      org.opencontainers.image.revision="${LLAMA_SHA}"
COPY --from=llama-build /rootfs/ /
COPY --from=probe-build /out/probe /app/probe
ENV HOME=/home/runtime LD_LIBRARY_PATH=/app/lib:/opt/intel/oneapi/compiler/latest/lib:/opt/intel/oneapi/mkl/latest/lib:/opt/intel/oneapi/mkl/latest/lib/intel64:/opt/intel/oneapi/tbb/latest/lib \
    ONEAPI_DEVICE_SELECTOR=level_zero:0
USER 1000:1000
WORKDIR /app
ENTRYPOINT ["/app/llama-server"]

FROM scratch AS whisper_runtime
ARG WHISPER_REF=v1.9.5
LABEL org.opencontainers.image.source="https://github.com/ggml-org/whisper.cpp" \
      org.opencontainers.image.revision="${WHISPER_REF}"
COPY --from=whisper-selected /rootfs/ /
COPY --from=probe-build /out/probe /app/probe
COPY --from=probe-build /out/whisper-entrypoint /app/whisper-entrypoint
ENV HOME=/home/runtime LD_LIBRARY_PATH=/app/lib:/opt/intel/oneapi/compiler/latest/lib:/opt/intel/oneapi/mkl/latest/lib:/opt/intel/oneapi/mkl/latest/lib/intel64:/opt/intel/oneapi/tbb/latest/lib \
    ONEAPI_DEVICE_SELECTOR=level_zero:0
USER 1000:1000
WORKDIR /app
ENTRYPOINT ["/app/whisper-entrypoint"]
