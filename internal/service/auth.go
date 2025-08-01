package service

import (
	"alert_agent/internal/config"
	"alert_agent/internal/middleware"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	config *config.Config
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	cfg := config.GetConfig()
	return &AuthService{
		config: &cfg,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         User   `json:"user"`
}

// User 用户信息
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// GenerateTokens 生成访问令牌和刷新令牌
func (s *AuthService) GenerateTokens(user User) (*LoginResponse, error) {
	// 生成访问令牌
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.config.Gateway.Auth.TokenExpiry * 3600, // 转换为秒
		User:         user,
	}, nil
}

// generateAccessToken 生成访问令牌
func (s *AuthService) generateAccessToken(user User) (string, error) {
	claims := &middleware.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.Gateway.Auth.TokenExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "alert-agent",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Server.JWTSecret))
}

// generateRefreshToken 生成刷新令牌
func (s *AuthService) generateRefreshToken(user User) (string, error) {
	claims := &middleware.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.Gateway.Auth.RefreshExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "alert-agent",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Server.JWTSecret))
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(tokenString string) (*middleware.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Server.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*middleware.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshTokens 刷新令牌
func (s *AuthService) RefreshTokens(refreshToken string) (*LoginResponse, error) {
	// 验证刷新令牌
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 构建用户信息
	user := User{
		ID:       claims.UserID,
		Username: claims.Username,
		Roles:    claims.Roles,
	}

	// 生成新的令牌对
	return s.GenerateTokens(user)
}

// Login 用户登录
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// 从数据库查询用户
	var dbUser model.User
	err := database.DB.Where("username = ? AND status = ?", req.Username, "active").First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, errors.New("database error")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 构建用户信息
	user := User{
		ID:       dbUser.ID,
		Username: dbUser.Username,
		Email:    dbUser.Email,
		Roles:    []string{dbUser.Role},
	}

	return s.GenerateTokens(user)
}

// Logout 用户登出
func (s *AuthService) Logout(tokenString string) error {
	// 在实际应用中，这里应该将令牌加入黑名单
	// 或者从Redis中删除令牌
	return nil
}

// GetUserProfile 获取用户资料
func (s *AuthService) GetUserProfile(userID string) (*User, error) {
	// 查询用户
	var dbUser model.User
	err := database.DB.Where("id = ? AND status = ?", userID, "active").First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("database error")
	}

	// 返回用户信息
	user := &User{
		ID:       dbUser.ID,
		Username: dbUser.Username,
		Email:    dbUser.Email,
		Roles:    []string{dbUser.Role},
	}

	return user, nil
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(userID string, req UpdateProfileRequest) (*User, error) {
	// 查询用户
	var dbUser model.User
	err := database.DB.Where("id = ? AND status = ?", userID, "active").First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("database error")
	}

	// 更新邮箱
	if req.Email != "" {
		dbUser.Email = req.Email
	}

	// 保存更新
	err = database.DB.Save(&dbUser).Error
	if err != nil {
		return nil, errors.New("failed to update profile")
	}

	// 返回更新后的用户信息
	user := &User{
		ID:       dbUser.ID,
		Username: dbUser.Username,
		Email:    dbUser.Email,
		Roles:    []string{dbUser.Role},
	}

	return user, nil
}