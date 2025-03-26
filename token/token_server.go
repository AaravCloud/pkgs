package token

import (
	"context"
	"sync"
)

// TokenService 令牌服务封装
type TokenService struct {
	tokenManager *TokenManager
}

var (
	instance *TokenService
	once     sync.Once
)

// NewTokenService token服务
func NewTokenService(cfg *TokenConfig) (*TokenService, error) {
	var initErr error
	once.Do(func() {
		tm, err := NewTokenManager(context.Background(), cfg)
		if err != nil {
			initErr = err
			return
		}
		instance = &TokenService{tokenManager: tm}
	})

	if initErr != nil {
		return nil, initErr
	}
	return instance, nil
}

// GenerateDeviceToken 生成设备令牌
func (ts *TokenService) GenerateDeviceToken(userID uint, clientIP, deviceID string) (string, error) {
	claims := EnterpriseClaims{
		UserID:   userID,
		ClientIP: clientIP,
		DeviceID: deviceID,
	}
	return ts.tokenManager.Generate(claims)
}

// VerifyAndParse 验证并解析令牌
func (ts *TokenService) VerifyAndParse(tokenString string) (*EnterpriseClaims, error) {
	return ts.tokenManager.Verify(tokenString)
}

// RotateKeys 安全密钥轮换
func (ts *TokenService) RotateKeys(ctx context.Context) error {
	return ts.tokenManager.rotateKey(ctx)
}
