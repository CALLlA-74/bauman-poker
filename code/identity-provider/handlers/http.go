package handlers

import (
	"bauman-poker/repo"
	"bauman-poker/usecases"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	HeaderAuth = "Authorization"
)

type IdentityProvHandler struct {
	uc *usecases.IdentityProvUsecase
}

func NewIdentityProvHandler(uc *usecases.IdentityProvUsecase) *IdentityProvHandler {
	return &IdentityProvHandler{uc: uc}
}

func (lp IdentityProvHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/.well-known/jwks.json", lp.getJWKs)
	router.POST("/register", lp.registerUser)
	router.POST("/oauth/token", lp.authUser)
	router.DELETE("/oauth/revoke", lp.logout)
}

func (lp IdentityProvHandler) getJWKs(context *gin.Context) {
	keys := lp.uc.GetJWKs()
	context.JSON(http.StatusOK, keys)
}

func (lp IdentityProvHandler) registerUser(context *gin.Context) {
	req := repo.SignUpReq{}
	if err := context.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("register Error in ShouldBindJSON")
		context.JSON(http.StatusBadRequest, repo.ErrorResponse{
			Message: "Bad Request",
		})
		return
	}

	resp, errResp := lp.uc.SignUpUser(&req)

	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		log.Errorf("ErrorResp: %v", errResp)
		return
	}

	log.Infof("RespBody: %v", resp)
	context.JSON(http.StatusOK, resp)
}

func (lp IdentityProvHandler) authUser(context *gin.Context) {
	req := repo.AuthReq{}
	if err := context.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("auth Error in ShouldBindJSON")
		context.JSON(http.StatusUnauthorized, repo.ErrorResponse{
			Message: "Unauthorized",
		})
		return
	}

	resp, errResp := lp.uc.AuthUser(&req)

	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}

	context.JSON(http.StatusOK, resp)
}

func (lp IdentityProvHandler) logout(context *gin.Context) {
	token := context.GetHeader(HeaderAuth)
	token = strings.Replace(token, "Bearer ", "", -1)
	errResp := lp.uc.Logout(token)
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	context.JSON(http.StatusNoContent, nil)
}
