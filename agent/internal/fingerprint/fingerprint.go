package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// GetHardwareUUID extracts the immutable motherboard / SMBIOS UUID of the host system.
func GetHardwareUUID() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsUUID()
	case "darwin":
		return getDarwinUUID()
	case "linux":
		return getLinuxUUID()
	default:
		return "", errors.New("sistema operatiu no suportat per a l'extracció de hardware UUID")
	}
}

// GetHardwareUUIDHash returns the hex-encoded SHA-256 hash of the hardware UUID.
func GetHardwareUUIDHash() (string, error) {
	rawUUID, err := GetHardwareUUID()
	if err != nil {
		return "", err
	}
	trimmed := strings.ToLower(strings.TrimSpace(rawUUID))
	if trimmed == "" {
		return "", errors.New("el hardware UUID obtingut és buit")
	}

	hash := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(hash[:]), nil
}

func getWindowsUUID() (string, error) {
	// Intent 1: PowerShell Get-CimInstance
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "(Get-CimInstance Win32_ComputerSystemProduct).UUID")
	out, err := cmd.Output()
	if err == nil {
		clean := strings.TrimSpace(string(out))
		if clean != "" && !strings.Contains(clean, " ") {
			return clean, nil
		}
	}

	// Intent 2: Fallback wmic
	cmdWmic := exec.Command("wmic", "csproduct", "get", "uuid")
	outWmic, errWmic := cmdWmic.Output()
	if errWmic == nil {
		lines := strings.Split(string(outWmic), "\n")
		for _, line := range lines {
			val := strings.TrimSpace(line)
			if val != "" && !strings.EqualFold(val, "uuid") {
				return val, nil
			}
		}
	}

	return "", errors.New("no s'ha pogut obtenir el UUID de la placa base a Windows")
}

func getDarwinUUID() (string, error) {
	// Consulta IORegistry per a IOPlatformExpertDevice
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", errors.New("no s'ha trobat IOPlatformUUID al registre d'E/S de macOS")
}

func getLinuxUUID() (string, error) {
	// Intent 1: /sys/class/dmi/id/product_uuid
	if data, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		clean := strings.TrimSpace(string(data))
		if clean != "" {
			return clean, nil
		}
	}

	// Intent 2: /etc/machine-id
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		clean := strings.TrimSpace(string(data))
		if clean != "" {
			return clean, nil
		}
	}

	// Intent 3: /var/lib/dbus/machine-id
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		clean := strings.TrimSpace(string(data))
		if clean != "" {
			return clean, nil
		}
	}

	return "", errors.New("no s'ha pogut obtenir el machine UUID a Linux")
}
