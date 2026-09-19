package encrypt

import (
	"testing"
)

func TestMD5Password(t *testing.T) {
	result := MD5Password("test123")
	if result == "" {
		t.Error("MD5Password returned empty string")
	}
	if len(result) != 32 {
		t.Errorf("MD5Password hash length = %d, want 32", len(result))
	}

	// deterministic
	result2 := MD5Password("test123")
	if result != result2 {
		t.Error("MD5Password should be deterministic")
	}

	// different passwords should produce different hashes
	result3 := MD5Password("test456")
	if result == result3 {
		t.Error("different passwords should produce different hashes")
	}
}

func TestEncMobileDecMobile(t *testing.T) {
	mobile := "13800138000"

	encrypted, err := EncMobile(mobile)
	if err != nil {
		t.Fatalf("EncMobile err: %v", err)
	}
	if encrypted == "" {
		t.Error("EncMobile returned empty string")
	}
	if encrypted == mobile {
		t.Error("EncMobile did not encrypt")
	}

	decrypted, err := DecMobile(encrypted)
	if err != nil {
		t.Fatalf("DecMobile err: %v", err)
	}
	if decrypted != mobile {
		t.Errorf("round-trip failed: got %q, want %q", decrypted, mobile)
	}
}

func TestEncMobileEmpty(t *testing.T) {
	result, err := EncMobile("")
	if err != nil {
		t.Fatalf("EncMobile err: %v", err)
	}
	if result == "" {
		t.Error("EncMobile empty string returned empty")
	}
}

func TestDecMobileInvalidBase64(t *testing.T) {
	_, err := DecMobile("!!!invalid!!!")
	if err == nil {
		t.Error("DecMobile should error on invalid base64")
	}
}

func TestMd5Sum(t *testing.T) {
	result := Md5Sum([]byte("hello"))
	if len(result) != 32 {
		t.Errorf("Md5Sum length = %d, want 32", len(result))
	}

	// deterministic
	result2 := Md5Sum([]byte("hello"))
	if result != result2 {
		t.Error("Md5Sum should be deterministic")
	}
}
