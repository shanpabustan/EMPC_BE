package mdlDataEncryptionV1

var ()

type (
	DatabaseData struct {
		SecretKey string `json:"secret_key,omitempty"`
		DBHost    string `json:"db_host"`
		DBName    string `json:"db_name"`
		DBUser    string `json:"db_user"`
		DBPass    string `json:"db_pass"`
	}
)
