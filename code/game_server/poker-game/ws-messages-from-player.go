package pokergame

type PongMessageType string

const PONG PongMessageType = "PONG"

type PongMessage struct {
	MessageType PongMessageType `validate:"required"`
	//MessageId   int64           `validate:"required"`
}

type AuthMessageType string

const AUTH AuthMessageType = "AUTH"

type AuthMessage struct {
	MessageType AuthMessageType `validate:"required"`
	MessageId   int64           `validate:"required"`
	RoomUid     string          `validate:"required"`
	Token       string          `validate:"required"`
	LastEventId int64           `validate:"required"`
}

type ActionMsgType string

const (
	GAME_ACTION ActionMsgType = "GAME-ACTION"
	VOTE        ActionMsgType = "VOTE"
)

type PlayerActionType BVariantType

const (
	INCOME  PlayerActionType = "INCOME"
	OUTCOME PlayerActionType = "OUTCOME"
)

type CoefType string

const (
	X1_5  CoefType = "X1_5"
	X2    CoefType = "X2"
	ALLIN CoefType = "ALL-IN"
)

type PlayerVoteType string

const (
	START PlayerVoteType = "START"
	WAIT  PlayerVoteType = "WAIT"
)

type ActionMessage struct {
	MessageType ActionMsgType `validate:"required"`
	MessageId   int64         `validate:"required"`
	RoomUid     string        `validate:"required"`
	UserUid     string        `validate:"required"`
	ActionType  PlayerActionType
	Coef        CoefType
	VoteType    PlayerVoteType
}
