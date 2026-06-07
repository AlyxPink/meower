# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS build-web
ENV CGO_ENABLED=0
WORKDIR /src/web

# Download deps. The web module replaces ../api, ../pkg/urls and
# ../pkg/observability, so all three sub-modules must be present for resolution.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=./api/,target=/src/api/ \
    --mount=type=bind,source=./pkg/urls/,target=/src/pkg/urls/ \
    --mount=type=bind,source=./pkg/observability/,target=/src/pkg/observability/ \
    --mount=type=bind,source=./web/go.sum,target=./go.sum \
    --mount=type=bind,source=./web/go.mod,target=./go.mod \
    go mod download -x

# Code-generation toolchain. The web imports the api's protobuf packages, so
# the build generates that Go code from the .proto sources before compiling.
COPY --from=ghcr.io/a-h/templ:v0.3.924 /ko-app/templ /usr/local/bin/templ
COPY --from=bufbuild/buf:1 /usr/local/bin/buf /usr/local/bin/buf
RUN apk add --no-cache protoc protobuf-dev
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Build the application. The api/ and scripts/ mounts are read-write so the
# generated protobuf .go files land alongside the api sources for the web build;
# they're discarded with the ephemeral mounts.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=rw,type=bind,source=./web/,target=/src/web/ \
    --mount=rw,type=bind,source=./api/,target=/src/api/ \
    --mount=rw,type=bind,source=./scripts/,target=/src/scripts/ \
    --mount=type=bind,source=./pkg/urls/,target=/src/pkg/urls/ \
    --mount=type=bind,source=./pkg/observability/,target=/src/pkg/observability/ \
    cd /src && PATH="$PATH:$(go env GOPATH)/bin" ./scripts/generate_protobuf.sh && \
    cd /src/web && templ generate -path /src/web/views && \
    go build -o /bin/web-server /src/web/

FROM oven/bun:alpine AS assets
ARG TARGETOS
ARG TARGETARCH
ARG TAILWIND_VERSION=3.4.17
WORKDIR /src/web/
# Adding a conditional statement to replace amd64 with x64 to match their release format
RUN if [ "${TARGETARCH}" = "amd64" ]; then \
    TARGETARCH="x64"; \
    fi && \
    wget https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-${TARGETOS}-${TARGETARCH} && \
    chmod +x tailwindcss-${TARGETOS}-${TARGETARCH} && \
    mv tailwindcss-${TARGETOS}-${TARGETARCH} /usr/local/bin/tailwindcss

RUN --mount=type=cache,target=/tmp/bun/cache \
    --mount=type=bind,source=./web/package.json,target=package.json \
    --mount=type=bind,source=./web/bun.lockb,target=bun.lockb \
    bun install \
    --cache-dir /tmp/bun/cache \
    --no-save \
    --frozen-lockfile \
    --production

RUN --mount=rw,type=bind,source=./web/,target=/usr/src/tailwind/ \
    /usr/local/bin/tailwindcss \
    -i /usr/src/tailwind/static/src/css/main.css \
    -o /opt/tailwind/main.css \
    --content /usr/src/tailwind/ui/**/*.templ \
    --minify

FROM alpine:3.21 AS production-web

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
COPY --chown=meower:meower --from=build-web /bin/web-server /opt/meower/
# Copy the production CSS files from the "assets" stage.
COPY --from=assets /opt/tailwind/main.css /opt/meower/main.css
# Copy JavaScript files
COPY --chown=meower:meower ./static/public/js/ /opt/meower/static/js/
# Expose the port that the application listens on.
EXPOSE 3000
WORKDIR /opt/meower/
# What the container should run when it is started.
ENTRYPOINT [ "/opt/meower/web-server" ]

################################################################################
# Create a development image that includes what's required to generate the protoc and CSS files
FROM build-web AS development-web

# TODO: Use versioned modules
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go install github.com/bokwoon95/wgo@v0.5.13

# Copy the executables from the "build" stage.
COPY --chown=meower:meower --from=build-web /bin/web-server /opt/meower/

# Copy tailwind CLI binary from the "assets" stage.
COPY --from=assets /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss

# Copy rustywind binary from the rustywind image
COPY --from=avencera/rustywind:latest /rustywind /usr/local/bin/rustywind

# Set the default target to be the final stage.
FROM production-web
