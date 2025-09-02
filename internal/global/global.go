package global

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"plane_war/internal/config"
)

var (
	Config   *config.Config
	DB       *gorm.DB
	Log      *logrus.Logger
	MySqlLog logger.Interface
	Redis    *redis.Client
)
