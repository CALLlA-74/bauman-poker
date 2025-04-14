package repo

type JWKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

type JWKResponse struct {
	Keys *[]JWKey
}

type GrantType string

const (
	PASSWORD      GrantType = "PASSWORD"
	REFRESH_TOKEN GrantType = "REFRESH-TOKEN"
)

type ScopeType string

const (
	OPENID ScopeType = "OPENID"
)

type SignUpReq struct {
	Scope    ScopeType
	Username string
	Password string
}

type AuthReq struct {
	Scope        ScopeType
	GrantType    GrantType
	Username     string
	Password     string
	RefreshToken string
}

type AuthResp struct {
	UserUid      string
	RefreshToken string
	AccessToken  string
	ExpiresIn    int64
	Scope        ScopeType
}

type ErrorResponse struct {
	StatusCode int `json:"-"`
	Message    string
}
