package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type fileStore struct{}

var validKeyName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$`)

func keysDir() string {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = filepath.Join(os.Getenv("APPDATA"), "VoidVPN")
	case "darwin":
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, "Library", "Application Support", "voidvpn")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			base = filepath.Join(xdg, "voidvpn")
		} else {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".config", "voidvpn")
		}
	}
	return filepath.Join(base, "keys")
}

func keyFile(name string) (string, error) {
	if !validKeyName.MatchString(name) {
		return "", fmt.Errorf("invalid key name %q: must be alphanumeric with hyphens or underscores", name)
	}
	path := filepath.Join(keysDir(), name+".key")
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(keysDir())) {
		return "", fmt.Errorf("invalid key name: path traversal detected")
	}
	return path, nil
}

func deriveKey() ([]byte, error) {
	dir := keysDir()
	saltPath := filepath.Join(dir, ".salt")

	// Read or generate salt
	salt, err := os.ReadFile(saltPath)
	if err != nil {
		// Generate new random salt
		salt = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
		if err := os.WriteFile(saltPath, salt, 0600); err != nil {
			return nil, err
		}
	}

	hostname, _ := os.Hostname()
	home, _ := os.UserHomeDir()
	material := fmt.Sprintf("voidvpn:%s:%s", hostname, home)

	// HMAC-like construction: SHA256(salt || material)
	combined := append(salt, []byte(material)...)
	hash := sha256.Sum256(combined)
	return hash[:], nil
}

func (f *fileStore) Store(name string, key string) error {
	if err := os.MkdirAll(keysDir(), 0700); err != nil {
		return err
	}

	path, err := keyFile(name)
	if err != nil {
		return err
	}

	dk, err := deriveKey()
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(dk)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(key), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return os.WriteFile(path, []byte(encoded), 0600)
}

func (f *fileStore) Load(name string) (string, error) {
	path, err := keyFile(name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", err
	}

	dk, err := deriveKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(dk)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed")
	}

	return string(plaintext), nil
}

func (f *fileStore) Delete(name string) error {
	path, err := keyFile(name)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (f *fileStore) Exists(name string) bool {
	path, err := keyFile(name)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
