package utils

type TokenType string

const (
	ACCESS  TokenType = "ACCESS"
	REFRESH TokenType = "REFRESH"
)

type Header struct {
	Alg       string    `json:"alg"`
	Typ       string    `json:"typ"`
	Kid       string    `json:"kid"`
	TokenType TokenType `json:"token_type"`
}

type AccessTokenPayload struct {
	Jti      string `json:"jti"`
	UserUid  string `json:"user_uid"`
	Iss      string `json:"iss"`
	Iat      int64  `json:"iat"`
	Exp      int64  `json:"exp"`
	DeviceId string `json:"device_id"`
}

type RefreshTokenPayload struct {
	Jti      string `json:"jti"`
	UserUid  string `json:"user_uid"`
	Iss      string `json:"iss"`
	Iat      int64  `json:"iat"`
	Exp      int64  `json:"exp"`
	DeviceId string `json:"device_id"`
}
