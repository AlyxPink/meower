# syntax=docker/dockerfile:1

ARG GO_VERSION=1.23
ARG SQLC_VERSION=1.27.0

# "COPY --from" does not expand ARG variables, so we need to use a temporary image to get the SQLC version
FROM sqlc/sqlc:${SQLC_VERSION} AS sqlc-version

FROM golang:${GO_VERSION}-alpine AS build-api
ENV CGO_ENABLED=0
WORKDIR /src/api

# Copy sqlc binary from the sqlc image
COPY --from=sqlc-version /workspace/sqlc /usr/local/bin/sqlc

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=./api/go.sum,target=./go.sum \
    --mount=type=bind,source=./api/go.mod,target=./go.mod \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=rw,type=bind,source=./api/,target=/src/api/ \
    sqlc generate -f /src/api/db/sqlc.yaml && \
    go build -o /bin/api-server ./

FROM alpine:latest AS production-api

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
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-health-probe@latest && \
    go install github.com/bokwoon95/wgo@latest

# Copy the executables from the "build" stage.
COPY --chown=meower:meower --from=build-api /bin/api-server /opt/meower/

# Copy buf binary from the buf image
COPY --from=bufbuild/buf:1 /usr/local/bin/buf /usr/local/bin/buf

# Set the default target to be the production-api stage.
FROM production-api
