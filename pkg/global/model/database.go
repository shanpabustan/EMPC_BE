package model

type (
	Database struct {
		Username string
		Password string
		Host     string
		DBList   []string
		Port     int
		SSLMode  string
		Timezone string
	}

	Redis struct {
		RedisAddress string
		Password     string
	}
)
