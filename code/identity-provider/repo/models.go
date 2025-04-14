package repo

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Uid          string `gorm:"column:uid;type:varchar(100);not null;unique"`
	Username     string `gorm:"column:username;type:varchar(100);not null;unique"`
	PasswordHash string `gorm:"column:password_hash;type:varchar;not null"`
}

type JWK struct {
	gorm.Model

	ID     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Kid    string `gorm:"column:kid;type:varchar(100);not null;unique"`
	Kty    string `gorm:"column:kty;type:varchar(10);not null;default:'RSA'"`
	Use    string `gorm:"column:use;type:varchar(10);not null;default:'sig'"`
	Keystr string `gorm:"column:keystr;type:varchar;not null"`
	Alg    string `gorm:"column:alg;type:varchar(10);not null;default:'RS256'"`
	Exp    int64  `gorm:"column:exp;type:integer;not null"`
}

type IssuedJWTToken struct {
	gorm.Model

	Jti       string `gorm:"column:jti;type:varchar(36);primaryKey"`
	Subject   string `gorm:"column:subject;type:varchar(100);not null;foreignKey:Uid"`
	DeviceId  string `gorm:"column:device_id;type:varchar(36);not null"`
	Revoked   bool   `gorm:"column:revoked;type:boolean;default:False"`
	ExpiredIn int64  `gorm:"column:expired_in;type:integer;not null"`
}
