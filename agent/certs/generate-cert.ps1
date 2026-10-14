# Script PowerShell per generar un certificat auto-signat de signatura de codi
$ErrorActionPreference = "Stop"

$PfxPasswordPlain = "Encertia2026!"
$PfxPassword = ConvertTo-SecureString -String $PfxPasswordPlain -Force -AsPlainText

Write-Host "🔐 Generant certificat de signatura de codi (Code Signing Cert)..." -ForegroundColor Cyan

$cert = New-SelfSignedCertificate -Type CodeSigningCert `
    -Subject "CN=Encertia Authority, O=Encertia, C=ES" `
    -KeyUsage DigitalSignature `
    -FriendlyName "Certificat Encertia Agent" `
    -CertStoreLocation "Cert:\CurrentUser\My" `
    -NotAfter (Get-Date).AddYears(5)

$CerPath = Join-Path $PSScriptRoot "encertia-codesign.cer"
$PfxPath = Join-Path $PSScriptRoot "encertia-codesign.pfx"

Export-Certificate -Cert $cert -FilePath $CerPath | Out-Null
Export-PfxCertificate -Cert $cert -FilePath $PfxPath -Password $PfxPassword | Out-Null

Write-Host "✅ Certificats generats amb èxit:" -ForegroundColor Green
Write-Host "   - $CerPath (Certificat públic per a equips)"
Write-Host "   - $PfxPath (PFX protegit per signar. Clau: $PfxPasswordPlain)"
