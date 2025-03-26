// Signer 接口定义

package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidIssuer   = errors.New("invalid issuer") // ErrInvalidAudience = errors.New("invalid audience")
	ErrUnsupportedAlgo = errors.New("unsupported algorithm")
)

const (
	RsaKeyType  = "RS256"
	HMACKeyType = "HS256"
)

// TokenConfig token 配置
type TokenConfig struct {
	Algorithm              string        `json:"algorithm"`    // 签名算法: HS256/RS256/ES256
	SecretKey              string        `json:"secret_key"`   // HMAC密钥(至少32字节)
	PublicKey              string        `json:"public_key"`   // RSA/ECDSA公钥(PEM格式)
	PrivateKey             string        `json:"private_key"`  // RSA/ECDSA私钥(PEM格式)
	Expiration             time.Duration `json:"expiration"`   // Token有效期
	Issuer                 string        `json:"issuer"`       // 签发机构
	KeyRotation            time.Duration `json:"key_rotation"` // 密钥轮换间隔
	Audience               []string      `json:"audience"`     // 允许的受众
	KeyRotationGracePeriod time.Duration // 密钥轮换宽限期
	MaxRetiredKeys         int           // 最大保留旧密钥数量
}

// EnterpriseClaims 定制token携带参数可自行增加
type EnterpriseClaims struct {
	UserID   uint   `json:"uid"`
	ClientIP string `json:"cip"`
	DeviceID string `json:"did"`
	jwt.RegisteredClaims
}

// Signer 签名器规范
type Signer interface {
	Sign(claims EnterpriseClaims) (string, error)
	Verify(tokenString string) (*EnterpriseClaims, error)
	Algorithm() string
	KeyID() string
}
