package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"
	"github.com/google/uuid"

	"github.com/ahmadraza100/dotlock/internal/crypto"
)

// Vault represents the top-level vault structure (versioned JSON).
type Vault struct {
	Version   int                `json:"version"`
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Profiles  map[string]Profile `json:"profiles"`
}

type Profile struct {
	Entries map[string]Entry `json:"entries"`
}

type Entry struct {
	ID        uuid.UUID `json:"id"`
	Value     string    `json:"value"` // base64-encoded encrypted bytes
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewEmptyVault returns an empty vault with version set.
func NewEmptyVault() Vault {
	return Vault{Version: 1, Profiles: map[string]Profile{}}
}

// AtomicWrite writes data to a temp file then renames it to path with perm.
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("open tmp: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("sync tmp: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}
	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	return nil
}

// MarshalAndEncryptVault marshals the vault JSON and encrypts using recipient.
func MarshalAndEncryptVault(v *Vault, recipient age.Recipient) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal vault: %w", err)
	}
	enc, err := crypto.Encrypt(b, recipient)
	if err != nil {
		return nil, fmt.Errorf("encrypt vault: %w", err)
	}
	return enc, nil
}

// LoadVault decrypts and unmarshals a vault from path using identity.
func LoadVault(path string, identity age.Identity) (Vault, error) {
	var v Vault
	b, err := os.ReadFile(path)
	if err != nil {
		return v, fmt.Errorf("read vault: %w", err)
	}
	plain, err := crypto.Decrypt(b, identity)
	if err != nil {
		return v, fmt.Errorf("cannot decrypt vault: %w", err)
	}
	if err := json.Unmarshal(plain, &v); err != nil {
		return v, fmt.Errorf("unmarshal vault: %w", err)
	}
	return v, nil
}

// SetEntry sets or updates an entry in the profile
// SetEntry encrypts the provided value with recipient and stores base64-encoded cipher in the vault.
func SetEntry(v *Vault, profile string, key string, value []byte, recipient age.Recipient) error {
	if len(key) > 256 {
		return fmt.Errorf("key too long")
	}
	if len(value) > 65536 {
		return fmt.Errorf("value too long")
	}
	prof, ok := v.Profiles[profile]
	if !ok {
		prof = Profile{Entries: map[string]Entry{}}
	}
	id := uuid.New()
	// encrypt the entry value with age recipient and base64 encode
	cipher, err := crypto.Encrypt(value, recipient)
	if err != nil {
		return fmt.Errorf("encrypt entry: %w", err)
	}
	enc := base64.StdEncoding.EncodeToString(cipher)
	now := time.Now().UTC()
	prof.Entries[key] = Entry{ID: id, Value: enc, CreatedAt: now, UpdatedAt: now}
	v.Profiles[profile] = prof
	v.UpdatedAt = now
	return nil
}

// GetEntry returns decrypted value bytes
// GetEntry decrypts the stored base64-encoded cipher using identity and returns plaintext bytes.
func GetEntry(v *Vault, profile, key string, identity age.Identity) ([]byte, error) {
	prof, ok := v.Profiles[profile]
	if !ok {
		return nil, fmt.Errorf("profile not found")
	}
	e, ok := prof.Entries[key]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	cipher, err := base64.StdEncoding.DecodeString(e.Value)
	if err != nil {
		return nil, fmt.Errorf("decode entry: %w", err)
	}
	plain, err := crypto.Decrypt(cipher, identity)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt entry: %w", err)
	}
	return plain, nil
}

// DeleteEntry removes a key
func DeleteEntry(v *Vault, profile, key string) error {
	prof, ok := v.Profiles[profile]
	if !ok {
		return fmt.Errorf("profile not found")
	}
	if _, exists := prof.Entries[key]; !exists {
		return fmt.Errorf("key not found")
	}
	delete(prof.Entries, key)
	v.Profiles[profile] = prof
	v.UpdatedAt = time.Now().UTC()
	return nil
}

// ListEntries returns sorted keys (deterministic order not required here)
func ListEntries(v *Vault, profile string) ([]string, error) {
	prof, ok := v.Profiles[profile]
	if !ok {
		return nil, fmt.Errorf("profile not found")
	}
	out := make([]string, 0, len(prof.Entries))
	for k := range prof.Entries {
		out = append(out, k)
	}
	return out, nil
}

// ProfileMap returns map[string][]byte of values for profile
// ProfileMap returns a map of decrypted values for a profile using identity.
func ProfileMap(v *Vault, profile string, identity age.Identity) (map[string][]byte, error) {
	prof, ok := v.Profiles[profile]
	if !ok {
		return nil, fmt.Errorf("profile not found")
	}
	out := make(map[string][]byte, len(prof.Entries))
	for k, e := range prof.Entries {
		cipher, err := base64.StdEncoding.DecodeString(e.Value)
		if err != nil {
			return nil, fmt.Errorf("decode entry %s: %w", k, err)
		}
		plain, err := crypto.Decrypt(cipher, identity)
		if err != nil {
			return nil, fmt.Errorf("decrypt entry %s: %w", k, err)
		}
		out[k] = plain
	}
	return out, nil
}

// CreateProfile creates an empty profile
func CreateProfile(v *Vault, name string) error {
	if _, ok := v.Profiles[name]; ok {
		return fmt.Errorf("profile exists")
	}
	v.Profiles[name] = Profile{Entries: map[string]Entry{}}
	v.UpdatedAt = time.Now().UTC()
	return nil
}

// DeleteProfile deletes a profile
func DeleteProfile(v *Vault, name string) error {
	if _, ok := v.Profiles[name]; !ok {
		return fmt.Errorf("profile not found")
	}
	delete(v.Profiles, name)
	v.UpdatedAt = time.Now().UTC()
	return nil
}
