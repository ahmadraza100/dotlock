package store

import (
	"path/filepath"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/google/uuid"
)

func TestAtomicWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.age")

	ident, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	recipient, err := age.ParseX25519Recipient(ident.Recipient().String())
	if err != nil {
		t.Fatalf("parse recipient: %v", err)
	}

	v := NewEmptyVault()
	v.ID = uuid.New()
	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt
	if err := SetEntry(&v, "dev", "FOO", []byte("bar"), recipient); err != nil {
		t.Fatalf("set entry: %v", err)
	}
	data, err := MarshalAndEncryptVault(&v, recipient)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := AtomicWrite(path, data, 0600); err != nil {
		t.Fatalf("atomic write: %v", err)
	}

	out, err := LoadVault(path, ident)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}
	val, err := GetEntry(&out, "dev", "FOO", ident)
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if string(val) != "bar" {
		t.Fatalf("unexpected value: %s", string(val))
	}
}
