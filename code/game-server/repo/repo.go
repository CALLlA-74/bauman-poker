package repo

import (
	"errors"

	"fmt"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

type GormPlayerRepo struct {
	db *gorm.DB
}

func NewIdentityProvRepo(db *gorm.DB) *GormPlayerRepo {
	g := GormPlayerRepo{db: db}
	return &g
}

func (g *GormPlayerRepo) CreatePlayer(uid, username string) (*PlayerAccount, error) {
	model := &PlayerAccount{
		Uid:      uid,
		Username: username,
	}
	if err := g.db.Save(model).Error; err != nil {
		log.WithError(err).Errorf("Error in saving model. repo.CreatePlayer; uid: %s; username: %s", uid, username)
		return nil, fmt.Errorf("500")
	}
	return model, nil
}

func (g *GormPlayerRepo) GetPlayerByUid(userUid string) (*PlayerAccount, error) {
	model := PlayerAccount{}
	if err := g.db.Where(PlayerAccount{Uid: userUid}).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("User (uid: %s) not found", userUid)
			return nil, fmt.Errorf("404")
		} else {
			log.WithError(err).Errorf("Error in repo.GetPlayerByUid; (uid: %s)", userUid)
			return nil, fmt.Errorf("500")
		}
	}
	return &model, nil
}
