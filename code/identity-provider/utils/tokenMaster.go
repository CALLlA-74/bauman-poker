package utils

import (
	"bauman-poker/config"
	"bauman-poker/repo"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type TokenMaster struct {
	repo *repo.GormIdentityProvRepo
}

func NewTokenMaster(repo *repo.GormIdentityProvRepo) *TokenMaster {
	return &TokenMaster{
		repo: repo,
	}
}

func (tm TokenMaster) GenerateSignedTokens(user *repo.User, scope repo.ScopeType) (*repo.AuthResp, *repo.ErrorResponse) {
	deviceId := uuid.NewString()
	iatTime := time.Now().Unix()
	expTimeAccess := iatTime + config.AccessTokenExpMinutes*60
	expTimeRefresh := iatTime + config.RefreshTokenExpHours*3600

	accessToken, err1 := tm.generateSignedAccessToken(user, deviceId, iatTime, expTimeAccess)
	refreshToken, err2 := tm.generateSignedRefreshToken(user, deviceId, iatTime, expTimeRefresh)
	if err1 != nil || err2 != nil {
		return nil, &repo.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}
	return &repo.AuthResp{
		UserUid:      user.Uid,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresIn:    expTimeAccess,
		Scope:        scope,
	}, nil
}

func (tm TokenMaster) generateSignedAccessToken(user *repo.User, deviceId string, iat, expIn int64) (string, error) {
	keys := tm.repo.GetJWKs()
	if keys == nil || len(*keys) <= 0 {
		log.Errorf("Key list is empty. token-utils.generateSignedAccessToken")
		return "", fmt.Errorf("500")
	}

	kIdx := rand.Intn(len(*keys) - 1)
	key, err := tm.repo.GetPrivKeyByKid((*keys)[kIdx].Kid)
	if err != nil {
		log.Errorf("Error in getting PKey. token-utils.generateSignedAccessToken")
		return "", err
	}
	iatTime := iat
	expTime := expIn

	header := Header{
		Alg:       (*keys)[kIdx].Alg,
		Typ:       config.TokenStandard,
		Kid:       (*keys)[kIdx].Kid,
		TokenType: ACCESS,
	}
	h, err2 := jwtB64Encode(header)
	if err2 != nil {
		log.Errorf("Error in marshalling header. token-utils.generateSignedAccessToken")
		return "", err2
	}

	payload := AccessTokenPayload{
		Jti:      uuid.NewString(),
		UserUid:  user.Uid,
		Iss:      config.Hostname,
		Iat:      iatTime,
		Exp:      expTime,
		DeviceId: deviceId,
	}
	p, err3 := jwtB64Encode(payload)
	if err3 != nil {
		log.Errorf("Error in marshalling payload. token-utils.generateSignedAccessToken")
		return "", err3
	}

	method := jwt.GetSigningMethod((*keys)[kIdx].Alg)
	sign, err4 := method.Sign(h+"."+p, key)
	if err4 != nil {
		log.Errorf("Error in signing. token-utils.generateSignedAccessToken: %s", err4)
		return "", err4
	}

	log.Debugf("Header: %s; Payload: %s", h, p)
	return h + "." + p + "." + sign, nil
}

func (tm TokenMaster) generateSignedRefreshToken(user *repo.User, deviceId string, iat, expIn int64) (string, error) {
	keys := tm.repo.GetJWKs()
	if keys == nil || len(*keys) <= 0 {
		log.Errorf("Key list is empty. token-utils.generateSignedRefreshToken")
		return "", fmt.Errorf("500")
	}

	kIdx := rand.Intn(len(*keys) - 1)
	key, err := tm.repo.GetPrivKeyByKid((*keys)[kIdx].Kid)
	if err != nil {
		log.Errorf("Error in getting PKey. token-utils.generateSignedRefreshToken")
		return "", err
	}
	iatTime := iat
	expTime := expIn

	header := Header{
		Alg:       (*keys)[kIdx].Alg,
		Typ:       config.TokenStandard,
		Kid:       (*keys)[kIdx].Kid,
		TokenType: REFRESH,
	}
	h, err2 := jwtB64Encode(header)
	if err2 != nil {
		log.Errorf("Error in marshalling header. token-utils.generateSignedRefreshToken")
		return "", err2
	}

	payload := RefreshTokenPayload{
		Jti:      uuid.NewString(),
		UserUid:  user.Uid,
		Iss:      config.Hostname,
		Iat:      iatTime,
		Exp:      expTime,
		DeviceId: deviceId,
	}
	p, err3 := jwtB64Encode(payload)
	if err3 != nil {
		log.Errorf("Error in marshalling payload. token-utils.generateSignedRefreshToken")
		return "", err3
	}

	method := jwt.GetSigningMethod((*keys)[kIdx].Alg)
	sign, err4 := method.Sign(h+"."+p, key)
	if err4 != nil {
		log.Errorf("Error in signing. token-utils.generateSignedRefreshToken: %s", err4)
		return "", err4
	}

	if err := tm.repo.AddIssuedJWT(payload.Jti, user.Uid, deviceId, payload.Exp); err != nil {
		log.Errorf("Error in adding ref-token to DB. token-utils.generateSignedRefreshToken: %s", err)
		return "", err
	}

	log.Debugf("Header: %s; Payload: %s", h, p)
	return h + "." + p + "." + sign, nil
}

func (tm TokenMaster) UpdateTokens(oldRefrToken string, scope repo.ScopeType) (*repo.AuthResp, *repo.ErrorResponse) {
	if !tm.verifyRefreshToken(oldRefrToken) {
		log.Infof("Verification failed. token: %s; repo.UpdateTokens.", oldRefrToken)
		return nil, &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	_, payload, _, err := tm.parseRefreshToken(oldRefrToken)
	if err != nil {
		log.Errorf("Error in parsing token. repo.UpdateTokens.")
		return nil, &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	p := payload.(*RefreshTokenPayload)
	if tm.repo.RevokeJWT(p.Jti, "", "") != nil {
		log.Errorf("Error in revoking old token. repo.UpdateTokens.")
		return nil, &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	user := tm.repo.GetUserByUid(p.UserUid)
	if user == nil {
		log.Errorf("Error in getting user. repo.UpdateTokens.")
		return nil, &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	iat := time.Now().Unix()
	expTimeAccess := iat + config.AccessTokenExpMinutes*60
	expTimeRefresh := iat + config.RefreshTokenExpHours*3600

	newAccessToken, err2 := tm.generateSignedAccessToken(user, p.DeviceId, iat, expTimeAccess)
	newRefreshToken, err3 := tm.generateSignedRefreshToken(user, p.DeviceId, iat, expTimeRefresh)
	if err2 != nil || err3 != nil {
		log.Errorf("Error in gen new tokens. repo.UpdateTokens.")
		return nil, &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	return &repo.AuthResp{
		UserUid:      user.Uid,
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
		ExpiresIn:    expTimeAccess,
		Scope:        scope,
	}, nil
}

func (tm TokenMaster) RevokeToken(token string) *repo.ErrorResponse {
	if !tm.verifyRefreshToken(token) {
		log.Infof("Verification failed. token: %s; repo.RevokeToken", token)
		return &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	_, payload, _, err := tm.parseRefreshToken(token)
	if err != nil {
		log.Errorf("Error in parsing token. repo.RevokeToken.")
		return &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	p := payload.(*RefreshTokenPayload)
	if tm.repo.RevokeJWT(p.Jti, "", "") != nil {
		log.Errorf("Error in revoking old token. repo.RevokeToken.")
		return &repo.ErrorResponse{
			StatusCode: 401,
			Message:    "Unauthorized",
		}
	}

	return &repo.ErrorResponse{
		StatusCode: 204,
	}
}

func jwtB64Encode(val any) (string, error) {
	res, err := json.Marshal(val)
	if err != nil {
		log.Errorf("Error in token-utils.jwtB64Encode: %s", err)
		return "", fmt.Errorf("500")
	}
	return base64.URLEncoding.EncodeToString(res), nil
}

func jwtB64Decode(inp string, val any) error {
	inpBytes, err := base64.URLEncoding.DecodeString(inp)
	if err != nil {
		log.WithError(err).Errorf("Error in (in UrlEncoding) token-utils.jwtB64Encode.")
		return fmt.Errorf("500")
	}
	if err := json.Unmarshal(inpBytes, val); err != nil {
		log.WithError(err).Errorf("Error in (in unmarshalling) token-utils.jwtB64Encode.")
		return fmt.Errorf("500")
	}
	return nil
}

func (tm TokenMaster) verifyRefreshToken(token string) bool {
	header, payload, _, err := tm.parseRefreshToken(token)
	if err != nil {
		log.Errorf("Error in repo.verifyRefreshToken")
		return false
	}
	key, err2 := tm.repo.GetPubKeyByKid(header.(*Header).Kid)
	if err2 != nil {
		log.Error("Error in getting PubKey. repo.verifyRefreshToken")
		return false
	}

	f1 := tm.verifyAlg(header.(*Header))
	f2 := tm.verifyTyp(header.(*Header))
	f3 := tm.verifyTokenType(header.(*Header))
	f4 := tm.verifyIss(payload.(*RefreshTokenPayload))
	f5 := tm.verifyExp(payload.(*RefreshTokenPayload))
	f6 := tm.verifySignature(token, key, header.(*Header).Alg)
	log.Infof("verification flags: %t, %t, %t, %t, %t, %t", f1, f2, f3, f4, f5, f6)

	if !f1 || !f2 || !f3 || !f4 || !f5 || !f6 {
		return false
	}

	issuedToken, err3 := tm.repo.GetIssuedJWTByJti(payload.(*RefreshTokenPayload).Jti)
	if err3 != nil {
		log.Error("Error in getting issuedToken. repo.verifyRefreshToken")
		return false
	}

	if issuedToken.Revoked {
		go func() {
			for tm.repo.RevokeJWT("", issuedToken.Subject, issuedToken.DeviceId) != nil {
			}
		}()
		log.Infof("try to revoke all tokens for userUid: %s; deviceId: %s", issuedToken.Subject, issuedToken.DeviceId)
		return false
	}

	return true
}

func (tm TokenMaster) parseRefreshToken(token string) (any, any, string, error) {
	v := strings.Split(token, ".")

	header := &Header{}
	if err := jwtB64Decode(v[0], header); err != nil {
		log.Errorf("Error in parsing header.token-utils.parseToken. Input: %s", v[0])
		return nil, nil, "", fmt.Errorf("500")
	}

	payload := &RefreshTokenPayload{}
	if err := jwtB64Decode(v[1], payload); err != nil {
		log.Errorf("Error in parsing payload.token-utils.parseToken. Input: %s", v[1])
		return nil, nil, "", fmt.Errorf("500")
	}

	return header, payload, v[2], nil
}

func (tm TokenMaster) verifySignature(token string, key *rsa.PublicKey, alg string) bool {
	v := strings.Split(token, ".")
	method := jwt.GetSigningMethod(alg)
	if err := method.Verify(v[0]+"."+v[1], v[2], key); err != nil {
		log.WithError(err).Errorf("Verify Sign error. h: %s; p: %s", v[0], v[1])
		return false
	}
	return true
}

func (tm TokenMaster) verifyAlg(header *Header) bool {
	return header.Alg == config.SigningAlg
}

func (tm TokenMaster) verifyTyp(header *Header) bool {
	return header.Typ == config.TokenStandard
}

func (tm TokenMaster) verifyTokenType(header *Header) bool {
	return header.TokenType == REFRESH
}

func (tm TokenMaster) verifyIss(payload *RefreshTokenPayload) bool {
	return payload.Iss == config.Hostname
}

func (tm TokenMaster) verifyExp(payload *RefreshTokenPayload) bool {
	now := time.Now().Unix()
	var leeway int64 = config.LeewaySeconds

	return payload.Exp+leeway >= now
}
