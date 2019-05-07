package token

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/integration-system/isp-lib/config"
	"msp-admin-service/conf"
	"time"
)

type customClaims struct {
	id string
	jwt.StandardClaims
}

func GetToken(id string) (string, string) {
	expired := ""
	claims := customClaims{id: id}
	cfg := config.GetRemote().(*conf.RemoteConfig)
	if cfg.ExpirySec != 0 {
		exp := time.Now().Add(time.Second * time.Duration(cfg.ExpirySec))
		expired = exp.String()
		claims.ExpiresAt = exp.Unix()
	}
	tokenString, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.SecretKey))
	return tokenString, expired
}
