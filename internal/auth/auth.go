package auth

import (
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/utils"

	"github.com/dgrijalva/jwt-go/v4"
)

type UIDKey struct{}

type Claims struct {
	jwt.StandardClaims
	UID string
}

func CreateToken(uid string, cfg *config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(time.Duration(cfg.JWTExpireTiming))),
			IssuedAt:  jwt.At(time.Now()),
		},
		UID: uid,
	})
	return token.SignedString([]byte(cfg.SecretKey))
}

func ParseToken(token string, key []byte) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", utils.ErrNotAuthorized
		}
		return key, nil
	})

	if err != nil {
		return "", utils.ErrNotAuthorized
	}

	if claims, ok := parsedToken.Claims.(*Claims); ok && parsedToken.Valid {
		return claims.ID, nil
	}
	return "", utils.ErrNotAuthorized
}
