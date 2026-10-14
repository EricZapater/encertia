#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$DIR"

mkdir -p "$DIR/bin"

# 1. Generar recurs d'icona per a Windows si es disposa de rsrc i de icon.ico
RSRC_BIN="$(command -v rsrc || echo "$HOME/go/bin/rsrc")"
if [ -x "$RSRC_BIN" ] && [ -f "$DIR/assets/icon.ico" ]; then
    echo "🎨 Generant recurs d'icona de Windows (rsrc_windows_amd64.syso)..."
    "$RSRC_BIN" -ico "$DIR/assets/icon.ico" -arch amd64 -o "$DIR/cmd/rsrc_windows_amd64.syso"
fi

echo "🔨 Compilant encertia-agent per a Linux (amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-linux" ./cmd

echo "🔨 Compilant encertia-agent per a macOS (universal / arm64 + amd64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-darwin-arm64" ./cmd
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-darwin-amd64" ./cmd
# Unió universal si estem a macOS amb lipo, o copia de l'arquitectura habitual
if command -v lipo >/dev/null 2>&1; then
    lipo -create -output "$DIR/bin/encertia-agent-darwin" "$DIR/bin/encertia-agent-darwin-arm64" "$DIR/bin/encertia-agent-darwin-amd64"
    rm -f "$DIR/bin/encertia-agent-darwin-arm64" "$DIR/bin/encertia-agent-darwin-amd64"
else
    cp "$DIR/bin/encertia-agent-darwin-arm64" "$DIR/bin/encertia-agent-darwin"
fi

# Aplicar icona a macOS si es disposa de Rez i SetFile
if command -v Rez >/dev/null 2>&1 && command -v SetFile >/dev/null 2>&1 && [ -f "$DIR/assets/icon.icns" ]; then
    echo "🎨 Incrustant icona a l'executable de macOS..."
    TMP_REZ="/tmp/encertia_icon_$$.r"
    echo "read 'icns' (-16455) \"$DIR/assets/icon.icns\";" > "$TMP_REZ"
    Rez -a "$TMP_REZ" -o "$DIR/bin/encertia-agent-darwin" 2>/dev/null || true
    SetFile -a C "$DIR/bin/encertia-agent-darwin" 2>/dev/null || true
    rm -f "$TMP_REZ"
fi

echo "🔨 Compilant encertia-agent per a Windows (amd64)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/encertia-agent-windows.exe" ./cmd

echo "📦 Empaquetant encertia-agent-windows.zip..."
TMP_ZIP_DIR="/tmp/encertia_win_pkg_$$"
mkdir -p "$TMP_ZIP_DIR"
cp "$DIR/bin/encertia-agent-windows.exe" "$TMP_ZIP_DIR/"
if [ -f "$DIR/certs/encertia-codesign.cer" ]; then
    cp "$DIR/certs/encertia-codesign.cer" "$TMP_ZIP_DIR/"
fi
cat << 'EOF' > "$TMP_ZIP_DIR/INSTRUCCIONS.txt"
ENCERTIA AGENT PER A WINDOWS
============================

Instruccions d'ús:

1. Descomprimeix aquest fitxer ZIP a una carpeta de la teva elecció.

2. Obre una consola de comandes (PowerShell o CMD) a la carpeta i executa:
   .\encertia-agent-windows.exe --register

   Seguiu les instruccions en pantalla per vincular aquest ordinador amb el compte de docent.

3. (Opcional) Si Windows Defender SmartScreen mostra un avís en executar-lo:
   - Fes clic a "Més informació" i seguidament a "Executa de totes maneres".
   - Per afegir el certificat de signatura de codi com a editor de confiança a l'equip i evitar avisos futurs, executa:
     certutil -user -addstore "TrustedPublisher" encertia-codesign.cer
EOF
(cd "$TMP_ZIP_DIR" && zip -q "$DIR/bin/encertia-agent-windows.zip" *)
rm -rf "$TMP_ZIP_DIR"

echo "✅ Tots els binaris s'han generat correctament a $DIR/bin/:"
ls -lh "$DIR/bin"
