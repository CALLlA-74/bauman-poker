package handlers

import (
	pokergame "bauman-poker/poker-game"
	"bauman-poker/schemas"
	"bauman-poker/usecases"
	"bauman-poker/utils"
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	HeaderAuth = "Authorization"
)

type GameServerHandler struct {
	uc        *usecases.GameServerUsecase
	validator *utils.TokenValidator
}

func NewGameServerHandler(uc *usecases.GameServerUsecase, v *utils.TokenValidator) *GameServerHandler {
	return &GameServerHandler{
		uc:        uc,
		validator: v,
	}
}

func (lp GameServerHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/register", lp.registerUser)
	router.POST("/oauth/token", lp.authUser)
	router.DELETE("/oauth/revoke", lp.logout)
	router.GET("/me", lp.validateToken, lp.getMe)
	router.GET("/rooms/matching", lp.validateToken, lp.matchingRoom)
	router.GET("/rooms/:roomUid", lp.validateToken, lp.getRoomInfo)
	router.GET("/players/:userUid", lp.validateToken, lp.getPlayerInfo)
	router.GET("/rooms-ws/:roomUid", lp.connectToRoom)
}

func (lp GameServerHandler) registerUser(context *gin.Context) {
	reqBytes, err := io.ReadAll(context.Request.Body)
	context.Request.Body.Close()
	if err != nil {
		log.WithError(err).Error("register Error in body readall")
		context.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Message: "Bad Request",
		})
		return
	}

	req := schemas.SignUpReq{}
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		log.WithError(err).Error("register Error in Unmarshalling json")
		context.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Message: "Bad Request",
		})
		return
	}

	context.Request.Body = io.NopCloser(bytes.NewBuffer(reqBytes))
	resp, errResp := lp.uc.SignUpPlayer(context.Request, req.Username)

	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}

	context.JSON(http.StatusOK, resp)
}

func (lp GameServerHandler) authUser(context *gin.Context) {
	resp, errResp := lp.uc.AuthUser(context.Request)

	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}

	context.JSON(http.StatusOK, resp)
}

func (lp GameServerHandler) logout(context *gin.Context) {
	errResp := lp.uc.Logout(context.Request)
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	context.JSON(http.StatusNoContent, nil)
}

func (lp GameServerHandler) validateToken(context *gin.Context) {
	token := strings.Replace(context.GetHeader(HeaderAuth), "Bearer ", "", -1)
	if !lp.validator.VerifyAccessToken(token) {
		context.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Message: "Unauthorized",
		})
		context.Abort()
		return
	}

	_, payload, _, _ := lp.validator.ParseAccessToken(token)
	context.Set("UserUid", payload.(*utils.AccessTokenPayload).UserUid)
	context.Next()
}

func (lp GameServerHandler) getMe(context *gin.Context) {
	userUid, _ := context.Get("UserUid")

	succResp, errResp := lp.uc.GetMe(userUid.(string))
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	context.JSON(http.StatusOK, succResp)
}

func (lp GameServerHandler) matchingRoom(context *gin.Context) {
	userUid, _ := context.Get("UserUid")

	succResp, errResp := lp.uc.MatchingRoom(userUid.(string))
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	context.JSON(http.StatusOK, succResp)
}

func (lp GameServerHandler) getRoomInfo(context *gin.Context) {
	userUid, _ := context.Get("UserUid")
	roomUid := context.Param("roomUid")

	succResp, errResp := lp.uc.GetRoomInfo(roomUid, userUid.(string))
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	context.JSON(http.StatusOK, succResp)
}

func (lp GameServerHandler) connectToRoom(context *gin.Context) {
	roomUid := context.Param("roomUid")
	userUid := context.Request.URL.Query().Get("uid")
	req := pokergame.WSRequest{
		RW:             context.Writer,
		Req:            context.Request,
		RoomUid:        roomUid,
		PlayerUid:      userUid,
		TokenValidator: lp.validator,
	}

	errResp := lp.uc.ConnectToRoom(req)
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	//context.Abort()
	//context.AbortWithStatus(http.StatusSwitchingProtocols)
}

func (lp GameServerHandler) getPlayerInfo(context *gin.Context) {
	playerUid := context.Param("userUid")
	pUidRelateOf, _ := context.Get("UserUid")
	succResp, errResp := lp.uc.GetPlayerInfo(playerUid, pUidRelateOf.(string))
	if errResp != nil {
		context.JSON(errResp.StatusCode, errResp)
		return
	}
	context.JSON(http.StatusOK, succResp)
}
