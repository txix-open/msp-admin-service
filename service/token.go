package service

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type customClaims struct {
	Id int64
	jwt.StandardClaims
}

type Token struct {
	ttl       time.Duration
	secretKey string
}

func NewToken(ttl time.Duration, secretKey string) Token {
	return Token{
		ttl:       ttl,
		secretKey: secretKey,
	}
}

func (t Token) GenerateToken(id int64) (string, string, error) {
	expired := ""
	claims := customClaims{Id: id}

	if t.ttl != 0 {
		exp := time.Now().Add(t.ttl)
		expired = exp.String()
		claims.ExpiresAt = exp.Unix()
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.secretKey))
	if err != nil {
		return "", "", errors.WithMessage(err, "generate jwt with claims")
	}

	return tokenString, expired, nil
}

func (t Token) GetUserId(token string) (int64, error) {
	parsed, err := jwt.ParseWithClaims(token, &customClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secretKey), nil
	})
	if err != nil {
		return 0, errors.WithMessage(err, "parse jwt with claims")
	}

	if claims, ok := parsed.Claims.(*customClaims); ok && parsed.Valid {
		return claims.Id, nil
	}

	return 0, errors.New("token is invalid")
}
