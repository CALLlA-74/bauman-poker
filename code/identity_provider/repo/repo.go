package repo

import (
	"bauman-poker/config"
	"bytes"
	"errors"
	"time"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"

	"gorm.io/gorm"
)

type GormIdentityProvRepo struct {
	db *gorm.DB
}

func NewIdentityProvRepo(db *gorm.DB) *GormIdentityProvRepo {
	g := GormIdentityProvRepo{db: db}
	g.checkJWKs()
	return &g
}

func (g GormIdentityProvRepo) checkJWKs() {
	genJWK := func() *JWK {
		key, err := rsa.GenerateKey(rand.Reader, config.PriviteKeyLen)
		if err != nil {
			log.Error("Error in generating repo.checkJWKs.genJWK: ", err)
			return nil
		}

		bufExp := new(bytes.Buffer)
		err2 := binary.Write(bufExp, binary.LittleEndian, int32(key.E))
		if err2 != nil {
			log.Errorf("Error in transform key.Exp to []bytes. repo.checkJWKs.genJWK: %s", err2)
			return nil
		}

		return &JWK{
			Kid:    uuid.NewString(),
			Kty:    config.KeyType,
			Keystr: privateKeyToString(key),
			Alg:    config.SigningAlg,
			Exp:    time.Now().Unix() + config.CryptoKeyExpPeriodDays*24*3600,
		}
	}

	var keys []JWK
	if err := g.db.Find(&keys).Error; err != nil {
		log.Error("Error in repo.checkJWKs: ", err)
		return
	}
	if len(keys) > 0 {
		if time.Now().Unix() < keys[0].Exp {
			return
		}
		g.db.Delete(&JWK{})
	}

	for idx := 0; idx < config.NumOfKeys; idx++ {
		key := genJWK()
		if key == nil {
			log.Error("Error in keygen. repo.checkJWKs")
			idx--
			continue
		}
		//log.Infof("%d) kid: %s; key: %s", idx+1, key.Kid, key.Keystr)

		if err := g.db.Save(key).Error; err != nil {
			log.Errorf("Error in adding key to DB. repo.checkJWKs: %s", err)
			idx--
			continue
		}
	}
}

func (g GormIdentityProvRepo) GetJWKs() *[]JWKey {
	var keys []JWK
	if err := g.db.Find(&keys).Error; err != nil {
		log.Error("Error in repo.GetJWKs: ", err)
		return nil
	}

	res := make([]JWKey, len(keys))
	for idx, key := range keys {
		k, err := g.jwkToPubKey(&key)
		if err != nil {
			log.Errorf("Error in getting PubKey .repo.GetJWKs")
			return &res
		}
		res[idx] = JWKey{
			Kty: key.Kty,
			Use: key.Use,
			N:   base64.URLEncoding.EncodeToString(k.N.Bytes()),
			E:   base64.URLEncoding.EncodeToString(intToByte(k.E)),
			Kid: key.Kid,
			Alg: key.Alg,
		}
	}
	return &res
}

func (g GormIdentityProvRepo) GetPrivKeyByKid(kid string) (*rsa.PrivateKey, error) {
	var key JWK
	if err := g.db.Where("kid = ?", kid).First(&key).Error; err != nil {
		log.Errorf("Error (kid = %s) in repo.GetPrivKeyByKid: %s", kid, err)
		return nil, fmt.Errorf("500")
	}

	res, err := loadPrivateKeyFromString(key.Keystr)
	if err != nil {
		log.Errorf("Error (kid = %s) in repo.GetPrivKeyByKid", kid)
		return nil, err
	}
	return res, nil
}

func (g GormIdentityProvRepo) GetPubKeyByKid(kid string) (*rsa.PublicKey, error) {
	var key JWK
	if err := g.db.Where("kid = ?", kid).First(&key).Error; err != nil {
		log.Errorf("Error (kid = %s) in repo.GetPubKeyByKid: %s", kid, err)
		return nil, fmt.Errorf("500")
	}
	return g.jwkToPubKey(&key)
}

func (g GormIdentityProvRepo) jwkToPubKey(key *JWK) (*rsa.PublicKey, error) {
	k, err := loadPrivateKeyFromString(key.Keystr)
	if err != nil {
		log.Errorf("Error (kid = %s) in repo.jwkToPubKey; key: %s", key.Kid, key.Keystr)
		return nil, err
	}
	return &k.PublicKey, nil
}

func (g GormIdentityProvRepo) CreateUser(username, password string) (*User, error) {
	if check, err := g.containsUsername(username); err != nil || check {
		if err != nil {
			log.Errorf("Error in (checking) repo.CreateUser: %s", err)
			return nil, fmt.Errorf("500")
		}
		log.Info("repo.CreateUser; username is not unique")
		return nil, fmt.Errorf("400")
	}

	model := User{
		Uid:          uuid.NewString(),
		Username:     username,
		PasswordHash: g.genPasswordHash(password, nil),
	}
	err := g.db.Save(&model).Error

	if err != nil {
		log.Error(fmt.Sprintf("Error in repo.CreateUser: %s", err))
		return nil, fmt.Errorf("500")
	}

	return &model, nil
}

func (g GormIdentityProvRepo) containsUsername(name string) (bool, error) {
	var res int64 = 0
	if err := g.db.Model(&User{}).Where("username = ?", name).Count(&res).Error; err != nil {
		return false, err
	}
	return res > 0, nil
}

func (g GormIdentityProvRepo) GetUserByUid(userUid string) *User {
	model := User{}
	if err := g.db.Where("uid = ?", userUid).First(&model).Error; err != nil {
		log.Info(fmt.Sprintf("Error in repo.GetUserByUid: %s", err))
		return nil
	}
	return &model
}

func (g GormIdentityProvRepo) GetUserByNameAndPassw(username, password string) (*User, error) {
	model := User{}
	if err := g.db.Where("username = ?", username).First(&model).Error; err != nil {
		log.Info(fmt.Sprintf("Error in repo.CheckUserExistByNameAndPassw: %s", err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("401")
		}
		return nil, fmt.Errorf("500")
	}
	if !g.verifyPassword(password, model.PasswordHash) {
		return nil, fmt.Errorf("401")
	}
	return &model, nil
}

func (g GormIdentityProvRepo) genPasswordHash(password string, salt *[]byte) string {
	var genSalt = func() *[]byte {
		salt := make([]byte, config.SaltSize)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			log.Error(err)
		}
		return &salt
	}

	mSalt := salt
	if salt == nil {
		mSalt = genSalt()
	}

	passwordHash := hex.EncodeToString(pbkdf2.Key([]byte(password), *mSalt, config.NumOfHashIter, config.KeySize, sha256.New))
	return passwordHash + "." + hex.EncodeToString(*mSalt)
}

func (g GormIdentityProvRepo) verifyPassword(password, passwordHash string) bool {
	salt, err := hex.DecodeString(strings.Split(passwordHash, ".")[1])
	if err != nil {
		log.Error(err)
	}

	tmp := g.genPasswordHash(password, &salt)
	log.Debug("Password hash legacy: ", passwordHash)
	log.Debug("Password hash new: ", tmp)
	log.Debug("legacy == new? ", tmp)
	return passwordHash == g.genPasswordHash(password, &salt)
}

func privateKeyToString(privkey *rsa.PrivateKey) string {
	privkeyBytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkeyBytes,
		},
	)
	return string(privkeyPem)
}

func loadPrivateKeyFromString(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		err := fmt.Errorf("empty decode result")
		log.Errorf("Error in decoding key. repo.loadPrivateKeyFromString: %s", err)
		return nil, fmt.Errorf("500")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("Error in parse key. repo.loadPrivateKeyFromString: %s", err)
		return nil, fmt.Errorf("500")
	}

	return priv, nil
}

func intToByte(v int) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint32(v)); err != nil {
		log.Errorf("Error in repo.IntToByte: %s", err)
	}

	res := buf.Bytes()
	return res
}

func (g GormIdentityProvRepo) AddIssuedJWT(jti string, userUid string, deviceId string, expIn int64) error {
	model := IssuedJWTToken{
		Jti:       jti,
		Subject:   userUid,
		DeviceId:  deviceId,
		ExpiredIn: expIn,
	}

	if err := g.db.Save(&model).Error; err != nil {
		log.Errorf("Error in repo.AddIssuedJWT: %s", err)
		return fmt.Errorf("500")
	}
	return nil
}

func (g GormIdentityProvRepo) GetIssuedJWTByJti(jti string) (*IssuedJWTToken, error) {
	model := IssuedJWTToken{}
	if err := g.db.Where("jti = ?", jti).First(&model).Error; err != nil {
		log.Errorf("Error in getting. repo.CheckJWTRevokeByJti: %s", err)
		return nil, fmt.Errorf("500")
	}
	return &model, nil
}

func (g GormIdentityProvRepo) RevokeJWT(jti, userUid, deviceId string) error {
	if jti == "" {
		if err := g.db.Model(&IssuedJWTToken{}).Where("subject = ?", userUid).
			Where("device_id = ?", deviceId).Updates(IssuedJWTToken{Revoked: true}).Error; err != nil {

			log.Errorf("Error in updating (by uid, deviceId). repo.RevokeJWT: %s", err)
			return fmt.Errorf("500")
		}
	} else {
		model := IssuedJWTToken{}
		if err := g.db.Where("jti = ?", jti).First(&model).Error; err != nil {
			log.Errorf("Error in getting. repo.RevokeJWTByJti: %s", err)
			return fmt.Errorf("500")
		}

		model.Revoked = true
		if err := g.db.Where("jti = ?", jti).Updates(model).Error; err != nil {
			log.Errorf("Error in updating. repo.RevokeJWTByJti: %s", err)
			return fmt.Errorf("500")
		}
	}

	return nil
}
