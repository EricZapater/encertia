package fingerprint

import (
	"testing"
)

func TestGetHardwareUUID(t *testing.T) {
	uuid, err := GetHardwareUUID()
	if err != nil {
		t.Logf("Avis: No s'ha pogut obtenir el Hardware UUID en aquest entorn (comú en contenidors/sandbox): %v", err)
		return
	}
	if uuid == "" {
		t.Error("S'esperava un UUID no buit")
	}
	t.Logf("Hardware UUID obtingut: %s", uuid)
}

func TestGetHardwareUUIDHash(t *testing.T) {
	hash, err := GetHardwareUUIDHash()
	if err != nil {
		t.Logf("Avis: No s'ha pogut calcular el hash en aquest entorn: %v", err)
		return
	}
	if len(hash) != 64 {
		t.Errorf("El hash SHA-256 ha de tenir exactament 64 caràcters hex, obtingut: %d", len(hash))
	}
	t.Logf("Hardware UUID Hash (SHA-256): %s", hash)
}
