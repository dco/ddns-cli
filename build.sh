#!/bin/bash

APP_NAME="ddns-cli"
APP_VERSION=${TAG_NAME:-"unknown"}

rm -rf dist
mkdir -p dist

PLATFORMS=(
  "aix/ppc64"
  "android/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "dragonfly/amd64"
  "freebsd/386"
  "freebsd/amd64"
  "freebsd/arm"
  "freebsd/arm64"
  "freebsd/riscv64"
  "illumos/amd64"
  "linux/386"
  "linux/amd64"
  "linux/arm/v5"
  "linux/arm/v6"
  "linux/arm/v7"
  "linux/arm64"
  "linux/loong64"
  "linux/mips"
  "linux/mips64"
  "linux/mips64le"
  "linux/mipsle"
  "linux/ppc64"
  "linux/ppc64le"
  "linux/riscv64"
  "linux/s390x"
  "netbsd/386"
  "netbsd/amd64"
  "netbsd/arm"
  "netbsd/arm64"
  "openbsd/386"
  "openbsd/amd64"
  "openbsd/arm"
  "openbsd/arm64"
  "openbsd/ppc64"
  "openbsd/riscv64"
  "plan9/386"
  "plan9/amd64"
  "plan9/arm"
  "solaris/amd64"
  "windows/386"
  "windows/amd64"
  "windows/arm"
  "windows/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS=$(echo "$PLATFORM" | cut -d'/' -f1)
  GOARCH=$(echo "$PLATFORM" | cut -d'/' -f2)
  GOVARIANT=$(echo "$PLATFORM" | cut -d'/' -f3)

  OUTPUT_DIR="dist/$APP_NAME-$GOOS-$GOARCH"
  if [ -n "$GOVARIANT" ]; then
    OUTPUT_DIR="$OUTPUT_DIR-$GOVARIANT"
  fi
  mkdir -p "$OUTPUT_DIR"

  OUTPUT="$OUTPUT_DIR/$APP_NAME"
  if [ "$GOOS" == "windows" ]; then
    OUTPUT="$OUTPUT.exe"
  fi

  echo "Building $PLATFORM -> $OUTPUT"
  GOOS=$GOOS GOARCH=$GOARCH GOARM=${GOVARIANT#v} go build -trimpath -ldflags="-s -w -X 'main.Version=$APP_VERSION'" -o "$OUTPUT" .

  if [ $? -ne 0 ]; then
    echo "Failed to build $PLATFORM"
    continue
  fi

  if [ "$GOOS" == "windows" ]; then
    zip -j "$OUTPUT_DIR.zip" "$OUTPUT"
  else
    tar -czf "$OUTPUT_DIR.tar.gz" -C "$OUTPUT_DIR" .
  fi
done

echo "All builds completed. Output files are in the dist/ directory."