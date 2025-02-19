package pokergame

import (
	"bauman-poker/repo"
	"bauman-poker/utils"
	"net/http"
)

type WSRequest struct {
	RW             http.ResponseWriter
	Req            *http.Request
	RoomUid        string
	PlayerUid      string
	TokenValidator *utils.TokenValidator
}

type LastActionLabelType string

const (
	LBL_NONE  LastActionLabelType = "NONE"
	LBL_FOLD  LastActionLabelType = "FOLD"
	LBL_CHECK LastActionLabelType = "CHECK"
	LBL_CALL  LastActionLabelType = "CALL"
	LBL_RAISE LastActionLabelType = "RAISE"
	LBL_ALLIN LastActionLabelType = "ALL-IN"
)

type PlayerInfo struct {
	UserUid          string `validate:"required"`
	Username         string `validate:"required"`
	ImageUrl         string
	Bet              int64               `validate:"required"`
	Deposit          int64               `validate:"required"`
	LastActionLabel  LastActionLabelType `validate:"required"`
	UserRank         repo.RankType       `validate:"required"`
	PersonalCardList *[]*PlayingCard     `validate:"required"`
	BestCombName     string              `validate:"required"`
	BoutVariants     *[]BoutVariantType  `validate:"required"`
	TimeEndBout      int64               `validate:"required"`
	VoteType         PlayerVoteType      `validate:"required"`
}

type RoomInfo struct { // состояние комнаты с точки зрения одного из игроков
	RoomUid           string          `validate:"required"`
	RoomState         RoomStateType   `validate:"required"`
	PlayerList        *[]*PlayerInfo  `validate:"required"`
	TableCardList     *[]*PlayingCard `validate:"required"`
	Stack             int64           `validate:"required"`
	Bout              string          `validate:"required"`
	DealerUid         string          `validate:"required"`
	LastEventId       int64           `validate:"required"`
	RoundNumber       int64           `validate:"required"`
	NumOfPlayers      int16           `validate:"required"`
	NumOfStartPlayers int16           `validate:"required"`
}
