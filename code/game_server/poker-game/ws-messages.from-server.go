package pokergame

type PingMessageType string

const PING PingMessageType = "PING"

type PingMessage struct {
	MessageType PingMessageType `validate:"required"`
	//MessageId   int64           `validate:"required"`
}

type ResponseMsgType string

const ACK ResponseMsgType = "ACK"

type RespStatusCodeType int16

const (
	StatusOK           RespStatusCodeType = 200
	StatusUnauthorized RespStatusCodeType = 401
	StatusBadReq       RespStatusCodeType = 400
)

type ResponseMessage struct {
	MessageType  ResponseMsgType    `validate:"required"`
	AckMessageId int64              `validate:"required"`
	StatusCode   RespStatusCodeType `validate:"required"`
}

type EventMsgType string

const EVENT EventMsgType = "EVENT"

type BVariantType string

const (
	FOLD  BVariantType = "FOLD"
	CHECK BVariantType = "CHECK"
	CALL  BVariantType = "CALL"
	RAISE BVariantType = "RAISE"
)

type ActType PlayerActionType

const (
	//INCOME ActType = "INCOME"
	BOUT         ActType = "BOUT"
	ALL_IN       ActType = "ALL-IN"
	SET_DEALER   ActType = "SET-DEALER"
	MIN_BLIND_IN ActType = "MIN-BLIND-IN"
	MAX_BLIND_IN ActType = "MAX-BLIND-IN"
)

type BoutVariantType struct {
	VariantType   BVariantType `validate:"required"`
	CallValue     int64
	RaiseVariants *[]CoefType
}

type PlayerActionEvent struct {
	EventId              int64   `validate:"required"`
	UserUid              string  `validate:"required"`
	ActionType           ActType `validate:"required"`
	BoutVariants         *[]BoutVariantType
	BestCombName         string
	TimeEndBoutOrForming int64
	NewBet               int64
	NewDeposit           int64
}

type PrepareEvent struct {
	EventId           int64 `validate:"required"`
	NumOfPlayers      int16 `validate:"required"`
	NumOfStartPlayers int16 `validate:"required"`
}

type GameEventType string

const (
	ROOM_STATE_UPDATE GameEventType = "ROOM_STATE_UPDATE"
	NEW_ROUND         GameEventType = "NEW_ROUND"
	NEW_TRADE_ROUND   GameEventType = "NEW_TRADE_ROUND"
	PERSONAL_CARDS    GameEventType = "PERSONAL_CARDS"
	CARDS_ON_TABLE    GameEventType = "CARDS_ON_TABLE"
	BET_ACCEPTED      GameEventType = "BET_ACCEPTED"
	WINNER_RESULT     GameEventType = "WINNER_RESULT"
)

type GameEvent struct {
	EventId          int64         `validate:"required"`
	EventType        GameEventType `validate:"required"`
	NewRoomState     RoomStateType
	RoundNumber      int16
	PlayingCardsList *[]*PlayingCard
	ClosedCards      *[]*[]*PlayingCard
	PlayerUids       *[]string
	WinnerUids       *[]string
	BestCombinations *[]*[]*PlayingCard
	BestCombName     string
	WinnerDeposits   *[]int64
	NewStack         int64
}

type TypeOfEvent string

const (
	PREPARE_EVENT       TypeOfEvent = "PREPARE-EVENT"
	GAME_EVENT          TypeOfEvent = "GAME-EVENT"
	PLAYER_ACTION_EVENT TypeOfEvent = "PLAYER-ACTION-EVENT"
)

type EventMessage[T PlayerActionEvent | PrepareEvent | GameEvent] struct {
	MessageType     EventMsgType `validate:"required"`
	MessageId       int64        `validate:"required"`
	EventType       TypeOfEvent  `validate:"required"`
	EventDescriptor *T           `validate:"required"`
}
