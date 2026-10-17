package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"sikapj-backend/internal/config"
	"testing"
	"time"
)

func TestValidateJWTRejectsUnexpectedSigningMethod(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", JWTExpirationHours: 1}
	claims := &JWTClaims{UserID: 1, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), Issuer: "sikapj-system"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	raw, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateJWTToken(raw, cfg); err == nil {
		t.Fatal("HS384 token should be rejected")
	}
}

func TestValidateJWTRejectsWrongIssuer(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", JWTExpirationHours: 1}
	claims := &JWTClaims{UserID: 1, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), Issuer: "other"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateJWTToken(raw, cfg); err == nil {
		t.Fatal("wrong issuer should be rejected")
	}
}
