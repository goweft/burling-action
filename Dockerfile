# syntax=docker/dockerfile:1

# Build stage: compile burling (from the published module, pinned by tag)
# and the action entrypoint wrapper (from this repo). Both are static,
# CGO-disabled binaries.
FROM golang:1.22-alpine AS build
ARG BURLING_VERSION=v0.2.0
ENV CGO_ENABLED=0 GOFLAGS=-trimpath

# burling: the validator binary, fetched from the public Go module. The
# version is stamped into the binary (main.Version) so its SARIF output
# reports the pinned release rather than the "dev" default.
RUN go install -ldflags="-X main.Version=${BURLING_VERSION}" "github.com/goweft/burling/cmd/burling@${BURLING_VERSION}"

# entrypoint: the thin action wrapper, built from local source. The
# wrapper has no third-party dependencies, so go.mod alone suffices.
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN go build -o /out/entrypoint .

# Runtime stage: distroless static. Carries only the two static binaries
# and a CA bundle -- no shell, no package manager. Runs as root so the
# entrypoint can write the SARIF file into the mounted workspace, which
# GitHub owns as root for container actions.
FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/burling /usr/local/bin/burling
COPY --from=build /out/entrypoint /usr/local/bin/entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint"]
