package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type TokenBlacklist interface {
	Add(token string, expiresAt time.Time) error
	IsBlacklisted(token string) (bool, error)
}

// internal/auth/jwt.go

type JWTService struct {
	algorithm     jwt.SigningMethod
	privateKey    interface{} // *ecdsa.PrivateKey или *rsa.PrivateKey
	publicKey     interface{} // *ecdsa.PublicKey или *rsa.PublicKey
	symmetricKey  []byte      // Для HS256
	keyID         string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	blacklist     TokenBlacklist
}

// NewJWTService создаёт сервис с поддержкой всех алгоритмов
func NewJWTService(algorithm, secret, privateKeyPEM, publicKeyPEM, keyID string,
	accessExpiry, refreshExpiry time.Duration) (*JWTService, error) {

	s := &JWTService{
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		keyID:         keyID,
	}

	switch strings.ToUpper(algorithm) {
	case "HS256":
		if len(secret) < 32 {
			return nil, fmt.Errorf("HS256 requires secret of at least 32 characters")
		}
		s.algorithm = jwt.SigningMethodHS256
		s.symmetricKey = []byte(secret)

	case "ES256":
		if privateKeyPEM == "" {
			return nil, fmt.Errorf("ES256 requires jwt_private_key")
		}
		s.algorithm = jwt.SigningMethodES256
		key, err := parseECDSAPrivateKey(privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EC private key: %w", err)
		}
		s.privateKey = key

		if publicKeyPEM != "" {
			pubKey, err := parseECDSAPublicKey(publicKeyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse EC public key: %w", err)
			}
			s.publicKey = pubKey
		} else {
			s.publicKey = &key.PublicKey
		}

	case "RS256":
		if privateKeyPEM == "" {
			return nil, fmt.Errorf("RS256 requires jwt_private_key")
		}
		s.algorithm = jwt.SigningMethodRS256
		key, err := parseRSAPrivateKey(privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
		}
		s.privateKey = key

		if publicKeyPEM != "" {
			pubKey, err := parseRSAPublicKey(publicKeyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
			}
			s.publicKey = pubKey
		} else {
			s.publicKey = &key.PublicKey
		}

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return s, nil
}

// GenerateTokenPair генерирует пару токенов
func (s *JWTService) GenerateTokenPair(userID int) (*TokenPair, error) {
	now := time.Now()

	// Access token
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(s.algorithm, accessClaims)
	if s.keyID != "" {
		accessToken.Header["kid"] = s.keyID
	}

	accessString, err := s.signToken(accessToken)
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(s.algorithm, refreshClaims)
	if s.keyID != "" {
		refreshToken.Header["kid"] = s.keyID
	}

	refreshString, err := s.signToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
		ExpiresAt:    now.Add(s.accessExpiry),
	}, nil
}

// signToken подписывает токен в зависимости от алгоритма
func (s *JWTService) signToken(token *jwt.Token) (string, error) {
	switch s.algorithm.(type) {
	case *jwt.SigningMethodHMAC:
		return token.SignedString(s.symmetricKey)
	case *jwt.SigningMethodECDSA:
		return token.SignedString(s.privateKey)
	case *jwt.SigningMethodRSA:
		return token.SignedString(s.privateKey)
	default:
		return "", fmt.Errorf("unsupported signing method")
	}
}

// ValidateToken проверяет токен
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что алгоритм совпадает с ожидаемым
		if token.Method.Alg() != s.algorithm.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}

		switch s.algorithm.(type) {
		case *jwt.SigningMethodHMAC:
			return s.symmetricKey, nil
		case *jwt.SigningMethodECDSA:
			return s.publicKey, nil
		case *jwt.SigningMethodRSA:
			return s.publicKey, nil
		default:
			return nil, fmt.Errorf("unsupported algorithm")
		}
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// internal/auth/jwt.go

// RefreshAccessToken генерирует новую пару токенов из refresh токена
// Работает одинаково для HS256, ES256, RS256
func (s *JWTService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	// 1. Проверяем blacklist (если есть)
	if s.blacklist != nil {
		listed, err := s.blacklist.IsBlacklisted(refreshToken)
		if err != nil {
			return nil, errors.New("token blacklist error")
		}
		if listed {
			return nil, errors.New("token has been revoked")
		}
	}

	// 2. Валидируем refresh токен (алгоритм проверяется внутри ValidateToken)
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 3. Добавляем старый токен в blacklist (rotation)
	if s.blacklist != nil {
		expiresAt := time.Now().Add(s.refreshExpiry)
		_ = s.blacklist.Add(refreshToken, expiresAt)
	}

	// 4. Генерируем новую пару токенов
	// 👇 Здесь используется текущий алгоритм (HS256/ES256/RS256)
	return s.GenerateTokenPair(claims.UserID)
}

// ===== Вспомогательные функции для парсинга ключей =====

func parseECDSAPrivateKey(pemStr string) (*ecdsa.PrivateKey, error) {
	// Убираем экранированные переносы строк (из env)
	pemStr = strings.ReplaceAll(pemStr, "\\n", "\n")

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return x509.ParseECPrivateKey(block.Bytes)
}

func parseECDSAPublicKey(pemStr string) (*ecdsa.PublicKey, error) {
	pemStr = strings.ReplaceAll(pemStr, "\\n", "\n")

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}

	return pubKey, nil
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = strings.ReplaceAll(pemStr, "\\n", "\n")

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	pemStr = strings.ReplaceAll(pemStr, "\\n", "\n")

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return pubKey, nil
}
