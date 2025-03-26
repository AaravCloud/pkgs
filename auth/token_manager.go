// TokenManager 核心管理类

package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logc"
	"go.uber.org/zap"
)

// TokenManager 访问控制令牌管理器
type TokenManager struct {
	cfg         *TokenConfig
	signer      Signer
	RsaSigner   *RSASigner
	HMACSigner  *HMACSigner
	currentKey  crypto.PrivateKey
	oldKeys     map[string]crypto.PublicKey
	mu          sync.RWMutex
	keyExpiry   time.Time         // 密钥过期时间
	keyUsageMap map[string]uint64 // 密钥使用统计
}

// 初始化签名器
func (tm *TokenManager) initSigner() error {
	switch strings.ToUpper(tm.cfg.Algorithm) {
	case HMACKeyType:
		if len(tm.cfg.SecretKey) < 32 {
			return errors.New("HMAC密钥长度至少32字节")
		}
		hmacSigner := NewHMACSigner(tm.cfg)
		tm.signer = hmacSigner
	case RsaKeyType:
		rsaSigner, err := NewRSASigner(tm.cfg)
		if err != nil {
			return err
		}
		tm.signer = rsaSigner
	default:
		return ErrUnsupportedAlgo
	}
	return nil
}

// keyRotationDaemon 密钥轮换守护进程
func (tm *TokenManager) keyRotationDaemon(ctx context.Context) error {
	ticker := time.NewTicker(tm.cfg.KeyRotation)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return tm.rotateKey(ctx)
		case <-ctx.Done():
			logc.Info(ctx, "密钥轮换守护进程已停止")
			return nil
		}
	}
}

// NewTokenManager 创建令牌管理器实例
func NewTokenManager(ctx context.Context, cfg *TokenConfig) (*TokenManager, error) {
	fmt.Println("cfg:", cfg)
	tm := &TokenManager{
		cfg:         cfg,
		oldKeys:     make(map[string]crypto.PublicKey),
		keyUsageMap: make(map[string]uint64),
	}

	if err := tm.initSigner(); err != nil {
		logc.Errorf(ctx, "签名器初始化失败 [algorithm:%s]: %v", cfg.Algorithm, err)
		return nil, fmt.Errorf("签名器初始化失败: %w", err)
	}

	if cfg.KeyRotation > 0 {
		go func() {
			// 添加panic保护和错误日志
			defer func() {
				if r := recover(); r != nil {
					logc.Errorf(ctx, "密钥轮换发生panic: %v", r)
				}
			}()

			err := tm.keyRotationDaemon(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				logc.Errorf(ctx, "密钥轮换异常: %v", err)
			}
		}()
	}

	logc.Infof(ctx, "令牌管理器初始化完成 [algorithm:%s]", cfg.Algorithm)
	return tm, nil
}

func (tm *TokenManager) Generate(claims EnterpriseClaims) (string, error) {
	// 添加防御性检查
	if tm.signer == nil {
		return "", errors.New("签名器未初始化，请检查算法配置")
	}

	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    tm.cfg.Issuer,
		Audience:  tm.cfg.Audience,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.cfg.Expiration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token, err := tm.signer.Sign(claims)
	return token, err
}

// Verify 验证 token 并返回声明
func (tm *TokenManager) Verify(tokenString string) (*EnterpriseClaims, error) {
	return tm.signer.Verify(tokenString)
}

// rotateKeys 实际执行密钥轮换操作
func (tm *TokenManager) rotateKey(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// 记录旧密钥
	if tm.currentKey != nil {
		if rsaKey, ok := tm.currentKey.(*rsa.PrivateKey); ok {
			pubKey := &rsaKey.PublicKey
			keyID := fmt.Sprintf("%x", pubKey.N)[:16]
			tm.oldKeys[keyID] = pubKey
			logc.Info(ctx, "归档旧密钥",
				zap.String("key_id", keyID),
				zap.Time("expire_at", time.Now().Add(tm.cfg.KeyRotation*2)))
		}
	}

	if len(tm.oldKeys) >= tm.cfg.MaxRetiredKeys {
		// 淘汰最旧的密钥
		var oldestKey string
		for k := range tm.oldKeys {
			if oldestKey == "" || k < oldestKey {
				oldestKey = k
			}
		}
		delete(tm.oldKeys, oldestKey)
	}

	// 生成新密钥（RSA 2048）
	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logc.Error(ctx, "密钥生成失败",
			zap.Error(err),
			zap.String("algorithm", tm.cfg.Algorithm))
		tm.mu.Unlock()
		return err
	}

	// 更新当前密钥
	tm.currentKey = newPrivateKey
	tm.keyExpiry = time.Now().Add(tm.cfg.KeyRotation * 3) // 设置新密钥过期时间
	return nil
}

// PrivateKeyToPEM 转换RSA私钥为PEM格式
func PrivateKeyToPEM(key *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
}

// PublicKeyToPEM 转换RSA公钥为PEM格式
func PublicKeyToPEM(key *rsa.PublicKey) string {
	bytes, _ := x509.MarshalPKIXPublicKey(key)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytes,
	}))
}
