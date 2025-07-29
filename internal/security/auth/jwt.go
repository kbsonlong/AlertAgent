package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Perms    []string `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken 生成JWT令牌
func (manager *JWTManager) GenerateToken(userID, username string, roles, permissions []string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		Perms:    permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "alertagent",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(manager.secretKey)
}

// VerifyToken 验证JWT令牌
func (manager *JWTManager) VerifyToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}
			return manager.secretKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshToken 刷新JWT令牌
func (manager *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := manager.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查令牌是否即将过期（剩余时间少于1小时）
	if time.Until(claims.RegisteredClaims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not eligible for refresh")
	}

	return manager.GenerateToken(claims.UserID, claims.Username, claims.Roles, claims.Perms)
}