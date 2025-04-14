package utils

import (
	"bauman-poker/config"
	externalServices "bauman-poker/external-services"
	"bauman-poker/schemas"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type TokenValidator struct {
	jwks             []schemas.JWKey
	identityProvider *externalServices.IdentityExterService
}

func NewTokenValidator(identityProv *externalServices.IdentityExterService) *TokenValidator {
	return &TokenValidator{
		jwks:             []schemas.JWKey{},
		identityProvider: identityProv,
	}
}

func (tv *TokenValidator) getJWKs() {
	if len(tv.jwks) <= 0 {
		tv.jwks = tv.identityProvider.GetJWKs()
		log.Infof("List len: %d", len(tv.jwks))
		log.Infof("key[0]: %v", tv.jwks[0])
	}
}

func (tv *TokenValidator) getJWK(kid string) *schemas.JWKey {
	tv.getJWKs()
	key := tv.getKeyById(kid)

	if key == nil {
		log.Infof("Key is nil")
		tv.jwks = []schemas.JWKey{}
		tv.getJWKs()
		key := tv.getKeyById(kid)
		if key == nil {
			return nil
		}
	}
	return key
}

func (tv *TokenValidator) getKeyById(kid string) *schemas.JWKey {
	for _, key := range tv.jwks {
		log.Infof("Kid: %s", key.Kid)
		if key.Kid == kid {
			return &key
		}
	}
	return nil
}

func (tv *TokenValidator) VerifyAccessToken(token string) bool {
	header, payload, _, err := tv.ParseAccessToken(token)
	if err != nil {
		log.Errorf("Error in repo.verifyAccessToken")
		return false
	}
	key := tv.getJWK(header.(*Header).Kid)
	if key == nil {
		log.Error("Error in getting PubKey. repo.verifyAccessToken")
		return false
	}

	pubKey := jwkeyToPubKey(key)
	if pubKey == nil {
		return false
	}

	f1 := tv.verifyAlg(header.(*Header))
	f2 := tv.verifyTyp(header.(*Header))
	f3 := tv.verifyTokenType(header.(*Header))
	f4 := tv.verifyIss(payload.(*AccessTokenPayload))
	f5 := tv.verifyExp(payload.(*AccessTokenPayload))
	f6 := tv.verifySignature(token, pubKey, header.(*Header).Alg)
	log.Infof("verification flags: %t, %t, %t, %t, %t, %t", f1, f2, f3, f4, f5, f6)

	if !f1 || !f2 || !f3 || !f4 || !f5 || !f6 {
		return false
	}

	return true
}

func jwkeyToPubKey(key *schemas.JWKey) *rsa.PublicKey {
	e_bytes, err := base64.URLEncoding.DecodeString(key.E)
	if err != nil {
		log.WithError(err).Errorf("Error in casting utils.JWKeyToPubKey. (E)")
		return nil
	}
	e := binary.LittleEndian.Uint32(e_bytes)

	n_bytes, err2 := base64.URLEncoding.DecodeString(key.N)
	if err2 != nil {
		log.WithError(err2).Errorf("Error in casting utils.JWKeyToPubKey. (N)")
		return nil
	}
	n := new(big.Int)
	n.SetBytes(n_bytes)

	res := &rsa.PublicKey{}
	res.E = int(e)
	res.N = n

	return res
}

func (tv *TokenValidator) verifySignature(token string, key *rsa.PublicKey, alg string) bool {
	v := strings.Split(token, ".")
	method := jwt.GetSigningMethod(alg)
	if err := method.Verify(v[0]+"."+v[1], v[2], key); err != nil {
		log.WithError(err).Errorf("Verify Sign error. h: %s; p: %s", v[0], v[1])
		return false
	}
	return true
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

func (tv *TokenValidator) ParseAccessToken(token string) (any, any, string, error) {
	v := strings.Split(token, ".")

	header := &Header{}
	if err := jwtB64Decode(v[0], header); err != nil {
		log.Errorf("Error in parsing header.token-utils.parseToken. Input: %s", v[0])
		return nil, nil, "", fmt.Errorf("500")
	}

	payload := &AccessTokenPayload{}
	if err := jwtB64Decode(v[1], payload); err != nil {
		log.Errorf("Error in parsing payload.token-utils.parseToken. Input: %s", v[1])
		return nil, nil, "", fmt.Errorf("500")
	}

	return header, payload, v[2], nil
}

func (tv *TokenValidator) verifyAlg(header *Header) bool {
	return header.Alg == config.SigningAlg
}

func (tv *TokenValidator) verifyTyp(header *Header) bool {
	return header.Typ == config.TokenStandard
}

func (tv *TokenValidator) verifyTokenType(header *Header) bool {
	return header.TokenType == ACCESS
}

func (tv *TokenValidator) verifyIss(payload *AccessTokenPayload) bool {
	return payload.Iss == config.IdentityExterBaseUrl+"/"
}

func (tv *TokenValidator) verifyExp(payload *AccessTokenPayload) bool {
	now := time.Now().Unix()
	var leeway int64 = config.LeewaySeconds

	return payload.Exp+leeway >= now
}
