package crypto

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

var testKey = []byte("01234567890123456789012345678901") // 32 bytes

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	plaintext := []byte("hello, coin")

	ct1, err := Encrypt(testKey, plaintext)
	if err != nil {
		t.Fatalf("Encrypt #1 failed: %v", err)
	}

	ct2, err := Encrypt(testKey, plaintext)
	if err != nil {
		t.Fatalf("Encrypt #2 failed: %v", err)
	}

	if ct1 == ct2 {
		t.Errorf("expected two encryptions of the same plaintext to produce different ciphertexts (random nonce), but they were identical")
	}

	got, err := Decrypt(testKey, ct1)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Errorf("Decrypt returned %q, want %q", got, plaintext)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	wrongKey := []byte("99999999999999999999999999999999") // different 32-byte key

	ct, err := Encrypt(testKey, []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(wrongKey, ct)
	if !errors.Is(err, ErrInvalidCiphertext) {
		t.Errorf("Decrypt with wrong key: got error %v, want ErrInvalidCiphertext", err)
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	ct, err := Encrypt(testKey, []byte("tamper me"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(ct)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}

	// Flip a byte in the ciphertext portion (after the 12-byte nonce).
	const nonceSize = 12
	raw[nonceSize] ^= 0xff

	tampered := base64.StdEncoding.EncodeToString(raw)

	_, err = Decrypt(testKey, tampered)
	if !errors.Is(err, ErrInvalidCiphertext) {
		t.Errorf("Decrypt with tampered ciphertext: got error %v, want ErrInvalidCiphertext", err)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	_, err := Decrypt(testKey, "this is not base64!!!")
	if !errors.Is(err, ErrInvalidCiphertext) {
		t.Errorf("Decrypt with invalid base64: got error %v, want ErrInvalidCiphertext", err)
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	// 4 bytes — shorter than the 12-byte AES-GCM nonce.
	short := base64.StdEncoding.EncodeToString([]byte{0x01, 0x02, 0x03, 0x04})

	_, err := Decrypt(testKey, short)
	if !errors.Is(err, ErrInvalidCiphertext) {
		t.Errorf("Decrypt with too-short input: got error %v, want ErrInvalidCiphertext", err)
	}
}
