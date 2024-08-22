# syntax=docker/dockerfile:1

# Comments are provided throughout this file to help you get started.
# If you need more help, visit the Dockerfile reference guide at
# https://docs.docker.com/engine/reference/builder/

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=1.22
FROM golang:${GO_VERSION}-alpine AS build
ENV CGO_ENABLED=0
WORKDIR /src

# Download dependencies as a separate step to take advantage of Docker's caching.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage bind mounts to go.sum and go.mod to avoid having to copy them into
# the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Copy sqlc binary from the sqlc image
COPY --from=sqlc/sqlc:1.27.0 /workspace/sqlc /usr/local/bin/sqlc

# Copy a-h/templ binary from the templ image
COPY --from=ghcr.io/a-h/templ:v0.2.747 /ko-app/templ /usr/local/bin/templ

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=rw,type=bind,target=. \
    sqlc generate -f ./internal/db/sqlc.yaml && \
    templ generate -path ./internal/web/views && \
    go build -o /bin/api-server ./cmd/api-server/ && \
    go build -o /bin/web-server ./cmd/web-server

################################################################################
# Create a stage for building the application web assets.
FROM oven/bun:alpine AS assets
ARG TARGETOS
ARG TARGETARCH
WORKDIR /usr/src/
# Adding a conditional statement to replace amd64 with x64 to match their release format
RUN if [ "${TARGETARCH}" = "amd64" ]; then \
    TARGETARCH="x64"; \
    fi && \
    wget https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-${TARGETOS}-${TARGETARCH} && \
    chmod +x tailwindcss-${TARGETOS}-${TARGETARCH} && \
    mv tailwindcss-${TARGETOS}-${TARGETARCH} /usr/local/bin/tailwindcss

RUN --mount=type=cache,target=/tmp/bun/cache \
    --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=bun.lockb,target=bun.lockb \
    bun install \
    --cache-dir /tmp/bun/cache \
    --no-save \
    --frozen-lockfile \
    --production

RUN --mount=rw,type=bind,source=./internal/web/,target=/usr/src/tailwind/ \
    /usr/local/bin/tailwindcss \
    -i /usr/src/tailwind/static/src/css/main.css \
    -o /usr/src/main.css \
    --content /usr/src/tailwind/ui/**/*.templ \
    --minify

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application. This often uses a different base
# image from the build stage where the necessary files are copied from the build
# stage.
#
# The example below uses the alpine image as the foundation for running the app.
# By specifying the "latest" tag, it will also use whatever happens to be the
# most recent version of that image when you build your Dockerfile. If
# reproducability is important, consider using a versioned tag
# (e.g., alpine:3.17.2) or SHA (e.g., alpine:sha256:c41ab5c992deb4fe7e5da09f67a8804a46bd0592bfdf0b1847dde0e0889d2bff).
FROM alpine:latest AS final

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

################################################################################
# Create a stage to finalize the web application image.
FROM final AS final-web
# Copy the executable from the "build" stage.
COPY --chown=meower:meower --from=build /bin/web-server /opt/meower/
# Copy the production CSS files from the "assets" stage.
COPY --from=assets /usr/src/main.css /opt/meower/main.css
# Copy the production node_modules from the "assets" stage.
COPY --from=assets /usr/src/node_modules node_modules
# Expose the port that the application listens on.
EXPOSE 3000

WORKDIR /opt/meower/

# What the container should run when it is started.
ENTRYPOINT [ "/opt/meower/web-server" ]

################################################################################
# Create a stage to finalize the gRPC API image.
FROM final AS final-api
# Copy the executable from the "build" stage.
COPY --chown=meower:meower --from=build /bin/api-server /opt/meower/

# Expose the port that the application listens on.
EXPOSE 50051

WORKDIR /opt/meower/

# What the container should run when it is started.
ENTRYPOINT [ "/opt/meower/api" ]


################################################################################
# Create a development image that includes what's required to generate the protoc and CSS files
FROM build AS development
RUN apk add --no-cache \
    protobuf-dev

# Copy the executables from the "build" stage.
COPY --chown=meower:meower --from=build /bin/web-server /opt/meower/
COPY --chown=meower:meower --from=build /bin/api-server /opt/meower/

RUN --mount=type=cache,target=/go/pkg/mod/ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN --mount=type=cache,target=/go/pkg/mod/ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN --mount=type=cache,target=/go/pkg/mod/ go install github.com/grpc-ecosystem/grpc-health-probe@latest
RUN --mount=type=cache,target=/go/pkg/mod/ go install github.com/bokwoon95/wgo@latest

# Copy tailwind CLI binary from the "assets" stage.
COPY --from=assets /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss
# Copy the production node_modules from the "assets" stage.
COPY --from=assets /usr/src/node_modules node_modules

# Set the default target to be the final stage.
FROM final
