package jwt

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

func TestBuildTokens(t *testing.T) {
	opt := TokenOptions{
		AccessSecret: "test-secret",
		AccessExpire: 3600,
		Fields: map[string]interface{}{
			"userId": float64(123),
		},
	}

	token, err := BuildTokens(opt)
	if err != nil {
		t.Fatalf("BuildTokens err: %v", err)
	}
	if token.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
	if token.AccessExpire == 0 {
		t.Error("AccessExpire is zero")
	}

	// verify token can be parsed
	parsed, err := jwt.Parse(token.AccessToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(opt.AccessSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Error("parsed token is not valid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("claims not MapClaims")
	}
	if v, ok := claims["userId"]; !ok || v.(float64) != 123 {
		t.Errorf("userId claim mismatch: %v", v)
	}
}

func TestBuildTokensEmptySecret(t *testing.T) {
	opt := TokenOptions{
		AccessSecret: "",
		AccessExpire: 3600,
	}

	token, err := BuildTokens(opt)
	if err != nil {
		t.Fatalf("BuildTokens err: %v", err)
	}
	if token.AccessToken == "" {
		t.Error("AccessToken should not be empty even with empty secret")
	}
}
