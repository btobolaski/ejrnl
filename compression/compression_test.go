package compression

import (
	"io/ioutil"
	"testing"

	"github.com/btobolaski/ejrnl/crypto"
)

var testData = []byte("{\"date\": \"2016-12-13T00:00:00Z\",\"id\": \"1234\",\"tags\":[],\"body\":\"Hello\"}")

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

func TestCompressionRoundtrip(t *testing.T) {
	t.Parallel()
	expected := []byte("data")
	input := []byte("data")
	key, err := crypto.GenerateKey([]byte("password"), []byte("salt"), 12)
	if err != nil {
		t.Errorf("Failed to generate a key because %s", err)
		return
	}
	compressed, err := CompressAndEncrypt(input, key)
	if err != nil {
		t.Errorf("Failed to compress because %s", err)
		return
	}
	roundtripped, err := DecryptAndDecompress(compressed, key)
	if err != nil {
		t.Errorf("Failed to decompress because %s", err)
		return
	}

	if !dataEqual(roundtripped, expected) {
		t.Errorf("Data didn't match expected.\ngot:    %#v\n expected: %#v", roundtripped, expected)
	}
}

func TestV1Decode(t *testing.T) {
	t.Parallel()
	key, err := crypto.GenerateKey([]byte("password"), []byte("salt"), 17)
	if err != nil {
		t.Errorf("Failed to create key because %s", err)
		return
	}

	data, err := ioutil.ReadFile("./format-1.cpt")
	if err != nil {
		t.Errorf("Failed to read file from disk because %s", err)
		return
	}

	decrypted, err := DecryptAndDecompress(data, key)
	if err != nil {
		t.Errorf("Failed to decrypt data because %s", err)
		return
	}

	if !dataEqual(decrypted, testData) {
		t.Errorf("Decrypted data didn't match expected.\nExpected: '%s'\nGot:      '%s'", testData, decrypted)
	}
}

func TestV2Decode(t *testing.T) {
	t.Parallel()
	key, err := crypto.GenerateKey([]byte("password"), []byte("salt"), 17)
	if err != nil {
		t.Errorf("Failed to create key because %s", err)
		return
	}

	cyphertext, err := CompressAndEncrypt(testData, key)
	if err != nil {
		t.Errorf("Failed to compress text because %s", err)
		return
	}

	err = ioutil.WriteFile("format-2.cpt", cyphertext, 0644)
	if err != nil {
		t.Errorf("Failed to write database %s", err)
		return
	}

	data, err := ioutil.ReadFile("./format-2.cpt")
	if err != nil {
		t.Errorf("Failed to read file from disk because %s", err)
		return
	}

	decrypted, err := DecryptAndDecompress(data, key)
	if err != nil {
		t.Errorf("Failed to decrypt data because %s", err)
		return
	}

	if !dataEqual(decrypted, testData) {
		t.Errorf("Decrypted data didn't match expected.\nExpected: '%s'\nGot:      '%s'", testData, decrypted)
	}
}
