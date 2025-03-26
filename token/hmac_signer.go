package token

import "github.com/golang-jwt/jwt/v5"

// HMACSigner  HMAC签名器
type HMACSigner struct {
	cfg    *TokenConfig
	secret []byte
}

func NewHMACSigner(cfg *TokenConfig) *HMACSigner {
	return &HMACSigner{
		cfg:    cfg,
		secret: []byte(cfg.SecretKey),
	}
}
func (h *HMACSigner) Sign(claims EnterpriseClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.secret)
}

func (h *HMACSigner) Verify(tokenString string) (*EnterpriseClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &EnterpriseClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*EnterpriseClaims), nil
}

func (h *HMACSigner) Algorithm() string {
	return h.cfg.Algorithm
}

func (h *HMACSigner) KeyID() string {
	return ""
}
