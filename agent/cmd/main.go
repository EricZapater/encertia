package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"encertia-agent/internal/api"
	"encertia-agent/internal/browser"
	"encertia-agent/internal/fingerprint"
)

const (
	defaultAPIServer = "https://api.encertia.ericzapater.cat/api/v1"
	defaultWebURL    = "https://encertia.ericzapater.cat"
)

func main() {
	registerMode := flag.Bool("register", false, "Mode interactiu de registre d'equip per a administradors")
	shortRegister := flag.Bool("r", false, "Mode interactiu de registre (àlies)")
	debugMode := flag.Bool("debug", false, "Mostra la informació de maquinari i sortida sense obrir el navegador")
	apiServer := flag.String("api", defaultAPIServer, "URL base de l'API d'Encertia")
	webURL := flag.String("web", defaultWebURL, "URL base de la plataforma web Encertia")
	flag.Parse()

	// 1. Obtenir el Hardware UUID Hash
	hwHash, err := fingerprint.GetHardwareUUIDHash()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error obtenint el Hardware UUID de l'equip: %v\n", err)
		os.Exit(1)
	}

	if *debugMode {
		rawUUID, _ := fingerprint.GetHardwareUUID()
		fmt.Println("=== INFORMACIÓ DE MAQUINARI (DEBUG) ===")
		fmt.Printf("Hardware UUID:      %s\n", rawUUID)
		fmt.Printf("Hardware SHA-256:   %s\n", hwHash)
		fmt.Printf("API Server:         %s\n", *apiServer)
		fmt.Printf("Web URL:            %s\n", *webURL)
		fmt.Println("=======================================")
	}

	client := api.NewClient(*apiServer)

	// Mode Registre
	if *registerMode || *shortRegister {
		runRegisterWizard(client, hwHash)
		return
	}

	// Mode Normal: Verificació i Obertura del Navegador
	runNormalLaunch(client, hwHash, *webURL, *debugMode)
}

func runNormalLaunch(client *api.Client, hwHash, webURL string, debug bool) {
	fmt.Println("==================================================")
	fmt.Println("        ENCERTIA — AGENT D'ACCÉS DOCENT           ")
	fmt.Println("==================================================")
	fmt.Printf("[%s] [1/2] Petjada digital de maquinari: %s...\n", time.Now().Format("15:04:05"), hwHash[:16])
	fmt.Printf("[%s] [2/2] Verificant autorització amb el servidor (%s)...\n", time.Now().Format("15:04:05"), client.BaseURL)

	resp, err := client.VerifyDevice(hwHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[%s] ❌ ACCÉS DENEGAT: %v\n", time.Now().Format("15:04:05"), err)
		fmt.Fprintln(os.Stderr, "\nAquest ordinador no està autoritzat o ha estat donat de baixa.")
		fmt.Fprintln(os.Stderr, "Si ets docent, sol·licita al teu administrador que registri aquest equip:")
		fmt.Fprintln(os.Stderr, "   encertia-agent --register")
		fmt.Println("\nPrem Enter per sortir...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(1)
	}

	fmt.Printf("[%s] ✅ Equip autoritzat: %s\n", time.Now().Format("15:04:05"), resp.DeviceName)
	fmt.Printf("[%s] 👤 Docent assignat:  %s\n", time.Now().Format("15:04:05"), resp.TeacherEmail)

	targetURL := fmt.Sprintf(
		"%s/login?deviceToken=%s&email=%s",
		strings.TrimRight(webURL, "/"),
		url.QueryEscape(resp.BootstrapToken),
		url.QueryEscape(resp.TeacherEmail),
	)

	if debug {
		fmt.Printf("[%s] 🔗 URL generada (debug): %s\n", time.Now().Format("15:04:05"), targetURL)
		return
	}

	fmt.Printf("[%s] 🚀 Obrint Encertia al navegador web...\n", time.Now().Format("15:04:05"))
	if err := browser.OpenURL(targetURL); err != nil {
		fmt.Printf("[%s] ⚠️ No s'ha pogut obrir el navegador automàticament: %v\n", time.Now().Format("15:04:05"), err)
		fmt.Printf("Obre aquest enllaç manualment:\n%s\n", targetURL)
	}
}

func runRegisterWizard(client *api.Client, hwHash string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("==================================================")
	fmt.Println("   REGISTRE D'EQUIP AUTORITZAT — ENCERTIA        ")
	fmt.Println("==================================================")
	fmt.Println("Aquest assistent registrarà aquest ordinador per")
	fmt.Println("a un professor concret a la plataforma Encertia.")
	fmt.Println()

	fmt.Print("1. Correu electrònic de l'ADMINISTRADOR: ")
	adminEmail, _ := reader.ReadString('\n')
	adminEmail = strings.TrimSpace(adminEmail)

	fmt.Print("2. Contrasenya de l'ADMINISTRADOR: ")
	adminPassword := readPassword()
	fmt.Println()

	fmt.Print("3. Correu electrònic del PROFESSOR assignat a aquest equip: ")
	teacherEmail, _ := reader.ReadString('\n')
	teacherEmail = strings.TrimSpace(teacherEmail)

	hostname, _ := os.Hostname()
	defaultName := "Ordinador Docent"
	if hostname != "" {
		defaultName = fmt.Sprintf("Ordinador (%s)", hostname)
	}
	fmt.Printf("4. Nom descriptiu de l'equip [%s]: ", defaultName)
	deviceName, _ := reader.ReadString('\n')
	deviceName = strings.TrimSpace(deviceName)
	if deviceName == "" {
		deviceName = defaultName
	}

	fmt.Println()
	fmt.Println("--------------------------------------------------")
	fmt.Println("📋 EXECUCIÓ DEL REGISTRE — LOGGING PER CONSOLA")
	fmt.Println("--------------------------------------------------")

	logStep := func(step string, format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf("[%s] [%s] %s\n", time.Now().Format("15:04:05"), step, msg)
	}

	logStep("INFO", "Identificador de maquinari (Hardware UUID Hash): %s", hwHash)
	logStep("INFO", "Punt d'accés API configurat: %s", client.BaseURL)
	logStep("AUTH", "Autenticant administrador (%s) a %s/auth/login...", adminEmail, client.BaseURL)

	adminToken, err := client.LoginAdmin(adminEmail, adminPassword)
	if err != nil {
		logStep("ERROR", "❌ Error d'autenticació com a administrador: %v", err)
		fmt.Println("\nPrem Enter per sortir...")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}
	logStep("AUTH", "✔️ Administrador autenticat amb èxit. Token de sessió obtingut.")

	logStep("REGISTER", "Enviant petició de registre a POST %s/admin/devices...", client.BaseURL)
	logStep("REGISTER", "Dades del registre:")
	logStep("REGISTER", "  - Nom descriptiu:  %s", deviceName)
	logStep("REGISTER", "  - Docent assignat: %s", teacherEmail)
	logStep("REGISTER", "  - Hardware Hash:   %s", hwHash)

	dev, err := client.RegisterDevice(adminToken, hwHash, deviceName, teacherEmail)
	if err != nil {
		logStep("ERROR", "❌ Error donant d'alta l'equip: %v", err)
		fmt.Println("\nPrem Enter per sortir...")
		_, _ = reader.ReadString('\n')
		os.Exit(1)
	}

	logStep("SUCCESS", "✔️ Resposta rebuda del servidor amb èxit (HTTP 201 Created).")
	fmt.Println("\n==================================================")
	fmt.Println("✅ EQUIP REGISTRAT AMB ÈXIT!")
	fmt.Printf("   Nom de l'equip:  %s\n", dev.Name)
	fmt.Printf("   Docent assignat: %s (%s)\n", dev.TeacherName, dev.TeacherEmail)
	fmt.Printf("   Identificador:   %s\n", dev.ID)
	fmt.Printf("   Estat:           Actiu (autoritzat per a la docència)\n")
	fmt.Println("==================================================")
	fmt.Println("A partir d'ara, el professor pot accedir fent doble clic")
	fmt.Println("a l'executable encertia-agent.")
	fmt.Println()
	fmt.Println("Prem Enter per finalitzar...")
	_, _ = reader.ReadString('\n')
}

// readPassword reads sensitive input from console without echoing
func readPassword() string {
	if runtime.GOOS == "windows" {
		// A Windows, Read-Host amb -AsSecureString oculta la contrasenya nativament
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			`$p = Read-Host -AsSecureString; [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($p))`)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	} else {
		// A macOS i Linux, desactivem l'eco del terminal mitjançant stty
		_ = exec.Command("stty", "-echo").Run()
		defer func() {
			_ = exec.Command("stty", "echo").Run()
		}()

		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		_ = exec.Command("stty", "echo").Run()
		return strings.TrimSpace(line)
	}

	// Fallback estàndard si falla stty o powershell
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
