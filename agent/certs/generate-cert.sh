#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$DIR"

echo "🔐 Generant certificat auto-signat de signatura de codi (OpenSSL)..."

PASS="Encertia2026!"

# 1. Generar clau privada RSA
openssl genrsa -out encertia-codesign.key 2048

# 2. Fitxer de configuració OpenSSL per a Code Signing (Extended Key Usage)
cat > openssl-codesign.cnf <<EOF
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
prompt              = no
x509_extensions     = v3_codesign

[ req_distinguished_name ]
C   = ES
ST  = Catalunya
L   = Barcelona
O   = Encertia
CN  = Encertia Authority

[ v3_codesign ]
keyUsage            = critical, digitalSignature
extendedKeyUsage    = codeSigning
basicConstraints    = critical, CA:FALSE
subjectKeyIdentifier = hash
EOF

# 3. Crear el certificat X.509
openssl req -new -x509 -days 1825 -key encertia-codesign.key -out encertia-codesign.crt -config openssl-codesign.cnf

# 4. Convertir a format DER (.cer per a Windows)
openssl x509 -outform der -in encertia-codesign.crt -out encertia-codesign.cer

# 5. Empaquetar a PKCS#12 (.pfx) per a signtool
openssl pkcs12 -export -out encertia-codesign.pfx -inkey encertia-codesign.key -in encertia-codesign.crt -passout pass:"$PASS"

rm -f openssl-codesign.cnf

echo "✅ Certificats generats amb èxit a $DIR:"
echo "   - encertia-codesign.pfx (Clau privada + certificat per signar executables. Contrasenya: $PASS)"
echo "   - encertia-codesign.cer (Certificat públic per instal·lar als ordinadors de l'aula)"
