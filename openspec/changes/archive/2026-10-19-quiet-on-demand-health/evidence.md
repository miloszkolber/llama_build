# Execution evidence

## Measurement before the change

`llama-runtime` while running, on the deployment (dockerd 29.1.3):

```
health=healthy failing=0      probes every 60 s, exit 0
```

`llama-runtime` immediately after `docker stop`, which is the same clean stop the gate performs when it idles a runtime:

```
state=exited exit=0 health=unhealthy failing=0
last probes: 15:04:33 exit=0 / 15:05:33 exit=0 / 15:06:30 exit=0    # no failing probe logged
```

Control container with a trivial check (`alpine`, `--health-cmd true`, 1–2 s interval):

```
while running:                       healthy
after self-exit with exit code 0:    unhealthy
after an external stop:              unhealthy
```

Docker therefore reports `unhealthy` for any stopped container that defines a healthcheck, whatever its probes said. An on-demand runtime, stopped most of the time, always looks failed.

## Why nothing depended on it here

`cmd/gate/main.go` asks Docker for container health only on the port-handoff path (`s.config.PortHandoff.RuntimeName`); the persistent-proxy path waits by writing `GET <health_path>` to the runtime and reading the answer. The Whisper and llama runtimes therefore need no container health status, while `wolf-runtime` must keep its image check because port handoff requires Docker health before it will forward.

## Checks

Both trees validate and report the intended configuration:

```
docker compose --profile runtime config --quiet     # deployment example and host deployment
llama-runtime / whisper-runtime healthcheck: {'disable': True}
```

Recreated containers, stopped and running:

```
whisper-runtime: state=exited   exit=0  health=none  hc={"Test":["NONE"]}
llama-runtime:   state=running          health=none  hc={"Test":["NONE"]}
wolf-runtime:    state=exited           health=unhealthy  hc=<image check>   # unchanged control
```

Acceptance through the gate with the check disabled: a bounded `d1-3b` completion returned HTTP 200 with `finish_reason=stop` and content `42`, and the deployment check transcribed its pinned fixture (`Whisper transcription check passed.`).

Repository checks: `bun scripts/check-docs.mjs`, `bun test scripts/check-docs.test.mjs`, `openspec validate --all --strict`, `go test ./... && go vet ./...`, `python3 -m unittest discover -s tests`, `actionlint`, and the Compose syntax check all pass.

## Residuals and an incident

- Recreating the two runtime containers by hand while they were still running interrupted two in-flight Magpie video-description requests: attempt 1 failed with an interrupted SSE stream, attempt 2 with a client deadline and one `503 Runtime is unavailable` while the runtime came back cold. The phase resumed and is processing normally on the recreated containers. The daily job checks for a running container immediately before recreating for exactly this reason; a manual recreate must do the same.
- `wolf-runtime` keeps its check, so a stopped Wolf runtime still reports `unhealthy`. That is inherent to a healthcheck on a deliberately stopped container, and the port-handoff gate needs the status while starting it.
- The gate's own `--health` remains a binary-execution smoke check; that is documented and unchanged.
- No image input changed, so no image was rebuilt for this change.
