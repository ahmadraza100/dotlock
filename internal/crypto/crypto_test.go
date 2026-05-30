package crypto

import (
	"testing"

	"filippo.io/age"
)

func TestEncryptDecrypt(t *testing.T) {
	ident, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	recipient, err := age.ParseX25519Recipient(ident.Recipient().String())
	if err != nil {
		t.Fatalf("parse recipient: %v", err)
	}
	plain := []byte("secret-value")
	defer ZeroBytes(plain)
	enc, err := Encrypt(plain, recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	dec, err := Decrypt(enc, ident)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(dec) != "secret-value" {
		t.Fatalf("unexpected plaintext: %s", string(dec))
	}
}
