# syntax=docker/dockerfile:1.7
# Context: the separate gate source repo (server/go.mod, server/main.go).
FROM docker.io/golang:1.25-bookworm AS build
WORKDIR /src
COPY server/go.mod server/main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/gate ./main.go
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gate /app/gate
USER 1000:1000
WORKDIR /app
HEALTHCHECK --interval=60s --timeout=5s --start-period=10s --retries=3 CMD ["/app/gate", "--health"]
ENTRYPOINT ["/app/gate"]
