package helper

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	DBConnList  []gorm.DB
	DBErr       error
	RedisClient *redis.Client
	RedisError  error
)

type (
	Database struct {
		Username string
		Password string
		Host     string
		DBList   map[string]string
		Port     int
		SSLMode  string
		Timezone string
	}

	Redis struct {
		RedisAddress string
		Password     string
	}
)
