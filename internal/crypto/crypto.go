package crypto

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"
)

// ZeroBytes overwrites b with zeros. Call with defer when handling secret bytes.
func ZeroBytes(b []byte) {
	if b == nil {
		return
	}
	for i := range b {
		b[i] = 0
	}
}

// Encrypt encrypts plaintext with recipient and returns age encrypted bytes.
func Encrypt(plain []byte, recipient age.Recipient) ([]byte, error) {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, fmt.Errorf("age encrypt: %w", err)
	}
	if _, err := w.Write(plain); err != nil {
		w.Close()
		return nil, fmt.Errorf("write to age writer: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close age writer: %w", err)
	}
	return buf.Bytes(), nil
}

// Decrypt decrypts age bytes using an identity and returns plaintext.
func Decrypt(data []byte, identity age.Identity) ([]byte, error) {
	r := bytes.NewReader(data)
	rd, err := age.Decrypt(r, identity)
	if err != nil {
		return nil, fmt.Errorf("age decrypt: %w", err)
	}
	out, err := io.ReadAll(rd)
	if err != nil {
		return nil, fmt.Errorf("read decrypted: %w", err)
	}
	return out, nil
}

// SaveIdentityFile writes the identity string to path with mode perm.
func SaveIdentityFile(path string, identity age.Identity, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("mkdir keys dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("open identity file: %w", err)
	}
	defer f.Close()
	var out []byte
	if tm, ok := identity.(encoding.TextMarshaler); ok {
		b, err := tm.MarshalText()
		if err != nil {
			return fmt.Errorf("marshal identity: %w", err)
		}
		out = b
	} else if s, ok := identity.(fmt.Stringer); ok {
		out = []byte(s.String())
	} else {
		out = []byte(fmt.Sprintf("%v", identity))
	}
	if _, err := f.Write(out); err != nil {
		return fmt.Errorf("write identity: %w", err)
	}
	return nil
}
