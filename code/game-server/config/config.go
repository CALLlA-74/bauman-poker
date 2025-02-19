package config

const (
	// Параметры ключей шифрования
	KeyType                = "RSA"
	SigningAlg             = "RS256"
	CryptoKeyExpPeriodDays = 365

	// Параметры токенов
	AccessTokenExpMinutes = 60
	RefreshTokenExpHours  = 72
	TokenStandard         = "JWT"
	LeewaySeconds         = 60

	GroupName          = "/poker/v1"
	HostPort           = 8080
	HealthCheckHandler = "/manage/health"

	// CircuitBreaker params
	Timeout       = 1000
	MaxNumOfFails = 2

	// IdentityExterService config
	IdentityExterBaseUrl = "http://identity-provider:8090"
	IdentityGroupName    = "/api/v1"
	SignUpHandler        = "/register"
	AuthHandler          = "/oauth/token"
	LogoutHandler        = "/oauth/revoke"
	JWKsHandler          = "/.well-known/jwks.json"

	// параметры WebSocket
	PingPeriodMilli int64 = 2000 //500 * (2 * 5 * 60) // период пингов в мс
	MsgLeewayMilli  int64 = 100  // допустимая задержка сообщения
)
