package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

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
	fmt.Println("🔐 Verificant autorització d'aquest equip amb Encertia...")

	resp, err := client.VerifyDevice(hwHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ ACCÉS DENEGAT: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nAquest ordinador no està autoritzat o ha estat donat de baixa.")
		fmt.Fprintln(os.Stderr, "Si ets docent, sol·licita al teu administrador que registri aquest equip:")
		fmt.Fprintln(os.Stderr, "   encertia-agent --register")
		fmt.Println("\nPrem Enter per sortir...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(1)
	}

	fmt.Printf("✅ Equip autoritzat: %s\n", resp.DeviceName)
	fmt.Printf("👤 Docent assignat:  %s\n", resp.TeacherEmail)

	targetURL := fmt.Sprintf(
		"%s/login?deviceToken=%s&email=%s",
		strings.TrimRight(webURL, "/"),
		url.QueryEscape(resp.BootstrapToken),
		url.QueryEscape(resp.TeacherEmail),
	)

	if debug {
		fmt.Printf("🔗 URL generada: %s\n", targetURL)
		return
	}

	fmt.Println("🚀 Obrint Encertia al teu navegador...")
	if err := browser.OpenURL(targetURL); err != nil {
		fmt.Printf("⚠️ No s'ha pogut obrir el navegador automàticament: %v\n", err)
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

	fmt.Println("\n⏳ Autenticant administrador...")
	adminToken, err := client.LoginAdmin(adminEmail, adminPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error d'autenticació com a administrador: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("⏳ Registrant equip a la plataforma...")
	dev, err := client.RegisterDevice(adminToken, hwHash, deviceName, teacherEmail)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error donant d'alta l'equip: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n==================================================")
	fmt.Println("✅ EQUIP REGISTRAT AMB ÈXIT!")
	fmt.Printf("   Nom de l'equip:  %s\n", dev.Name)
	fmt.Printf("   Docent assignat: %s (%s)\n", dev.TeacherName, dev.TeacherEmail)
	fmt.Printf("   Identificador:   %s\n", dev.ID)
	fmt.Println("==================================================")
	fmt.Println("A partir d'ara, el professor pot accedir fent doble clic")
	fmt.Println("a l'executable encertia-agent.")
	fmt.Println()
}

// readPassword reads sensitive input from console without echoing on supported terminals
func readPassword() string {
	var passBytes []byte
	var err error

	// Intent de lectura oculta mitjançant syscall si està suportat
	fd := int(syscall.Stdin)
	if isTerminal(fd) {
		passBytes, err = readPasswordStdin()
		if err == nil {
			return string(passBytes)
		}
	}

	// Fallback lector estàndard si no és TTY o falla
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func isTerminal(fd int) bool {
	// Comprovació bàsica de descriptor
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func readPasswordStdin() ([]byte, error) {
	// Fallback directe a lectura de caràcters
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return []byte(strings.TrimRight(string(line), "\r\n")), nil
}
