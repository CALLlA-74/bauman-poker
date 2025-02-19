package config

const (
	// Параметры ключей шифрования
	NumOfKeys              = 3
	PriviteKeyLen          = 2048
	KeyType                = "RSA"
	SigningAlg             = "RS256"
	CryptoKeyExpPeriodDays = 365 // дней

	// Параметры хэширования паролей
	SaltSize      = 16  // длина хэш-соли
	NumOfHashIter = 1e5 // количество итарций хэширования
	KeySize       = 64  // длина хэша

	// Параметры токенов
	AccessTokenExpMinutes = 60
	RefreshTokenExpHours  = 72
	TokenStandard         = "JWT"
	LeewaySeconds         = 120

	GroupName = "/api/v1"
	HostPort  = 8090
	Hostname  = "http://identity-provider:8090/"
)
