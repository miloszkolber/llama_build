# llama_edge

All-in-one inference image for Intel GPUs: `llama.cpp` (SYCL) plus
`whisper.cpp` (SYCL) plus `gate`, built from pinned sources with a current
Intel driver stack, shipped on a distroless runtime stage.

Status: draft. Nothing here runs production traffic yet. Each integration
carries a verification gate in `docs/architecture.md`. No benchmark runs
while production inference is active.

## Layout

| Path | Content |
| --- | --- |
| `Dockerfile` | Multi-stage all-in-one build |
| `.github/workflows/build.yml` | Pinned build plus scheduled refresh to GHCR |
| `.github/workflows/verify.yml` | Smoke plus benchmark gate before promotion |
| `docs/architecture.md` | Phase plan, gate integration, healthcheck rework, risks |
| `compose/edge.compose.yml` | Migration example with placeholder variables |
| `compose/.env.example` | Values the deployment host provides |

## Quickstart

```sh
cp compose/.env.example compose/.env
# Fill in model, device and registry values for this host.
docker compose -f compose/edge.compose.yml config --quiet
```

Local full build (needs the oneAPI toolchain download, tens of minutes):

```sh
docker buildx build -f Dockerfile \
  --build-context gate-image=core/gate \
  --build-arg LLAMA_SHA=<pinned> --build-arg WHISPER_REF=<pinned> .
```

## Version policy

Every upstream input is a pinned `ARG`: `llama.cpp` and `whisper.cpp`
commits, oneAPI toolkit, Level Zero, compute runtime and IGC. The weekly
schedule rebuilds with the pinned values to catch silent drift; bumps are
deliberate commits that must pass `verify.yml` before the `stable` tag
moves. Rolling latest is a non-goal.

## Why this exists

The upstream `server-intel` image trails the Intel driver stack (it
bundles compute runtime 26.18 and IGC 2.34.4 while 26.35 and IGC 2.41.5
are available), and its final stage ships the full 16 GB compiler
toolkit. A driver-only rebuild once measured +43% MoE generation on this
class of hardware (upstream issue 24045). This repo exists to capture
that class of win repeatably, minus the 16 GB.
