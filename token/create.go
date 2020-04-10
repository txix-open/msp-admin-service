package token

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/integration-system/isp-lib/v2/config"
	"msp-admin-service/conf"
)

type customClaims struct {
	Id string
	jwt.StandardClaims
}

func GetToken(id string) (string, string, error) {
	expired := ""
	claims := customClaims{Id: id}
	cfg := config.GetRemote().(*conf.RemoteConfig)
	if cfg.ExpireSec != 0 {
		exp := time.Now().Add(time.Second * time.Duration(cfg.ExpireSec))
		expired = exp.String()
		claims.ExpiresAt = exp.Unix()
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", "", err
	}
	return tokenString, expired, nil
}
