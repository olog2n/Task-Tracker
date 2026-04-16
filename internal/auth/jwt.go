package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — JWT claims с user_id
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenPair — пара access + refresh токенов
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTService — сервис для работы с JWT
type JWTService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService создаёт новый JWTService
func NewJWTService(secretKey string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair генерирует пару access + refresh токенов
func (s *JWTService) GenerateTokenPair(userID int) (*TokenPair, error) {
	// Access token (короткоживущий)
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return nil, err
	}

	// Refresh token (долгоживущий)
	refreshClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
		ExpiresAt:    time.Now().Add(s.accessExpiry),
	}, nil
}

// RefreshAccessToken генерирует новый access токен из refresh токена
func (s *JWTService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	// Проверяем refresh токен
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Генерируем новую пару
	return s.GenerateTokenPair(claims.UserID)
}

// ValidateToken проверяет и парсит JWT токен
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
