#!/usr/bin/env bash
set -euo pipefail

version="${1:?Usage: build-release.sh VERSION}"
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Version must have the form vMAJOR.MINOR.PATCH" >&2
  exit 1
fi

mkdir -p dist
output="$(cd dist && pwd)"
staging="$(mktemp -d)"
trap 'rm -rf "$staging"' EXIT

for os in windows linux darwin; do
  for arch in amd64 arm64; do
    name="uix_${version}_${os}_${arch}"
    package="$staging/$name"
    mkdir -p "$package"
    binary=uix
    if [[ "$os" == windows ]]; then binary=uix.exe; fi
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
      -trimpath -buildvcs=false -ldflags "-s -w -X main.version=$version" \
      -o "$package/$binary" ./cmd/uix
    cp README.md "$package/README.md"
    if [[ "$os" == windows ]]; then
      (cd "$package" && zip -q "$output/$name.zip" "$binary" README.md)
    else
      tar -czf "dist/$name.tar.gz" -C "$package" "$binary" README.md
    fi
  done
done

(cd dist && sha256sum "uix_${version}_"* > checksums.txt)
