package service

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/integration-system/isp-lib/v2/config"
	"github.com/pkg/errors"
	"msp-admin-service/conf"
)

type customClaims struct {
	Id int64
	jwt.StandardClaims
}

func GenerateToken(id int64) (string, string, error) {
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

func GetUserId(token string) (int64, error) {
	cfg := config.GetRemote().(*conf.RemoteConfig)
	parsed, err := jwt.ParseWithClaims(token, &customClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.SecretKey), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := parsed.Claims.(*customClaims); ok && parsed.Valid {
		return claims.Id, nil
	} else {
		return 0, errors.New("token is invalid")
	}
}
