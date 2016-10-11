package crypto

import (
	"testing"
)

func dataEqual(first, second []byte) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func TestCryptoRoundtrip(t *testing.T) {
	expected := []byte("data")
	input := []byte("data")
	password, err := GenerateKey([]byte("password"), []byte("salt"), 15)

	if err != nil {
		t.Errorf("Failed to generate key because %s", err)
		return
	}

	encrypted, err := Encrypt(input, password)
	if err != nil {
		t.Errorf("Failed to encrypt data because %s", err)
		return
	}

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Errorf("Failed to decrypt data because %s", err)
		return
	}

	if !dataEqual(expected, decrypted) {
		t.Errorf("Value didn't match expected\nexpected: %#v\ngot:%#v", expected, decrypted)
	}
}
