package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username   string `gorm:"size:32;unique" json:"username"`
	Password   string `gorm:"size:128" json:"-"`
	Nickname   string `gorm:"size:32" json:"nickname"`
	TotalScore int    `json:"total_score"`
}
