# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26

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
COPY --from=ghcr.io/a-h/templ:v0.3.1020 /ko-app/templ /usr/local/bin/templ
COPY --from=bufbuild/buf:1 /usr/local/bin/buf /usr/local/bin/buf
RUN apk add --no-cache protoc protobuf-dev
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2

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

# Tailwind v4 builds via the @tailwindcss/cli package (run with bunx) instead of
# the standalone binary used in v3. v4 has no --content flag; class detection is
# driven by the `@source` directives in static/src/css/main.css (which point at
# the .templ views), so the views must be present in the build context.
FROM oven/bun:alpine AS assets
ARG TAILWIND_VERSION=4.3.0
WORKDIR /src/web/

RUN --mount=type=cache,target=/tmp/bun/cache \
    --mount=type=bind,source=./web/package.json,target=package.json \
    --mount=type=bind,source=./web/bun.lock,target=bun.lock \
    bun install \
    --cache-dir /tmp/bun/cache \
    --no-save \
    --frozen-lockfile \
    --production

# Source files needed for the CSS build: the input stylesheet and the .templ
# views/components that `@source` scans for utility classes.
COPY ./web/static ./static
COPY ./web/views ./views

RUN --mount=type=cache,target=/tmp/bun/cache \
    bunx @tailwindcss/cli@${TAILWIND_VERSION} \
    -i ./static/src/css/main.css \
    -o /opt/tailwind/main.css \
    --minify

FROM alpine:3.23 AS production-web

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
# Development image for the asset pipeline (CSS + JS). Bun-based, with the web
# package's node_modules pre-installed so the compose `tailwind` and `javascript`
# watcher services run `bunx @tailwindcss/cli` / `bun build` without a per-start
# install. Source files are bind-mounted by compose; node_modules comes from
# this image layer.
FROM oven/bun:alpine AS development-assets
WORKDIR /src/web/
COPY ./web/package.json ./web/bun.lock ./
RUN bun install --frozen-lockfile

################################################################################
# Create a development image that includes what's required to generate the protoc and CSS files
FROM build-web AS development-web

# TODO: Use versioned modules
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go install github.com/bokwoon95/wgo@v0.6.4 && \
    go install mvdan.cc/gofumpt@v0.10.0

# Copy the executables from the "build" stage.
COPY --chown=meower:meower --from=build-web /bin/web-server /opt/meower/

# Copy rustywind binary from the rustywind image (used by the compose `formatter`
# service alongside gofumpt + templ fmt).
COPY --from=avencera/rustywind:latest /rustywind /usr/local/bin/rustywind

# Set the default target to be the final stage.
FROM production-web
