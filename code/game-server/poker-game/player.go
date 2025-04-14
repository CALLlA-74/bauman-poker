package pokergame

import (
	"bauman-poker/repo"
	"container/list"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Player struct {
	uid              string
	username         string
	bet              int64
	deposit          int64
	lastActionLabel  LastActionLabelType
	rank             repo.RankType
	personalCardList *[]*PlayingCard
	boutVariants     *[]BoutVariantType
	bestComb         *BestComb
	//bestCombName     string
	timeEndBout   int64
	eventQ        *list.List
	lastEventIter *list.Element
	lastEvent     any
	lastEventMtx  *sync.Mutex
	wsConn        *WSConnection
	roomInfoCh    chan *RoomInfo   // канал, через который передается инфа о текущем состоянии комнаты
	playerInfoCh  chan *PlayerInfo // канал, через который передается инфа о запрошенном PlayerInfo
	isStartWanted bool
}

func newPlayer(userUid string, rp *repo.GormPlayerRepo, initEventQ *list.List) *Player {
	playerAcc, err := rp.GetPlayerByUid(userUid)
	if err != nil {
		log.Errorf("Error in Player.newPlayer()")
		return nil
	}
	if initEventQ == nil {
		log.Errorf("Error in Player.newPlayer(). initEvenQ is nil")
		return nil
	}
	return &Player{
		uid:              userUid,
		username:         playerAcc.Username,
		bet:              0,
		deposit:          StartDeposit,
		lastActionLabel:  LBL_NONE,
		rank:             playerAcc.UserRank,
		personalCardList: &[]*PlayingCard{},
		boutVariants:     &[]BoutVariantType{},
		//bestCombName:     "",
		bestComb:      nil,
		timeEndBout:   0,
		eventQ:        initEventQ,
		lastEventIter: nil, //initEventQ.Back(), // потому что все события до добавления игрока были переданы ему в *RoomInfo
		lastEvent:     nil,
		lastEventMtx:  new(sync.Mutex),
		roomInfoCh:    make(chan *RoomInfo),
		playerInfoCh:  make(chan *PlayerInfo),
		isStartWanted: false,
	}
}

func (p *Player) pushEvent(e any) {
	// p.lastEvent =
	p.eventQ.PushBack(e)
	/*if p.wsConn != nil && !p.wsConn.isTerminated {
		p.wsConn.eventMsgManager(p.eventQ.PushBack(e))
	}*/
}

func (p *Player) setCurrentEvent(eventId int64) {
	id := int64(-100)
	for eIter := p.eventQ.Front(); eIter != nil; eIter = eIter.Next() {
		switch e := eIter.Value.(type) {
		case *GameEvent:
			id = int64(e.EventId)
		case *PrepareEvent:
			id = int64(e.EventId)
		case *PlayerActionEvent:
			id = int64(e.EventId)
		}
		if eventId == id {
			if eIter.Next() != nil {
				p.lastEventMtx.Lock()
				p.lastEventIter = eIter.Next()
				p.lastEvent = nil
				p.lastEventMtx.Unlock()
			} else {
				p.lastEventMtx.Lock()
				p.lastEventIter = eIter
				p.lastEvent = p.lastEventIter.Value
				p.lastEventMtx.Unlock()
			}
			break
		}
	}
}

/*
возвращает указатель на очередной Event на отправку. Если новых ивентов нет, то вернет nil
*/
func (p *Player) getNextEvent() any {
	defer p.lastEventMtx.Unlock()

	p.lastEventMtx.Lock()
	if p.lastEventIter == nil {
		if p.eventQ.Len() <= 0 {
			return nil
		}
		p.lastEventIter = p.eventQ.Front()
	}
	res := p.lastEventIter.Value
	if p.lastEventIter.Next() != nil {
		p.lastEventIter = p.lastEventIter.Next()
	}
	if p.lastEvent == res {
		return nil
	}

	p.lastEvent = res
	return res
}

func (p *Player) containsBoutVar(variantType BVariantType, coef CoefType) bool {
	for _, v := range *p.boutVariants {
		if v.VariantType == variantType {
			if v.VariantType == RAISE {
				for _, c := range *v.RaiseVariants {
					if c == coef {
						return true
					}
				}
				break
			} else {
				return true
			}
		}
	}
	return false
}

func (p *Player) resetForNewRound() {
	p.bet = 0
	p.lastActionLabel = LBL_NONE
	p.personalCardList = &[]*PlayingCard{}
	//p.bestCombName = ""
	p.bestComb = nil
}

func (p *Player) resetForNewTradeRound() {
	p.bet = 0
	if p.lastActionLabel != LBL_FOLD && p.lastActionLabel != LBL_ALLIN {
		p.lastActionLabel = LBL_NONE
	}
}

/*
Обновляет значение поля BoutVariants в зависимости от размера депозита, текущего размера ставки на столе. isCheck если true,

	будет добавлен вариант CHECK, иначе будет добавлен FOLD
*/
func (p *Player) updateBoutVars(currentBet int64) {
	if p.lastActionLabel == LBL_ALLIN || p.lastActionLabel == LBL_FOLD {
		return
	}

	bouts := list.New()
	if currentBet == p.bet {
		bouts.PushBack(BoutVariantType{
			VariantType: CHECK,
		})
	} else {
		bouts.PushBack(BoutVariantType{
			VariantType: FOLD,
		})
	}

	if currentBet-p.bet < p.deposit && currentBet != p.bet {
		bouts.PushBack(BoutVariantType{
			VariantType: CALL,
			CallValue:   currentBet - p.bet,
		})
	}

	if newBet := currentBet * 2; newBet-p.bet < p.deposit {
		bouts.PushBack(BoutVariantType{
			VariantType:   RAISE,
			RaiseVariants: &[]CoefType{X1_5, X2, ALLIN},
		})
	} else if newBet := 3 * currentBet / 2; newBet-p.bet < p.deposit {
		bouts.PushBack(BoutVariantType{
			VariantType:   RAISE,
			RaiseVariants: &[]CoefType{X1_5, ALLIN},
		})
	} else {
		bouts.PushBack(BoutVariantType{
			VariantType:   RAISE,
			RaiseVariants: &[]CoefType{ALLIN},
		})
	}

	newBouts := make([]BoutVariantType, bouts.Len())
	for bIter, idx := bouts.Front(), 0; bIter != nil; bIter, idx = bIter.Next(), idx+1 {
		newBouts[idx] = bIter.Value.(BoutVariantType)
	}
	p.boutVariants = &newBouts
}

func (p *Player) resetBoutVars() {
	p.boutVariants = &[]BoutVariantType{}
}

/*
Вносит блайнд, если возможно. Иначе делает ALL-IN. blindTypeAction тип блайнда:

	MIN_BLIND_IN | MAX_BLIND_IN
*/
func (p *Player) setBlind(blindSize int64, blindTypeAction ActType) *PlayerActionEvent {
	if blindSize < p.deposit {
		p.bet += blindSize
		p.deposit -= blindSize
		return &PlayerActionEvent{
			EventId:    GenId(),
			UserUid:    p.uid,
			ActionType: blindTypeAction,
			NewBet:     p.bet,
			NewDeposit: p.deposit,
		}
	} else {
		p.bet = p.deposit
		p.deposit = 0
		return &PlayerActionEvent{
			EventId:    GenId(),
			UserUid:    p.uid,
			ActionType: ALL_IN,
			NewBet:     p.bet,
			NewDeposit: p.deposit,
		}
	}
}

/*
На вход подается список открытых карт на столе. Функция ищет лучшую комбу среди открытых + своих карт
*/
func (p *Player) findBestComb(tableCards *[]*PlayingCard) {
	cards := make([]*PlayingCard, len(*tableCards)+len(*p.personalCardList))
	if len(cards) < 2 {
		return
	}
	log.Info("Try to find best Comb")

	copy(cards, *p.personalCardList)
	for i := range *tableCards {
		cards[i+2] = (*tableCards)[i]
	}

	p.bestComb = GetBestComb(cards)
	if p.bestComb != nil {
		log.Infof("BestCombName (p username: %s): %v", p.username, p.bestComb)
	} else {
		log.Infof("BestCombName (p username: %s): no comb :(", p.username)
	}
}
