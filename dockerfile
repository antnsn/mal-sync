# Dockerfile for mal-sync

# TODO: Replace 'grafana/mimirtool:latest' with the specific mimirtool base image and tag you want to use.
# For example, it might be something like 'grafana/mimirtool:2.11.1' or another official image.
FROM grafana/mimirtool:latest AS base

# It's good practice to switch to root for installations if the base image isn't already root.
# Some images might require this, others might already be root or have sudo.
# If the base image is distroless or very minimal, you might need a different approach
# or use a multi-stage build starting from a fuller OS image for the download step.
USER root

# Install necessary packages for downloading and unzipping.
# The required package manager depends on the base image's OS.
# Example for Alpine (common for Grafana tools):
RUN apk add --no-cache curl unzip

# Example for Debian/Ubuntu based images (if apk fails):
# RUN apt-get update && apt-get install -y curl unzip --no-install-recommends && rm -rf /var/lib/apt/lists/*

# Define LokiTool version and URL for clarity and easy updates
ARG LOKITOOL_VERSION=v3.5.1
ARG LOKITOOL_DOWNLOAD_URL=https://github.com/grafana/loki/releases/download/${LOKITOOL_VERSION}/lokitool-linux-amd64.zip

# Download, unzip, and install lokitool
RUN echo "Downloading lokitool from ${LOKITOOL_DOWNLOAD_URL}..." && \
    curl -sSL ${LOKITOOL_DOWNLOAD_URL} -o /tmp/lokitool-linux-amd64.zip && \
    echo "Unzipping lokitool..." && \
    unzip -o /tmp/lokitool-linux-amd64.zip -d /usr/local/bin/ && \
    # The zip might contain the binary directly or in a folder. Assuming it's directly named lokitool-linux-amd64
    # If it's just 'lokitool', adjust accordingly.
    # Let's assume the binary inside the zip is named 'lokitool-linux-amd64'
    mv /usr/local/bin/lokitool-linux-amd64 /usr/local/bin/lokitool && \
    echo "Setting execute permissions for lokitool..." && \
    chmod +x /usr/local/bin/lokitool && \
    echo "Cleaning up..." && \
    rm /tmp/lokitool-linux-amd64.zip && \
    echo "lokitool installation complete. Version:" && \
    lokitool version

# --- Go Application Integration ---
# At this stage, the image contains both mimirtool (from base) and lokitool.
# Now, you would typically add your Go application.
# This usually involves a multi-stage build for the Go app itself.

# Example: Multi-stage build for your Go application
# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire module context (including cmd/ and internal/)
COPY . .

# Build the Go application
# The build context is now the root of the module.
# We target the main package within cmd/mal-sync/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /mal-sync-app ./cmd/mal-sync

# Stage 2: Final image (using the 'base' stage created above)
FROM base AS final

# Copy the mimirtool (already in base) and lokitool (added in base)
# Copy your compiled Go application from the builder stage
COPY --from=builder /mal-sync-app /usr/local/bin/mal-sync

# Ensure your Go application is executable
RUN chmod +x /usr/local/bin/mal-sync

# Set the entrypoint to your Go application
ENTRYPOINT [ "/usr/local/bin/mal-sync" ]

# If your Go app was designed to run as a non-root user, and the base mimirtool image
# had a non-root user (e.g., 'mimir'), you might switch back:
# USER mimir

# Expose any ports if your application listens on them (unlikely for a CLI sync tool)
# CMD ["--help"] # Default command for your app