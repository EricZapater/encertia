#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$DIR"

mkdir -p "$DIR/bin"

echo "🔨 Compilant encertia-agent per a Linux (amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-linux" ./cmd/main.go

echo "🔨 Compilant encertia-agent per a macOS (universal / arm64 + amd64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-darwin-arm64" ./cmd/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-darwin-amd64" ./cmd/main.go
# Unió universal si estem a macOS amb lipo, o copia de l'arquitectura habitual
if command -v lipo >/dev/null 2>&1; then
    lipo -create -output "$DIR/bin/encertia-agent-darwin" "$DIR/bin/encertia-agent-darwin-arm64" "$DIR/bin/encertia-agent-darwin-amd64"
    rm -f "$DIR/bin/encertia-agent-darwin-arm64" "$DIR/bin/encertia-agent-darwin-amd64"
else
    cp "$DIR/bin/encertia-agent-darwin-arm64" "$DIR/bin/encertia-agent-darwin"
fi

echo "🔨 Compilant encertia-agent per a Windows (amd64)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-windows.exe" ./cmd/main.go

echo "✅ Tots els binaris s'han generat correctament a $DIR/bin/:"
ls -lh "$DIR/bin"
