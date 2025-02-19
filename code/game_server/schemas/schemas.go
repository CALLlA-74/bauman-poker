package schemas

import (
	repo "bauman-poker/repo"
)

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

type ScopeType string

const (
	OPENID ScopeType = "OPENID"
)

type SignUpReq struct {
	Scope    ScopeType `validate:"required"`
	Username string    `validate:"required"`
	Password string    `validate:"required"`
}

type AuthResp struct {
	UserUid      string    `validate:"required"`
	RefreshToken string    `validate:"required"`
	AccessToken  string    `validate:"required"`
	ExpiresIn    int64     `validate:"required"`
	Scope        ScopeType `validate:"required"`
}

type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Message    string `validate:"required"`
}

type UserInfo struct {
	UserUid    string             `validate:"required"`
	Username   string             `validate:"required"`
	NumOfGames int64              `validate:"required"`
	NumOfWins  int64              `validate:"required"`
	UserRank   repo.RankType      `validate:"required"`
	UserState  repo.UserStateType `validate:"required"`
	RoomUid    string             `validate:"required"`
}
