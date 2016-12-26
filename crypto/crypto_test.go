package crypto

import (
	"io/ioutil"
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
	t.Parallel()
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

func TestCryptoRead(t *testing.T) {
	t.Parallel()
	key, err := GenerateKey([]byte("password"), []byte("salt"), 17)
	if err != nil {
		t.Errorf("Failed to create key because %s", err)
		return
	}

	plaintext := []byte("This is test data")

	data, err := ioutil.ReadFile("./format-1.cpt")
	if err != nil {
		t.Errorf("Failed to read file from disk because %s", err)
		return
	}

	decrypted, err := Decrypt(data, key)
	if err != nil {
		t.Errorf("Failed to decrypt data because %s", err)
		return
	}

	if !dataEqual(decrypted, plaintext) {
		t.Errorf("Decrypted data didn't match expected.\nExpected: '%s'\nGot:      '%s'", plaintext, decrypted)
	}
}
