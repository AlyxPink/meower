# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS build-web
ENV CGO_ENABLED=0
WORKDIR /src/web

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=./api/,target=/src/api/ \
    --mount=type=bind,source=./web/go.sum,target=./go.sum \
    --mount=type=bind,source=./web/go.mod,target=./go.mod \
    go mod download -x

# Copy a-h/templ binary from the templ image
COPY --from=ghcr.io/a-h/templ:v0.2.747 /ko-app/templ /usr/local/bin/templ

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=rw,type=bind,source=./web/,target=/src/web/ \
    --mount=type=bind,source=./api/,target=/src/api/ \
    templ generate -path /src/web/views && \
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

FROM alpine:3 AS production-web

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
    go install github.com/bokwoon95/wgo@v0.5.11

# Copy the executables from the "build" stage.
COPY --chown=meower:meower --from=build-web /bin/web-server /opt/meower/

# Copy tailwind CLI binary from the "assets" stage.
COPY --from=assets /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss
# Copy the production node_modules from the "assets" stage.
COPY --from=assets /src/web/node_modules node_modules

# Set the default target to be the final stage.
FROM production-web
