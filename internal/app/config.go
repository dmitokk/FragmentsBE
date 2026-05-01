package app

type Config struct {
	HTTPPort string

	DBUrl string

	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucket     string
	MinioUseSSL     bool
	MinioPublicURL  string

	JWTSecret string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}