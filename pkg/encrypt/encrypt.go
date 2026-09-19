package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func getMobileAesKey() []byte {
	key := os.Getenv("MOBILE_AES_KEY")
	if key == "" {
		key = "5A2E746B08D846502F37A6E2D85D583B"
	}
	return []byte(key)
}

// HashPassword creates a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(password)), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks a password against a hash.
// Returns (valid, needsUpgrade).
// needsUpgrade is true when the password is correct but stored with the old MD5 scheme.
func VerifyPassword(password, hash string) (bool, bool) {
	trimmed := strings.TrimSpace(password)

	// bcrypt hashes start with $2a$
	if strings.HasPrefix(hash, "$2a$") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(trimmed))
		return err == nil, false
	}

	// Legacy MD5 verification
	if MD5Password(trimmed) == hash {
		return true, true // correct but needs upgrade
	}
	return false, false
}

// MD5Password is the legacy password hasher — kept for migration compatibility.
func MD5Password(password string) string {
	const passwordEncryptSeed = "(ThinkTalk)@#$"
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(password+passwordEncryptSeed))))
}

// EncMobile encrypts a mobile number using AES-ECB with PKCS7 padding.
// Kept as ECB for compatibility with existing database records.
// TODO: migrate to AES-GCM with a version field, then switch encryption here.
func EncMobile(mobile string) (string, error) {
	key := getMobileAesKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plaintext := []byte(mobile)
	// PKCS7 padding
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padtext := make([]byte, len(plaintext)+padding)
	copy(padtext, plaintext)
	for i := len(plaintext); i < len(padtext); i++ {
		padtext[i] = byte(padding)
	}
	ciphertext := make([]byte, len(padtext))
	for i := 0; i < len(padtext); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], padtext[i:i+aes.BlockSize])
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecMobile decrypts a mobile number. Supports both legacy ECB and new GCM formats.
func DecMobile(mobile string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(mobile)
	if err != nil {
		return "", err
	}
	key := getMobileAesKey()

	// Try GCM first
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) >= nonceSize {
		plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
		if err == nil {
			return string(plaintext), nil
		}
	}

	// Fallback to legacy ECB
	block, err = aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(data)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	plaintext := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}
	// Remove PKCS7 padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > aes.BlockSize || padding == 0 {
		return "", errors.New("invalid padding")
	}
	return string(plaintext[:len(plaintext)-padding]), nil
}

func Md5Sum(data []byte) string {
	return hex.EncodeToString(byte16ToBytes(md5.Sum(data)))
}

func byte16ToBytes(in [16]byte) []byte {
	tmp := make([]byte, 16)
	for _, value := range in {
		tmp = append(tmp, value)
	}
	return tmp[16:]
}
