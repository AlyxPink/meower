# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26
ARG SQLC_VERSION=1.31.1

# "COPY --from" does not expand ARG variables, so we need to use a temporary image to get the SQLC version
FROM sqlc/sqlc:${SQLC_VERSION} AS sqlc-version

FROM golang:${GO_VERSION}-alpine AS build-api
ENV CGO_ENABLED=0
WORKDIR /src/api

# Code-generation toolchain — the template ships .proto and .sql sources, so the
# build stage generates the Go code (protobuf + sqlc) before compiling.
COPY --from=sqlc-version /workspace/sqlc /usr/local/bin/sqlc
COPY --from=bufbuild/buf:1 /usr/local/bin/buf /usr/local/bin/buf
RUN apk add --no-cache protoc protobuf-dev

# Download deps. The api module has `replace => ../pkg/observability`, so that
# sub-module must be present (at /src/pkg/observability) for resolution. go.mod
# is mounted read-write because `go mod download` may reconcile the module graph
# (the replace'd local module) and rewrite it; the change is discarded with the
# ephemeral mount.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=rw,type=bind,source=./api/go.sum,target=./go.sum \
    --mount=rw,type=bind,source=./api/go.mod,target=./go.mod \
    --mount=type=bind,source=./pkg/observability/,target=/src/pkg/observability/ \
    go mod download -x

# Install the protobuf Go plugins (pinned), then generate protobuf + sqlc code
# and build. The bind mount is read-write so generated .go files land alongside
# the sources for the build; they're discarded with the ephemeral mount.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=rw,type=bind,source=./api/,target=/src/api/ \
    --mount=type=bind,source=./pkg/observability/,target=/src/pkg/observability/ \
    --mount=rw,type=bind,source=./scripts/,target=/src/scripts/ \
    cd /src && PATH="$PATH:$(go env GOPATH)/bin" ./scripts/generate_protobuf.sh && \
    cd /src/api && sqlc generate -f /src/api/db/sqlc.yaml && \
    go build -o /bin/api-server ./

FROM alpine:3.23 AS production-api

# Install any runtime dependencies that are needed to run your application.
# Leverage a cache mount to /var/cache/apk/ to speed up subsequent builds.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
    ca-certificates \
    tzdata \
    && \
    update-ca-certificates

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    meower
USER meower

# Copy the executable from the "build" stage.
COPY --chown=meower:meower --from=build-api /bin/api-server /opt/meower/
# Expose the port that the application listens on.
EXPOSE 50051
WORKDIR /opt/meower/
# What the container should run when it is started.
ENTRYPOINT [ "/opt/meower/api-server" ]

################################################################################
# Create a development image that includes what's required to generate the protoc and CSS files
FROM build-api AS development-api
RUN apk add --no-cache \
    protobuf-dev

# TODO: Use versioned modules
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2 && \
    go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.52 && \
    go install github.com/bokwoon95/wgo@v0.6.4

# Copy the executables from the "build" stage.
COPY --chown=meower:meower --from=build-api /bin/api-server /opt/meower/

# Copy buf binary from the buf image
COPY --from=bufbuild/buf:1 /usr/local/bin/buf /usr/local/bin/buf

# Set the default target to be the production-api stage.
FROM production-api
