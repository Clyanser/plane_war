package jwts

import "github.com/golang-jwt/jwt/v4"

// JwtPayload 保存用户信息
type JwtPayload struct {
	UserID   uint   `json:"userid"`
	Nickname string `json:"nickname"`
}

var MySecret []byte

// CustomClaims 自定义 Claims，使用 RegisteredClaims 代替 StandardClaims
type CustomClaims struct {
	JwtPayload
	jwt.RegisteredClaims
}
