package main

import (
	"bauman-poker/config"
	externalServices "bauman-poker/external-services"
	handler "bauman-poker/handlers"
	"bauman-poker/repo"
	"bauman-poker/usecases"
	"bauman-poker/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbConfig := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dbConfig), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("can't connect to dbms")
	}
	err = db.AutoMigrate(repo.PlayerAccount{})
	if err != nil {
		log.WithError(err).Fatal("can't migrate db")
		return
	}

	router := gin.Default()
	apiV1 := router.Group(config.GroupName)
	router.Use(utils.JSONLogMiddleware)

	repo := repo.NewIdentityProvRepo(db)
	identityProv := externalServices.NewIdentityExterService()

	uc := usecases.NewGameServerUsecase(repo, identityProv)
	validator := utils.NewTokenValidator(identityProv)
	handler.NewGameServerHandler(uc, validator).RegisterRoutes(apiV1)

	router.GET("/manage/health", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	router.Run(fmt.Sprintf(":%d", config.HostPort))
}
