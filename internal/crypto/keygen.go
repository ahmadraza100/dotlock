package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/google/uuid"
)

// GenerateIdentity creates a new age identity and returns the identity, its recipient, and an id.
func GenerateIdentity() (age.Identity, age.Recipient, uuid.UUID, error) {
	ident, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("generate identity: %w", err)
	}
	rec, err := age.ParseX25519Recipient(ident.Recipient().String())
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("parse recipient: %w", err)
	}
	id := uuid.New()
	return ident, rec, id, nil
}

// LoadDefaultRecipientAndIdentity looks for keys in ~/.config/dotlock/keys and returns recipient+identity.
// If no key exists, it creates and saves a new default identity automatically.
func LoadDefaultRecipientAndIdentity() (age.Recipient, age.Identity, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot determine home dir: %w", err)
	}
	keyDir := filepath.Join(home, ".config", "dotlock", "keys")

	// Try to load existing identity
	files, err := os.ReadDir(keyDir)
	if err == nil {
		for _, fi := range files {
			if fi.IsDir() {
				continue
			}
			p := filepath.Join(keyDir, fi.Name())
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			id, err := age.ParseX25519Identity(string(b))
			if err != nil {
				continue
			}
			rec, err := age.ParseX25519Recipient(id.Recipient().String())
			if err != nil {
				continue
			}
			return rec, id, nil
		}
	}

	// No identity found; create and save a default one
	ident, rec, _, err := GenerateIdentity()
	if err != nil {
		return nil, nil, fmt.Errorf("generate default identity: %w", err)
	}
	if err := SaveIdentityFile(filepath.Join(keyDir, "default.agekey"), ident, 0600); err != nil {
		return nil, nil, fmt.Errorf("save default identity: %w", err)
	}
	return rec, ident, nil
}
