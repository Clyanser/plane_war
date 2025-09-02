package jwts

import (
	"github.com/golang-jwt/jwt/v4"
	"plane_war/internal/global"
	"time"
)

func GenAccessToken(user JwtPayload) (string, error) {
	MySecret = []byte(global.Config.Auth.AccessSecret)
	Claims := CustomClaims{
		JwtPayload: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(global.Config.Auth.AccessExpire))),
			Issuer:    "plane_war",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	return token.SignedString(MySecret)
}
func GenRefreshToken(user JwtPayload) (string, error) {
	MySecret = []byte(global.Config.Auth.AccessSecret)
	Claims := CustomClaims{
		JwtPayload: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(global.Config.Auth.RefreshExpire))),
			Issuer:    "plane_war",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	return token.SignedString(MySecret)
}
