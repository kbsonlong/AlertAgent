package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// EncryptionManager 加密管理器
type EncryptionManager struct {
	key []byte
}

// NewEncryptionManager 创建加密管理器
func NewEncryptionManager(password string, salt []byte) (*EncryptionManager, error) {
	if len(salt) == 0 {
		return nil, errors.New("salt cannot be empty")
	}

	// 使用PBKDF2派生密钥
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)

	return &EncryptionManager{
		key: key,
	}, nil
}

// NewEncryptionManagerWithScrypt 使用Scrypt创建加密管理器
func NewEncryptionManagerWithScrypt(password string, salt []byte) (*EncryptionManager, error) {
	if len(salt) == 0 {
		return nil, errors.New("salt cannot be empty")
	}

	// 使用Scrypt派生密钥（更安全但更慢）
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return &EncryptionManager{
		key: key,
	}, nil
}

// Encrypt 加密数据
func (em *EncryptionManager) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(em.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 返回base64编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (em *EncryptionManager) Decrypt(ciphertext string) ([]byte, error) {
	// 解码base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(em.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString 加密字符串
func (em *EncryptionManager) EncryptString(plaintext string) (string, error) {
	return em.Encrypt([]byte(plaintext))
}

// DecryptString 解密字符串
func (em *EncryptionManager) DecryptString(ciphertext string) (string, error) {
	plaintext, err := em.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// GenerateSalt 生成随机盐值
func GenerateSalt(length int) ([]byte, error) {
	if length <= 0 {
		length = 32 // 默认32字节
	}

	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	return salt, nil
}

// HashPassword 哈希密码
func HashPassword(password string, salt []byte) (string, error) {
	if len(salt) == 0 {
		return "", errors.New("salt cannot be empty")
	}

	hash := pbkdf2.Key([]byte(password), salt, 10000, 64, sha256.New)
	return base64.StdEncoding.EncodeToString(hash), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hashedPassword string, salt []byte) (bool, error) {
	expectedHash, err := HashPassword(password, salt)
	if err != nil {
		return false, err
	}

	return expectedHash == hashedPassword, nil
}

// SecureConfig 安全配置结构
type SecureConfig struct {
	Encrypted map[string]string `json:"encrypted"`
	Salt      string            `json:"salt"`
}

// EncryptConfig 加密配置
func (em *EncryptionManager) EncryptConfig(config map[string]interface{}) (*SecureConfig, error) {
	encrypted := make(map[string]string)

	for key, value := range config {
		// 将值转换为JSON字符串
		valueStr := fmt.Sprintf("%v", value)
		encryptedValue, err := em.EncryptString(valueStr)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt %s: %w", key, err)
		}
		encrypted[key] = encryptedValue
	}

	// 生成新的盐值用于存储
	salt, err := GenerateSalt(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	return &SecureConfig{
		Encrypted: encrypted,
		Salt:      base64.StdEncoding.EncodeToString(salt),
	}, nil
}

// DecryptConfig 解密配置
func (em *EncryptionManager) DecryptConfig(secureConfig *SecureConfig) (map[string]string, error) {
	decrypted := make(map[string]string)

	for key, encryptedValue := range secureConfig.Encrypted {
		decryptedValue, err := em.DecryptString(encryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %w", key, err)
		}
		decrypted[key] = decryptedValue
	}

	return decrypted, nil
}