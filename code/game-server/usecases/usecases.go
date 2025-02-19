package usecases

import (
	externalServices "bauman-poker/external-services"
	pokergame "bauman-poker/poker-game"
	repo "bauman-poker/repo"

	"bauman-poker/schemas"
	"net/http"
)

type GameServerUsecase struct {
	repo             *repo.GormPlayerRepo
	identityProvider *externalServices.IdentityExterService
	pokerGame        *pokergame.GameBalancer
}

func NewGameServerUsecase(repo *repo.GormPlayerRepo, identityProv *externalServices.IdentityExterService) *GameServerUsecase {
	return &GameServerUsecase{
		repo:             repo,
		identityProvider: identityProv,
		pokerGame:        pokergame.NewGameBalancer(repo),
	}
}

func (uc *GameServerUsecase) SignUpPlayer(req *http.Request, username string) (*schemas.AuthResp, *schemas.ErrorResponse) {
	resp, errResp := uc.identityProvider.RegisterUser(req)
	if errResp != nil {
		return nil, errResp
	}

	if _, err := uc.repo.CreatePlayer(resp.UserUid, username); err != nil {
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
	}
	return resp, errResp
}

func (uc *GameServerUsecase) AuthUser(req *http.Request) (*schemas.AuthResp, *schemas.ErrorResponse) {
	return uc.identityProvider.AuthUser(req)
}

func (uc *GameServerUsecase) Logout(req *http.Request) *schemas.ErrorResponse {
	return uc.identityProvider.Logout(req)
}

func (uc *GameServerUsecase) GetMe(userUid string) (*schemas.UserInfo, *schemas.ErrorResponse) {
	player, err := uc.repo.GetPlayerByUid(userUid)
	if err != nil {
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
	}
	roomUid := uc.pokerGame.GetRoomUidByPlayer(player.Uid)
	userState := repo.IN_GAME
	if roomUid == "" {
		userState = repo.MENU
	}

	return &schemas.UserInfo{
		UserUid:    userUid,
		Username:   player.Username,
		NumOfGames: player.NumOfGames,
		NumOfWins:  player.NumOfWins,
		UserRank:   player.UserRank,
		UserState:  userState,
		RoomUid:    roomUid,
	}, nil
}

func (uc *GameServerUsecase) MatchingRoom(userUid string) (*pokergame.RoomInfo, *schemas.ErrorResponse) {
	res := uc.pokerGame.MatchingRoom(userUid)
	if res == nil {
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
	}
	return res, nil
}

func (uc *GameServerUsecase) GetRoomInfo(roomUid, userUid string) (*pokergame.RoomInfo, *schemas.ErrorResponse) {
	res := uc.pokerGame.GetRoom(roomUid, userUid)
	if res == nil {
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
	}
	return res, nil
}

func (uc *GameServerUsecase) ConnectToRoom(wsReq pokergame.WSRequest) *schemas.ErrorResponse {
	if uc.pokerGame.ConnectToRoom(wsReq) {
		return nil
	}
	return &schemas.ErrorResponse{
		StatusCode: 500,
		Message:    "Internal Server Error",
	}
}

func (uc *GameServerUsecase) GetPlayerInfo(playerUid, pUidRelateOf string) (*pokergame.PlayerInfo, *schemas.ErrorResponse) {
	res, err := uc.pokerGame.GetPlayerInfo(playerUid, pUidRelateOf)
	if err != nil {
		if err.Error() == "404" {
			return nil, &schemas.ErrorResponse{
				StatusCode: 404,
				Message:    "Not Found",
			}
		} else if err.Error() == "505" {
			return nil, &schemas.ErrorResponse{
				StatusCode: 500,
				Message:    "Internal Server Error",
			}
		}
	}
	return res, nil
}
