# 身份验证服务使用指南

## 1. 服务配置

```go
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log"
	"test/auth"
	"time"

	"github.com/gin-gonic/gin"
)

var cfg = &auth.TokenConfig{
    Expiration:  time.Hour * 8,   // Token有效期
    KeyRotation: time.Hour * 24,  // 密钥轮换间隔
    MaxRetiredKeys: 3,           // 保留历史密钥数量
}

func initRsaAlgorithm() {
    // 生成2048位RSA密钥对
    	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
      // 密钥安全检查
	cfg.Algorithm = auth.RsaKeyType
	err = auth.ValidateRSAKeys(privateKey, &privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
        // 转换为PEM格式
	cfg.PrivateKey = auth.PrivateKeyToPEM(privateKey)
	cfg.PublicKey = auth.PublicKeyToPEM(&privateKey.PublicKey)
}

func initHmacAlgorithm() {
    // 生成32字节随机密钥
  	secret := make([]byte, 32) 
	if _, err := rand.Read(secret); err != nil {
		panic(fmt.Sprintf("HMAC密钥生成失败: %v", err))
	}
    
    cfg.Algorithm = auth.HMACKeyType
    cfg.SecretKey = base64.StdEncoding.EncodeToString(secret)
}

func main() {
	r := gin.Default()
	r.GET("/get/token", func(c *gin.Context) {
		//initRsaAlgorithm()//rsa加密
		initHmacAlgorithm()//hmac加密
		// 初始化令牌服务（单例模式）
		ts, err := auth.NewTokenService(cfg)
		if err != nil {
			log.Fatalf("令牌服务初始化失败: %v", err)
		}
		// 生成设备令牌
		token, err := ts.GenerateDeviceToken(1001, "192.168.1.100", "DEV-2024-M1")
		if err != nil {
			log.Fatalf("令牌生成失败: %v", err)
		}
		fmt.Println("生成的令牌:")
		fmt.Println(token)
		c.String(200, "服务运行正常")
	})

	r.GET("/check/token", func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		fmt.Println(token)
		// 初始化令牌服务（单例模式）
		ts, err := auth.NewTokenService(cfg)
		if err != nil {
			log.Fatalf("令牌服务初始化失败: %v", err)
		}
		// 验证并解析令牌
		claims, err := ts.VerifyAndParse(token)
		if err != nil {
			log.Fatalf("令牌验证失败: %v", err)
		}

		fmt.Printf("验证成功：用户ID %d，设备 %s\n", claims.UserID, claims.DeviceID)

		fmt.Println(token)

	})

	r.Run(":8080")
}