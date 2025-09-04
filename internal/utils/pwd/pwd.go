package pwd

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

// HashPassword 对密码进行哈希加密
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("密码加密失败", err)
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePasswords 比较输入的密码和数据库存储的密码是否匹配
func ComparePasswords(storedPassword, enteredPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(enteredPassword))
	if err != nil {
		log.Println("密码比对失败", err)
		return false
	}
	return true
}
