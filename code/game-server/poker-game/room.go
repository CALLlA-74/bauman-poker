package pokergame

import (
	"bauman-poker/config"
	"bauman-poker/repo"
	"container/list"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type RoomStateType string

const (
	FORMING     RoomStateType = "FORMING"
	GAMING      RoomStateType = "GAMING"
	DISSOLUTION RoomStateType = "DISSOLUTION"
)

type GameRoom struct {
	gameBalancer      *GameBalancer
	roomUid           string
	roomState         RoomStateType
	playerMap         map[string]*list.Element //*Player
	playerList        *list.List
	deckOfCards       *[]*PlayingCard
	tableCardList     *[]*PlayingCard
	stack             int64
	boutIter          *list.Element //*Player
	dealerIter        *list.Element //*Player
	lastEventId       int64
	minBlind          int64
	maxBlind          int64
	roundNumber       int16
	tradeRoundNumber  int16
	msgQ              chan *ActionMessage
	numOfPlayers      int16
	numOfStartPlayers int16
	repo              *repo.GormPlayerRepo
	gTimer            *gameTimer
	cardIdx           int16

	ctrlBetSum int64 // контрольная сумма ставок.
	// В нее суммируются ставки игроков внутри круга торгов.
	// При BET_ACCEPTED сумма прибавляется к полю stack, а это поле становится 0

	currentBet     int64
	isIncreasedBet bool
	lastTraderIter *list.Element // указатель на игрока, на котором завершается текущий круг торгов
	//lastRaiserIter *list.Element // указатель на игрока, который последним поднимал ставку
}

func newRoom(repo *repo.GormPlayerRepo, gb *GameBalancer) *GameRoom {
	room := &GameRoom{
		gameBalancer:      gb,
		roomUid:           uuid.NewString(),
		roomState:         FORMING,
		playerMap:         make(map[string]*list.Element), //*Player
		playerList:        list.New(),
		deckOfCards:       &[]*PlayingCard{},
		tableCardList:     &[]*PlayingCard{},
		stack:             0,
		boutIter:          nil,
		dealerIter:        nil,
		lastEventId:       1,
		minBlind:          0, // StartMinBlind
		maxBlind:          0, // StartMinBlind * 2
		roundNumber:       0,
		tradeRoundNumber:  0,
		msgQ:              make(chan *ActionMessage, 1000),
		numOfPlayers:      0,
		numOfStartPlayers: 0,
		repo:              repo,
		gTimer:            newGameTimer(),
		cardIdx:           0,
		isIncreasedBet:    false,
	}

	go func() {
		for room.roomState != DISSOLUTION {
			switch room.roomState {
			case FORMING:
				room.forming()
			case GAMING:
				room.gaming()
			}
		}
	}()

	return room
}

func (gr *GameRoom) forming() {
	check := func() bool {
		return len(gr.playerMap) >= MaxNumOfPlayersPerGame ||
			(len(gr.playerMap) >= MinNumOfPlayersPerGame && (gr.numOfStartPlayers*2 > gr.numOfPlayers || gr.gTimer.isAlarmed))
	}

	for gr.roomState == FORMING {
		if check() {
			gr.setRoomState(GAMING)
			gr.gTimer.reset()
			return
		}
		if len(gr.msgQ) > 0 {
			msg := <-gr.msgQ
			if msg.MessageId == -1 { // запос на фиксацию состояния
				gr.fixRoomStateInfo(msg.UserUid)
				continue
			} else if msg.MessageId == -2 {
				u := strings.Split(msg.UserUid, ".")
				gr.fixPlayerInfo(u[0], u[1])
				continue
			}
			switch msg.MessageType {
			case VOTE:
				gr.voteActionHandling(msg)
			case GAME_ACTION:
				gr.ioPrepareActionHandling(msg)
			}
		}
	}
}

func (gr *GameRoom) gaming() {
	for gr.roundNumber = 1; gr.playerList.Len() > 1; gr.roundNumber++ { // цикл по раундам
		gr.notifyNewRound()
		gr.notifySetDealer()

		for gr.tradeRoundNumber = 1; gr.tradeRoundNumber <= 4 && gr.playerList.Len() > 1; gr.tradeRoundNumber++ { // цикл по торговым кругам
			gr.notifyNewTradeRound()
			if gr.tradeRoundNumber == 1 {
				gr.notifyMinBlindIn()
				gr.notifyMaxBlindIn()
				gr.notifyPersonalCards()
			} else {
				//gr.currentBet = gr.maxBlind
				//gr.lastTraderIter = gr.dealerIter
				gr.notifyTableCards()
				// выдача карт на стол воткрытую
			}
			for gr.playerList.Len() > 1 && !gr.isEndOfTradeRound() { // цикл торгового круга. Завершится, когда последний из игроков, поднимавших ставку, выберет CHECK
				gr.notifyBout()
				p := gr.boutIter.Value.(*Player)
				gr.gTimer.start(time.Duration((p.timeEndBout - time.Now().UnixMilli()) * int64(time.Millisecond)))
				f := false
				for !f { // обрабатываем входящие сообщения, пока не получим действие игрока, у кого ход, или пока не истекло время таймера
					if gr.playerList.Len() <= 1 {
						break
					}
					select {
					case msg, ok := <-gr.msgQ:
						if ok {
							f = gr.playerActionHandling(msg)
							continue
						}
					default:
					}
					if gr.gTimer.isAlarmed {
						// делаем фолд или чек
						p := gr.boutIter.Value.(*Player)
						p.lastActionLabel = LBL_FOLD
						actType := ActType(FOLD)
						if p.containsBoutVar(CHECK, "ALL-IN") {
							actType = ActType(CHECK)
							p.lastActionLabel = LBL_CHECK
						}

						gr.broadcast(&PlayerActionEvent{
							EventId:    GenId(),
							UserUid:    p.uid,
							ActionType: actType,
						}, true)
						f = true
					}
				}
				gr.gTimer.stop()
				// принять ответ о действии игрока. Проверить, что это действие ему доступно.
				// Если CHECK, то выходим из цикла торгового круга
			}

			gr.notifyBetAccepted()
		}

		gr.notifyWinnerResult()
		log.Info("Winner-pause 10s.")
		time.Sleep(time.Second * WinnerResultPauseSecond)

		for _, pIter := range gr.playerMap {
			p := pIter.Value.(*Player)
			if p.deposit <= 0 {
				gr.outcomeUser(pIter)
			}
		}
	}
	gr.setRoomState(DISSOLUTION)
}

func (gr *GameRoom) isEndOfTradeRound() bool {
	if gr.boutIter == nil {
		return false
	}
	p := gr.boutIter.Value.(*Player)
	defer p.resetBoutVars()

	//---
	numOfFold := 0
	numOfALLIN := 0
	for _, pIter := range gr.playerMap {
		if pIter.Value.(*Player).lastActionLabel == LBL_FOLD {
			numOfFold++
		}
		if pIter.Value.(*Player).lastActionLabel == LBL_ALLIN {
			numOfALLIN++
		}
	}
	if numOfFold >= gr.playerList.Len()-1 || numOfALLIN >= gr.playerList.Len() {
		gr.tradeRoundNumber = 4
		gr.notifyTableCards()
		return true
	}
	//---

	if gr.boutIter == gr.lastTraderIter {
		/*if p.lastActionLabel == LBL_RAISE {
			return false
		}
		if p.lastActionLabel == LBL_ALLIN {
			return !gr.isIncreasedBet
			// он мог сделать ALL-IN:
			// 	1) это повысило текущую ставку	-- продолжаем круг торгов (return false)
			// 	2) не повысило текущую ставку -- принимаем ставки, идем на след. круг торгов (return true)
		}*/
		return !gr.isIncreasedBet
	}
	return false
}

func (gr *GameRoom) notifyNewRound() {
	for _, pIter := range gr.playerMap {
		pIter.Value.(*Player).resetForNewRound()
	}
	gr.tableCardList = &[]*PlayingCard{}

	gr.shuffleDeck()
	gr.recalcBlinds()
	gr.broadcast(&GameEvent{
		EventId:     GenId(),
		EventType:   NEW_ROUND,
		RoundNumber: gr.roundNumber,
	}, true)
}

func (gr *GameRoom) notifySetDealer() {
	gr.dealerIter = gr.nextPlayer(gr.dealerIter)
	gr.broadcast(&PlayerActionEvent{
		EventId:    GenId(),
		UserUid:    gr.dealerIter.Value.(*Player).uid,
		ActionType: SET_DEALER,
	}, true)
}

func (gr *GameRoom) notifyNewTradeRound() {
	gr.boutIter = nil
	gr.currentBet = 0
	gr.ctrlBetSum = 0
	gr.lastTraderIter = nil

	for _, pIter := range gr.playerMap {
		pIter.Value.(*Player).resetForNewTradeRound()
	}

	//gr.lastRaiserIter = nil
	gr.broadcast(&GameEvent{
		EventId:     GenId(),
		EventType:   NEW_TRADE_ROUND,
		RoundNumber: int16(gr.roundNumber),
	}, true)
}

func (gr *GameRoom) notifyMinBlindIn() {
	p := gr.nextPlayer(gr.dealerIter).Value.(*Player)

	delta := int64(math.Min(float64(gr.minBlind), float64(p.deposit)))
	gr.ctrlBetSum += delta
	gr.currentBet = int64(math.Max(float64(gr.currentBet), float64(gr.minBlind)))

	gr.broadcast(p.setBlind(gr.minBlind, MIN_BLIND_IN), true)
}

func (gr *GameRoom) notifyMaxBlindIn() {
	pIter := gr.nextPlayer(gr.nextPlayer(gr.dealerIter))
	p := pIter.Value.(*Player)

	delta := int64(math.Min(float64(gr.maxBlind), float64(p.deposit)))
	gr.ctrlBetSum += delta
	gr.currentBet = int64(math.Max(float64(gr.currentBet), float64(gr.maxBlind)))

	gr.broadcast(p.setBlind(gr.maxBlind, MAX_BLIND_IN), true)

	//gr.lastRaiserIter = pIter
	gr.lastTraderIter = pIter
}

func (gr *GameRoom) notifyPersonalCards() {
	eventId := GenId()
	for _, pIter := range gr.playerMap {
		p := pIter.Value.(*Player)
		cards := make([]*PlayingCard, 2)
		for idx := range cards {
			cards[idx] = gr.getCard()
		}
		p.personalCardList = &cards
		p.findBestComb(&[]*PlayingCard{})
		bestCombName := ""
		/*if p.bestComb != nil {
			bestCombName = p.bestComb.name
		}*/
		gr.sendPersonalEvent(&GameEvent{
			EventId:          eventId,
			EventType:        PERSONAL_CARDS,
			PlayingCardsList: p.personalCardList,
			BestCombName:     bestCombName,
		}, p, true)
	}
}

func (gr *GameRoom) notifyBout() {
	if gr.boutIter != nil {
		gr.boutIter = gr.nextPlayer(gr.boutIter)
	}
	for gr.boutIter == nil ||
		gr.boutIter.Value.(*Player).lastActionLabel == LBL_ALLIN ||
		gr.boutIter.Value.(*Player).lastActionLabel == LBL_FOLD {

		if gr.boutIter == nil {
			if gr.tradeRoundNumber == 1 {
				gr.boutIter = gr.nextPlayer(gr.nextPlayer(gr.nextPlayer(gr.dealerIter)))
			} else {
				gr.boutIter = gr.nextPlayer(gr.dealerIter)
				gr.lastTraderIter = gr.boutIter
			}
		} else {
			gr.boutIter = gr.nextPlayer(gr.boutIter)
		}
	}

	eventId := GenId()
	p := gr.boutIter.Value.(*Player)

	if gr.currentBet == 0 {
		p.updateBoutVars(gr.maxBlind)
	} else {
		p.updateBoutVars(gr.currentBet)
	}
	p.timeEndBout = time.Now().Add(time.Second * BoutTime).Add(time.Millisecond * time.Duration(config.MsgLeewayMilli)).UnixMilli()
	bestCombName := ""
	if p.bestComb != nil {
		bestCombName = p.bestComb.name
	}

	gr.sendPersonalEvent(&PlayerActionEvent{
		EventId:              eventId,
		UserUid:              p.uid,
		ActionType:           BOUT,
		BoutVariants:         p.boutVariants,
		BestCombName:         bestCombName,
		TimeEndBoutOrForming: p.timeEndBout,
	}, p, true)

	for _, anotherPlIter := range gr.playerMap {
		anotherP := anotherPlIter.Value.(*Player)
		if anotherP == p {
			continue
		}
		gr.sendPersonalEvent(&PlayerActionEvent{
			EventId:              eventId,
			UserUid:              p.uid,
			ActionType:           BOUT,
			TimeEndBoutOrForming: p.timeEndBout,
		}, anotherP, true)
	}
}

func (gr *GameRoom) notifyBetAccepted() {
	gr.stack += gr.ctrlBetSum
	gr.broadcast(&GameEvent{
		EventId:   GenId(),
		EventType: BET_ACCEPTED,
		NewStack:  gr.stack,
	}, true)
}

func (gr *GameRoom) notifyTableCards() {
	switch gr.tradeRoundNumber {
	case 2:
		deck := make([]*PlayingCard, 3)
		for idx := 0; idx < len(deck); idx++ {
			deck[idx] = gr.getCard()
		}
		gr.tableCardList = &deck
	case 3:
		deck := make([]*PlayingCard, 4)
		copy(deck, *gr.tableCardList)
		for idx := len(*gr.tableCardList); idx < len(deck); idx++ {
			deck[idx] = gr.getCard()
		}
		gr.tableCardList = &deck
	case 4:
		deck := make([]*PlayingCard, 5)
		copy(deck, *gr.tableCardList)
		for idx := len(*gr.tableCardList); idx < len(deck); idx++ {
			deck[idx] = gr.getCard()
		}
		gr.tableCardList = &deck
	default:
		return
	}

	eventId := GenId()
	for _, pIter := range gr.playerMap {
		p := pIter.Value.(*Player)
		bestCombName := ""
		p.findBestComb(gr.tableCardList)
		if p.bestComb != nil {
			bestCombName = p.bestComb.name
		}
		gr.sendPersonalEvent(&GameEvent{
			EventId:          eventId,
			EventType:        CARDS_ON_TABLE,
			PlayingCardsList: gr.tableCardList,
			BestCombName:     bestCombName,
		}, p, true)
	}
}

func (gr *GameRoom) notifyWinnerResult() {
	/*
		1) алгоритм определения наилучшей комбы и ее названия по списку карт (длины от 2, 5, 6, 7)
		2) алгоритм определения лучшей комбинации в списке нескольких комбинаций
		3) обновление в БД статистики игроков, покидающих комнату
	*/
	idx := 0
	pUids := make([]string, gr.playerList.Len())
	closedCards := make([]*[]*PlayingCard, gr.playerList.Len())

	maxCombW := 0
	winners := list.New()
	for _, pIter := range gr.playerMap {
		p := pIter.Value.(*Player)
		pUids[idx] = p.uid
		closedCards[idx] = p.personalCardList
		idx++

		if p.bestComb != nil {
			if p.bestComb.weight > maxCombW {
				maxCombW = p.bestComb.weight
				winners = list.New()
				winners.PushBack(p)
			} else if p.bestComb.weight == maxCombW {
				winners.PushBack(p)
			}
		}
	}

	absWinners := list.New()
	maxWCard := 0
	for winIter := winners.Front(); winIter != nil; winIter = winIter.Next() {
		p := winIter.Value.(*Player)
		if p.bestComb != nil {
			if p.bestComb.wCard > maxWCard {
				maxWCard = p.bestComb.wCard
				absWinners = list.New()
				absWinners.PushBack(p)
			} else if p.bestComb.wCard == maxWCard {
				absWinners.PushBack(p)
			}
		}
	}

	winUids := make([]string, absWinners.Len())
	winDepos := make([]int64, absWinners.Len())
	bestCombs := make([]*[]*PlayingCard, absWinners.Len())
	bestCombName := ""
	idx = 0
	for winIter := absWinners.Front(); winIter != nil; winIter = winIter.Next() {
		p := winIter.Value.(*Player)
		p.deposit += gr.stack / int64(absWinners.Len())

		winUids[idx] = p.uid
		winDepos[idx] = p.deposit
		bestCombs[idx] = p.bestComb.cards
		bestCombName = p.bestComb.name
		idx++
	}
	gr.stack = 0

	gr.broadcast(&GameEvent{
		EventId:          GenId(),
		EventType:        WINNER_RESULT,
		ClosedCards:      &closedCards,
		PlayerUids:       &pUids,
		WinnerUids:       &winUids,
		BestCombinations: &bestCombs,
		BestCombName:     bestCombName,
		WinnerDeposits:   &winDepos,
		NewStack:         gr.stack,
	}, false)
}

func (gr *GameRoom) setRoomState(state RoomStateType) {
	log.Infof("Room %s: change state %s -> %s", gr.roomUid, gr.roomState, state)
	gr.roomState = state
	gr.broadcast(&GameEvent{
		EventId:      GenId(),
		EventType:    ROOM_STATE_UPDATE,
		NewRoomState: state,
	}, true)
}

func (gr *GameRoom) addPlayer(playerUid string) bool {
	if len(gr.playerMap) < MaxNumOfPlayersPerGame && gr.roomState == FORMING {
		gr.pushMsgQ(&ActionMessage{
			MessageType: GAME_ACTION,
			MessageId:   GenId(),
			RoomUid:     gr.roomUid,
			UserUid:     playerUid,
			ActionType:  INCOME,
		})
	}

	for len(gr.playerMap) < MaxNumOfPlayersPerGame && gr.roomState == FORMING {
		time.Sleep(10 * time.Millisecond)
		if gr.playerMap[playerUid] != nil {
			return true
		}
	}

	return false
}

func (gr *GameRoom) connectToRoom(wsReq WSRequest) bool {
	p := gr.playerMap[wsReq.PlayerUid].Value.(*Player)
	if p == nil {
		log.Errorf("Error in GameRoom.connectToRoom(). No such player with uid: %s", wsReq.PlayerUid)
		return false
	}
	if p.wsConn != nil && !p.wsConn.isTerminated {
		log.Errorf("Error in GameRoom.connectToRoom(). Old wsConn is not terminated")
		return false
	}

	p.wsConn = newWSConnection(wsReq.RW, wsReq.Req, wsReq.TokenValidator, p, gr)
	return p.wsConn != nil
}

/*
Добавляет сообщение *ActionMessage в очередь на обработку главной горутиной комнаты
*/
func (gr *GameRoom) pushMsgQ(msg *ActionMessage) bool {
	push := func(msg *ActionMessage) {
		gr.msgQ <- msg
	}

	check := []func() bool{
		func() bool { // запрос на фиксацию состояния комнаты или конкретного игрока
			return (msg.MessageId == -1 || msg.MessageId == -2) && msg.RoomUid == gr.roomUid
		},
		func() bool {
			return msg.RoomUid == gr.roomUid && msg.ActionType == INCOME && gr.roomState == FORMING
		},
		func() bool {
			return gr.validateMsg(msg)
		},
	}

	for _, ch := range check {
		if ch() {
			push(msg)
			return true
		}
	}

	return false
}

func (gr *GameRoom) fixRoomStateInfo(playerUid string) {
	p := gr.playerMap[playerUid].Value.(*Player)
	if p == nil {
		return
	}

	bout := ""
	if gr.boutIter != nil {
		bout = gr.boutIter.Value.(*Player).uid
	}

	dealer := ""
	if gr.dealerIter != nil {
		dealer = gr.dealerIter.Value.(*Player).uid
	}

	p.roomInfoCh <- &RoomInfo{
		RoomUid:           gr.roomUid,
		RoomState:         gr.roomState,
		PlayerList:        gr.makePlayerInfoList(p),
		TableCardList:     gr.tableCardList,
		Stack:             gr.stack,
		Bout:              bout,
		DealerUid:         dealer,
		LastEventId:       gr.lastEventId,
		NumOfPlayers:      gr.numOfPlayers,
		NumOfStartPlayers: gr.numOfStartPlayers,
	}
}

func (gr *GameRoom) fixPlayerInfo(pUid, pUidRelateOf string) {
	p := gr.playerMap[pUid].Value.(*Player)
	pRelateOf := gr.playerMap[pUidRelateOf].Value.(*Player)
	pRelateOf.playerInfoCh <- gr.playerToPlayerInfoRelateOf(p, pRelateOf)
}

/*
Формирует список инфомации об игроках. player *Player - игрок, относительно которого это делается
*/
func (gr *GameRoom) makePlayerInfoList(player *Player) *[]*PlayerInfo {
	res := make([]*PlayerInfo, gr.playerList.Len())
	for idx, iter := 0, gr.playerList.Front(); iter != nil; idx, iter = idx+1, iter.Next() {
		p := iter.Value.(*Player)

		res[idx] = gr.playerToPlayerInfoRelateOf(p, player)
	}

	return &res
}

/*
Выгружает PlayerInfo из Player. p -- указатель на игрока, инфо которого выгружаем. pRelateOf -- указатель на игрока для которого это делаем
*/
func (gr *GameRoom) playerToPlayerInfoRelateOf(p, pRelateOf *Player) *PlayerInfo {
	cardList := &[]*PlayingCard{}
	boutVars := &[]BoutVariantType{}
	bestCombName := ""
	timeEndBout := int64(0)
	voteType := WAIT

	if p.uid == pRelateOf.uid {
		cardList = p.personalCardList

		if p.bestComb != nil {
			bestCombName = p.bestComb.name
		}
	}

	switch e := p.eventQ.Back().Value.(type) {
	/*case *GameEvent:
	if e.EventType == WINNER_RESULT {
		cardList = p.personalCardList
	}*/
	case *PlayerActionEvent:
		if p.uid == pRelateOf.uid && e.ActionType == BOUT {
			boutVars = p.boutVariants
			timeEndBout = p.timeEndBout
		}
	}

	if p.isStartWanted {
		voteType = START
	}

	return &PlayerInfo{
		UserUid:          p.uid,
		Username:         p.username,
		ImageUrl:         "",
		Bet:              p.bet,
		Deposit:          p.deposit,
		LastActionLabel:  p.lastActionLabel,
		UserRank:         p.rank,
		PersonalCardList: cardList,
		BoutVariants:     boutVars,
		BestCombName:     bestCombName,
		TimeEndBout:      timeEndBout,
		VoteType:         voteType,
	}
}

/*
обработка сообщений-действий при roomState = GAMING. Возвращает true, если получено корректное сообщение от того игрока,
которому принадлежит очередь хода
*/
func (gr *GameRoom) playerActionHandling(msg *ActionMessage) bool {
	if gr.roomState != GAMING {
		return false
	}

	p := gr.boutIter.Value.(*Player)

	switch msg.MessageType {
	case GAME_ACTION:
		if msg.ActionType == OUTCOME {
			gr.outcomeUser(gr.playerMap[msg.UserUid])
			return p.uid == msg.UserUid
		} else {
			if p.uid != msg.UserUid {
				return false
			}
			if p.containsBoutVar(BVariantType(msg.ActionType), msg.Coef) {
				actType := ActType(msg.ActionType)
				gr.isIncreasedBet = false

				switch msg.ActionType {
				case PlayerActionType(FOLD):
					p.lastActionLabel = LBL_FOLD
				case PlayerActionType(CHECK):
					p.lastActionLabel = LBL_CHECK // завершать круг торгов
				case PlayerActionType(CALL):
					p.lastActionLabel = LBL_CALL

					if gr.currentBet == 0 {
						gr.isIncreasedBet = true
						gr.currentBet = gr.maxBlind
						gr.lastTraderIter = gr.boutIter
					}
					p.deposit -= (gr.currentBet - p.bet)
					gr.ctrlBetSum += (gr.currentBet - p.bet)
					p.bet = gr.currentBet
				case PlayerActionType(RAISE): // сдвигать сюда lastRaiserIter, lastTraderIter
					delta := int64(0)
					if msg.Coef == ALLIN {
						p.lastActionLabel = LBL_ALLIN
						actType = ALL_IN

						delta = p.deposit
						if delta+p.bet > gr.currentBet { // чекаем повышается ли ставка от ALL-IN игрока
							//gr.lastRaiserIter = gr.boutIter // сдвинул
							gr.lastTraderIter = gr.boutIter
							gr.isIncreasedBet = true
						}
					} else {
						p.lastActionLabel = LBL_RAISE
						if msg.Coef == X1_5 {
							delta = 3*gr.currentBet/2 - p.bet // p.bet
						} else if msg.Coef == X2 {
							delta = 2*gr.currentBet - p.bet // p.bet
						}

						//gr.lastRaiserIter = gr.boutIter // сдвинул
						gr.lastTraderIter = gr.boutIter
						gr.isIncreasedBet = true
					}

					p.bet += delta
					gr.ctrlBetSum += delta
					p.deposit -= delta
					gr.currentBet = int64(math.Max(float64(gr.currentBet), float64(p.bet)))
				}

				gr.broadcast(&PlayerActionEvent{
					EventId:    GenId(),
					UserUid:    p.uid,
					ActionType: actType,
					NewBet:     p.bet,
					NewDeposit: p.deposit,
				}, true)
				return true
			}
		}
	case VOTE:
		log.Info("VOTE is incorrect")
		return false
	}
	return false
}

/*
обработка сообщений входа/выхода при roomState = FORMING
*/
func (gr *GameRoom) ioPrepareActionHandling(msg *ActionMessage) {
	if gr.roomState != FORMING {
		return
	}

	switch msg.ActionType {
	case INCOME:
		p := newPlayer(msg.UserUid, gr.repo, gr.getInitEventQCopy())
		gr.playerMap[msg.UserUid] = gr.playerList.PushBack(p)
		gr.numOfPlayers++
		event := &PlayerActionEvent{
			EventId:    GenId(),
			UserUid:    msg.UserUid,
			ActionType: ActType(INCOME),
		}
		gr.broadcast(event, true)

		if len(gr.playerMap) >= MinNumOfPlayersPerGame {
			gr.gTimer.start(time.Duration(WaitStartTimeSecond * time.Second))
		}
	case OUTCOME:
		iter := gr.playerMap[msg.UserUid]
		if iter.Value.(*Player).isStartWanted {
			gr.numOfStartPlayers--
		}

		gr.outcomeUser(iter)
		if len(gr.playerMap) < MinNumOfPlayersPerGame {
			gr.gTimer.stop()
		}
	}
}

func (gr *GameRoom) getInitEventQCopy() *list.List {
	if len(gr.playerMap) <= 0 {
		return list.New()
	}
	res := list.New()
	keys := make([]string, 0, 1)
	for k := range gr.playerMap {
		keys = append(keys, k)
		break
	}
	p := gr.playerMap[keys[0]].Value.(*Player)
	res.PushBackList(p.eventQ)
	return res
}

/*
обработка сообщения-голосования при roomState = FORMING
*/
func (gr *GameRoom) voteActionHandling(msg *ActionMessage) {
	if gr.roomState != FORMING || msg.MessageType != VOTE {
		return
	}

	p := gr.playerMap[msg.UserUid].Value.(*Player)
	switch msg.VoteType {
	case START:
		if !p.isStartWanted {
			p.isStartWanted = true
			gr.numOfStartPlayers++
		}
	case WAIT:
		if p.isStartWanted {
			p.isStartWanted = false
			gr.numOfStartPlayers--
		}
	default:
		return
	}
	gr.broadcast(&PrepareEvent{
		EventId:           GenId(),
		NumOfPlayers:      gr.numOfPlayers,
		NumOfStartPlayers: gr.numOfStartPlayers,
	}, true)
}

func (gr *GameRoom) broadcast(event any, doUpdateLastEventId bool) {
	if doUpdateLastEventId {
		switch e := event.(type) {
		case *PlayerActionEvent:
			gr.lastEventId = e.EventId
		case *PrepareEvent:
			gr.lastEventId = e.EventId
		case *GameEvent:
			gr.lastEventId = e.EventId
		}
	}
	for _, p := range gr.playerMap {
		p.Value.(*Player).pushEvent(event)
	}
}

func (gr *GameRoom) sendPersonalEvent(event any, p *Player, doUpdateLastEventId bool) {
	if doUpdateLastEventId {
		switch e := event.(type) {
		case *PlayerActionEvent:
			gr.lastEventId = e.EventId
		case *PrepareEvent:
			gr.lastEventId = e.EventId
		case *GameEvent:
			gr.lastEventId = e.EventId
		}
	}
	p.pushEvent(event)
}

func GenId() int64 {
	time.Sleep(time.Nanosecond)
	return time.Now().UnixNano()
}

func (gr *GameRoom) GetRoomInfo(playerUid string) *RoomInfo {
	if p := gr.playerMap[playerUid].Value.(*Player); p != nil {
		gr.pushMsgQ(&ActionMessage{
			MessageId:   -1,
			MessageType: GAME_ACTION,
			RoomUid:     gr.roomUid,
			UserUid:     playerUid,
		})
		return <-p.roomInfoCh
	}

	return nil
}

/*
Получает PlayerInfo по playerUid. pUidRelateOf -- uid игрока, для которого делается выгрузка
*/
func (gr *GameRoom) GetPlayerInfo(playerUid, pUidRelateOf string) *PlayerInfo {
	if gr.playerMap[playerUid].Value.(*Player) == nil {
		return nil
	}
	if p := gr.playerMap[pUidRelateOf].Value.(*Player); p != nil {
		gr.pushMsgQ(&ActionMessage{
			MessageId:   -2,
			MessageType: GAME_ACTION,
			RoomUid:     gr.roomUid,
			UserUid:     playerUid + "." + pUidRelateOf,
		})
		return <-p.playerInfoCh
	}
	return nil
}

func (gr *GameRoom) validateMsg(msg *ActionMessage) bool {
	if msg.RoomUid != gr.roomUid {
		return false
	}
	iterP := gr.playerMap[msg.UserUid]
	if iterP == nil {
		return false
	}

	p := iterP.Value.(*Player)
	switch msg.MessageType {
	case VOTE:
		return gr.roomState == FORMING
	case GAME_ACTION:
		if gr.roomState == DISSOLUTION {
			return false
		}
		if msg.ActionType == PlayerActionType(CHECK) || msg.ActionType == PlayerActionType(FOLD) ||
			msg.ActionType == PlayerActionType(CALL) || msg.ActionType == PlayerActionType(RAISE) {
			if gr.boutIter.Value.(*Player).uid != p.uid {
				return false
			}
			return p.containsBoutVar(BVariantType(msg.ActionType), msg.Coef)
		}
	default:
		return false
	}
	return true
}

/*
Возвращает указатель на следующего игрока относительно переданного в параметре
*/
func (gr *GameRoom) nextPlayer(playerIter *list.Element) *list.Element {
	if playerIter == nil {
		return gr.playerList.Front()
	} else {
		res := playerIter.Next()
		if res == nil {
			res = gr.playerList.Front()
		}
		return res
	}
}

/*
пересчитывает размеры блайндов каждые 2 раунда
*/
func (gr *GameRoom) recalcBlinds() {
	gr.minBlind = int64(StartMinBlind) * (int64(gr.roundNumber)/2 + 1)
	gr.maxBlind = 2 * gr.minBlind
}

/*
инициализирует колоду
*/
func (gr *GameRoom) initCardDeck() {
	deck := make([]*PlayingCard, len(cardSuits)*len(cardIndexes))
	for idxS, suit := range cardSuits {
		for idxI, index := range cardIndexes {
			deck[idxS*len(cardIndexes)+idxI] = NewPlayingCard(suit, index)
		}
	}
	gr.deckOfCards = &deck
}

/*
тассует колоду
*/
func (gr *GameRoom) shuffleDeck() {
	gr.initCardDeck()
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	rand := func() uint64 {
		return gen.Uint64()
	}

	for i := len(*gr.deckOfCards) - 1; i > 1; i-- {
		j := rand() % uint64(i+1)
		(*gr.deckOfCards)[i], (*gr.deckOfCards)[j] = (*gr.deckOfCards)[j], (*gr.deckOfCards)[i]
	}
	gr.cardIdx = 0
}

/*
Возвращает указатель на верхнюю карту колоды, "изымая" карту из колоды. И переходит к следующей карте
*/
func (gr *GameRoom) getCard() *PlayingCard {
	if gr.cardIdx >= int16(len(*gr.deckOfCards)) {
		return nil
	}
	gr.cardIdx++
	return (*gr.deckOfCards)[gr.cardIdx-1]
}

/*
обрабатывает событие покидания комнаты игроком
*/
func (gr *GameRoom) outcomeUser(playerIterator *list.Element) {
	p := playerIterator.Value.(*Player)

	gr.gameBalancer.removePlayerFromRoom(p.uid)

	if gr.roomState == GAMING {
		if playerIterator == gr.dealerIter {
			if playerIterator.Prev() != nil {
				gr.dealerIter = playerIterator.Prev()
			} else {
				gr.dealerIter = gr.playerList.Back()
			}
		}
		if playerIterator == gr.boutIter {
			if playerIterator.Next() != nil {
				gr.boutIter = playerIterator.Next()
			} else {
				gr.boutIter = gr.playerList.Back()
			}
		}
		if playerIterator == gr.lastTraderIter {
			if playerIterator.Prev() != nil {
				gr.lastTraderIter = playerIterator.Prev()
			} else {
				gr.lastTraderIter = gr.playerList.Back()
			}
		}
		/*if playerIterator == gr.lastRaiserIter {
			gr.lastRaiserIter = nil
		}*/
	}

	gr.playerList.Remove(playerIterator)
	delete(gr.playerMap, p.uid)
	gr.numOfPlayers--
	event := &PlayerActionEvent{
		EventId:    GenId(),
		UserUid:    p.uid,
		ActionType: ActType(OUTCOME),
	}
	//p.wsConn.close()
	gr.broadcast(event, true)
}
