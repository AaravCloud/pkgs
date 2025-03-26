// RSASigner 实现

package token

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"

	"github.com/golang-jwt/jwt/v5"
)

type RSASigner struct {
	cfg        *TokenConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
}

func NewRSASigner(cfg *TokenConfig) (*RSASigner, error) {
	privateKey, publicKey, err := parseRSAKeys(cfg)
	if err != nil {
		return nil, fmt.Errorf("RSA密钥解析失败: %w", err)
	}

	return &RSASigner{
		cfg:        cfg,
		privateKey: privateKey,
		publicKey:  publicKey,
		keyID:      generateKeyID(publicKey),
	}, nil
}

// 生成符合RFC 7638标准的JWK指纹作为KeyID
func generateKeyID(pubKey *rsa.PublicKey) string {
	jwk := map[string]string{
		"kty": "RSA",
		"n":   base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes()),
	}

	data, _ := json.Marshal(jwk)
	hash := sha256.Sum256(data)

	return base64.RawURLEncoding.EncodeToString(hash[:16])
}

// 安全解析RSA密钥对（支持PKCS#1和PKCS#8格式）
func parseRSAKeys(cfg *TokenConfig) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if cfg == nil {
		return nil, nil, errors.New("配置对象不能为空")
	}

	privateKeyBlock, _ := pem.Decode([]byte(cfg.PrivateKey))
	if privateKeyBlock == nil {
		return nil, nil, errors.New("无效的PEM私钥格式")
	}

	var privateKey *rsa.PrivateKey
	var err error

	if privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes); err != nil {
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("PKCS#1和PKCS#8解析均失败: %w", err)
		}

		var ok bool
		if privateKey, ok = pkcs8Key.(*rsa.PrivateKey); !ok {
			return nil, nil, errors.New("PKCS#8密钥不是RSA类型")
		}
	}

	publicKeyBlock, _ := pem.Decode([]byte(cfg.PublicKey))
	if publicKeyBlock == nil {
		return nil, nil, errors.New("无效的PEM公钥格式")
	}

	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("公钥解析失败: %w", err)
	}

	rsaPublicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("公钥不是RSA类型")
	}

	// 密钥对匹配验证
	if privateKey.PublicKey.N.Cmp(rsaPublicKey.N) != 0 ||
		privateKey.PublicKey.E != rsaPublicKey.E {
		return nil, nil, errors.New("公私钥不匹配")
	}

	return privateKey, rsaPublicKey, nil
}

// ValidateRSAKeys 在解析后添加密钥安全检查
func ValidateRSAKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) error {
	if privateKey.Size() < 256 { // 2048 bits
		return errors.New("密钥长度不符合安全要求")
	}

	if publicKey.E != 65537 {
		return errors.New("非标准RSA指数")
	}

	return nil
}

func (s *RSASigner) Sign(claims EnterpriseClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *RSASigner) Verify(tokenString string) (*EnterpriseClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &EnterpriseClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, fmt.Errorf("%w: %v", ErrInvalidToken, "token格式错误")
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, fmt.Errorf("%w: %v", ErrInvalidToken, "token尚未生效")
		default:
			return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err.Error())
		}
	}

	if claims, ok := token.Claims.(*EnterpriseClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *RSASigner) Algorithm() string {
	return s.cfg.Algorithm
}
func (s *RSASigner) KeyID() string {
	return s.keyID
}
