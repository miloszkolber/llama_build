# Evidence

## Repository checks

- `gofmt` produced no remaining formatting changes.
- `go test ./...` and `go vet ./...` passed in the development container. The gate suite includes the new discovery integration test with a real HTTP runtime and the existing persistent-proxy, handoff, lifecycle and race tests.
- `check-docs.mjs` and strict OpenSpec validation passed.

## Behaviour exercised

- A cold plaintext discovery request started the fake runtime, waited for readiness and returned the runtime response.
- After the idle stop, an identical request returned the same response from the cache and the runtime was not started.
- A non-discovery request started the runtime again.
- HEAD returned headers without the cached body.
- An unknown host or a missing cache skipped the cached path.
- The inference profile started its llama and whisper listeners from the same image with discovery settings present, proving the settings are ignored outside port handoff.

## Publication

- Gate-only workflow run `40127780253` from revision `e5d62a0` passed the check, build, final-image smoke check and publication.
- Published gate identity: `sha256:803d36eda9789db2ba88f0420222d76e71d232e89f55ab92b5f742981c4dbfcc`, promoted as `latest`.

## Consuming deployment check

- The streaming deployment set `GATE_DISCOVERY_PORT` and `GATE_DISCOVERY_PATHS` and recreated its gate.
- A discovery probe populated the cache by starting the runtime. After the runtime was stopped, the same probe returned an identical response while the runtime stayed stopped.
- A discovery request over the TLS handoff port still started the runtime.
- The deployment's gate health endpoint and idle listener reacquisition continued to work.

## Gaps

- The cache is in memory per gate process. A gate restart clears it, so the next discovery probe starts the runtime once.
- Cache keys use the request host. A client using a different hostname or address than the one that populated the cache starts the runtime once for its own key.
- The application workflow and runtime image checks remain separate; no GPU, transcription or streaming session was exercised by this change.
