package jwts

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"plane_war/internal/global"
)

// 解析 Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	MySecret = []byte(global.Config.Auth.AccessSecret)
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return MySecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
