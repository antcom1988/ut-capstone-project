package auth

import (
	"errors"
	"sikapj-backend/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
 SessionVersion int `json:"session_version"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	NIP      string `json:"nip,omitempty"`
	jwt.RegisteredClaims
}

// HashPassword mengenkripsi plain password dengan bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash memvalidasi kesesuaian hash password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWTToken membuat token JWT baru
func GenerateJWTToken(userID int, username, role, nip string, cfg *config.Config, sessionVersion ...int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpirationHours) * time.Hour)
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		NIP:      nip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sikapj-system",
		},
	}

	if len(sessionVersion)>0 { claims.SessionVersion=sessionVersion[0] }
 token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ValidateJWTToken memverifikasi token JWT dari header
func ValidateJWTToken(tokenString string, cfg *config.Config) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("metode tanda tangan token tidak valid")
		}
		return []byte(cfg.JWTSecret), nil
	}, jwt.WithIssuer("sikapj-system"), jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}
