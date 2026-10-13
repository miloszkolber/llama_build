# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS helper
WORKDIR /src
COPY go.mod ./
COPY cmd/probe/ cmd/probe/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /probe ./cmd/probe

FROM intel/oneapi-toolkit:2026.1.1-devel-ubuntu24.04 AS compile
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG SOURCE_SHA
ARG BUILD_DATE
RUN apt-get update && apt-get install -y --no-install-recommends \
    git cmake ninja-build python3 ca-certificates libssl-dev ocl-icd-libopencl1
COPY scripts/install-drivers.py /install-drivers.py
COPY .build/drivers.json /drivers.json
# Install the matched vendor stack even if the toolkit carries a newer GMM revision.
RUN python3 /install-drivers.py /drivers.json \
    && apt-get -o Dpkg::Options::="--force-overwrite" install -y --allow-downgrades /tmp/debs/*.deb \
    && rm -rf /tmp/debs /var/lib/apt/lists/*
WORKDIR /src
RUN git init -q && git remote add origin https://github.com/ggml-org/llama.cpp.git \
    && git fetch --depth 1 origin "$SOURCE_SHA" && git checkout -q FETCH_HEAD \
    && test "$(git rev-parse HEAD)" = "$SOURCE_SHA"
RUN source /opt/intel/oneapi/setvars.sh --force \
    && cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx \
      -DGGML_SYCL=ON -DGGML_SYCL_F16=ON -DGGML_NATIVE=OFF \
      -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON \
      -DLLAMA_BUILD_TESTS=OFF -DLLAMA_CURL=OFF \
      -DLLAMA_BUILD_UI=OFF -DLLAMA_USE_PREBUILT_UI=OFF \
    && cmake --build build --target llama-server --parallel 1

FROM compile AS build
COPY scripts/collect-rootfs.py /collect-rootfs.py
RUN source /opt/intel/oneapi/setvars.sh --force \
    && python3 /collect-rootfs.py build/bin llama-server /rootfs

FROM scratch
ARG SOURCE_SHA
LABEL org.opencontainers.image.revision="$SOURCE_SHA"
COPY --from=build /rootfs/ /
COPY --from=build /drivers.json /licenses/drivers.json
COPY --from=helper /probe /app/probe
ENV HOME=/home/runtime \
    LD_LIBRARY_PATH=/app:/opt/intel/oneapi/compiler/latest/lib:/opt/intel/oneapi/mkl/latest/lib:/opt/intel/oneapi/mkl/latest/lib/intel64:/opt/intel/oneapi/tbb/latest/lib \
    ONEAPI_DEVICE_SELECTOR=level_zero:0 \
    LLAMA_ARG_HOST=127.0.0.1 LLAMA_ARG_PORT=9668
USER 1000:1000
WORKDIR /app
ENTRYPOINT ["/app/llama-server"]
HEALTHCHECK CMD ["/app/probe", "llama"]
