#!/bin/bash
# AINAR Build Script
# Builds the application for the current platform or specified targets

set -e

VERSION=${VERSION:-"dev"}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS="-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}"

echo "Building AINAR ${VERSION}..."

# Build frontend
echo "Building frontend..."
cd web
npm ci
npm run build
cd ..

# Copy frontend to embed directory
echo "Copying frontend assets..."
rm -rf internal/webui/dist/*
cp -r web/dist/* internal/webui/dist/

# Build Go binary
echo "Building backend..."
if [ -z "$1" ]; then
    # Build for current platform
    go build -ldflags="${LDFLAGS}" -o ainar ./cmd/ainar
    echo "Built: ainar"
else
    # Build for specified targets
    for target in "$@"; do
        GOOS=${target%/*}
        GOARCH=${target#*/}
        OUTPUT="ainar-${GOOS}-${GOARCH}"
        if [ "$GOOS" = "windows" ]; then
            OUTPUT="${OUTPUT}.exe"
        fi
        echo "Building ${OUTPUT}..."
        CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="${LDFLAGS}" -o "${OUTPUT}" ./cmd/ainar
    done
fi

echo "Build complete!"
