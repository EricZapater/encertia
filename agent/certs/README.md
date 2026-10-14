# Certificats de Signatura de Codi per a Encertia Agent

Aquest directori conté les eines per generar certificats de signatura de codi (Code Signing) per signar l'executable de Windows (`encertia-agent-windows.exe`) i evitar advertències de Windows Defender SmartScreen en equips corporatius o d'aula.

---

## 1. Generació del Certificat de Signatura

### Opció A: A Windows amb PowerShell (Recomanada)
Executa l'script [generate-cert.ps1](file:///Users/eric.zapater/Developer/encertia/agent/certs/generate-cert.ps1) o el següent bloc de comandes en una consola PowerShell:

```powershell
# 1. Crear el certificat auto-signat de Code Signing (vàlid 5 anys)
$cert = New-SelfSignedCertificate -Type CodeSigningCert `
    -Subject "CN=Encertia Authority, O=Encertia, C=ES" `
    -KeyUsage DigitalSignature `
    -FriendlyName "Certificat Encertia Agent" `
    -CertStoreLocation "Cert:\CurrentUser\My" `
    -NotAfter (Get-Date).AddYears(5)

# 2. Exportar la Clau Pública (.cer) - AQUESTA ÉS LA QUE S'INSTAL·LA ALS EQUIPS
Export-Certificate -Cert $cert -FilePath "encertia-codesign.cer"

# 3. Exportar el fitxer PFX protegit per contrasenya (per signar executables)
$pwd = ConvertTo-SecureString -String "Encertia2026!" -Force -AsPlainText
Export-PfxCertificate -Cert $cert -FilePath "encertia-codesign.pfx" -Password $pwd
```

### Opció B: A Linux / macOS amb OpenSSL
Executa l'script [generate-cert.sh](file:///Users/eric.zapater/Developer/encertia/agent/certs/generate-cert.sh):

```bash
chmod +x generate-cert.sh
./generate-cert.sh
```

---

## 2. Com signar el binari de Windows (`.exe`)

Un cop obtingut el fitxer `encertia-codesign.pfx`:

```cmd
signtool sign /f encertia-codesign.pfx /p "Encertia2026!" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 encertia-agent-windows.exe
```

---

## 3. Com instal·lar el certificat públic (`.cer`) a l'ordinador d'aula

Perquè Windows consideri l'executable com a editor de confiança sense necessitat de permisos d'administrador:

```cmd
certutil -user -addstore "TrustedPublisher" encertia-codesign.cer
```
*(No demana cap contrasenya d'administrador de Windows perquè s'instal·la al magatzem de l'usuari actual).*
