package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"
)

// Golden: AES-256-CBC PKCS#7, hex(IV || ciphertext), fixed IV 00..0f.
// Key = config.yaml encryption.key; plaintext = "admin123".
const loginCipherGoldenAdmin123 = "000102030405060708090a0b0c0d0e0f17f1b26aff75e950ec141048626a9ed8"

func TestDecryptLoginPasswordCipher_golden(t *testing.T) {
	const keyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := InitEncryption(keyHex); err != nil {
		t.Fatal(err)
	}

	got, err := DecryptLoginPasswordCipher(loginCipherGoldenAdmin123)
	if err != nil {
		t.Fatal(err)
	}
	if got != "admin123" {
		t.Fatalf("got %q want admin123", got)
	}
}

func TestDecryptLoginPasswordCipher_rebuildMatchesGolden(t *testing.T) {
	const keyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := InitEncryption(keyHex); err != nil {
		t.Fatal(err)
	}

	iv := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	padded := pkcs7Pad([]byte("admin123"), aes.BlockSize)
	key, _ := hex.DecodeString(keyHex)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ct := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ct, padded)
	built := hex.EncodeToString(append(append([]byte{}, iv...), ct...))
	if built != loginCipherGoldenAdmin123 {
		t.Fatalf("rebuild %s != golden %s", built, loginCipherGoldenAdmin123)
	}
}
