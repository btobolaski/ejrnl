package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

// generateKey uses a key derivation function to create an key for use in aes encryption.
func GenerateKey(password, salt []byte, pow uint) ([]byte, error) {

	workFactor := 1
	workFactor <<= pow
	return scrypt.Key(password, salt, workFactor, 8, 1, 16)
}

// encrypt encrypts the data using aes. Note that the key must be 16 or 32 bytes.
// the output is in this format {{nonce}}{{null}}{{null}}{{ciphertext}}
func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	nounce := make([]byte, aead.NonceSize())
	_, err = rand.Read(nounce)
	if err != nil {
		return []byte{}, err
	}

	cyphertext := aead.Seal(data[:0], nounce, data, []byte{})
	return append(append(nounce, []byte("\x00\x00")...), cyphertext...), nil
}

// decrypt decrypts the passed in data. The data is expected to be in the format {{nonce}}{{null}}{{null}}{{ciphertext}}
func Decrypt(cyphertext, key []byte) ([]byte, error) {
	parts := bytes.SplitN(cyphertext, []byte("\x00\x00"), 2)
	if len(parts) != 2 {
		return []byte{}, fmt.Errorf("cyphertext isn't in the proper format")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	dest := make([]byte, len(parts[1]))
	nounce := parts[0]
	plaintext, err := aead.Open(dest, nounce, parts[1], []byte{})

	// aead pads the input data with null values at the beginning. Applications won't care about this
	// padding so, we just remove them
	return bytes.Replace(plaintext, []byte("\x00"), []byte{}, -1), err
}
