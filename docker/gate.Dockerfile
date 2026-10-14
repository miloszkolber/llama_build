# syntax=docker/dockerfile:1
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/gate/ cmd/gate/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /gate ./cmd/gate
FROM gcr.io/distroless/static-debian12:nonroot
ARG SOURCE_SHA
LABEL org.opencontainers.image.revision="$SOURCE_SHA"
COPY --from=build /gate /app/gate
USER 1000:1000
ENTRYPOINT ["/app/gate"]
HEALTHCHECK CMD ["/app/gate", "--health"]
