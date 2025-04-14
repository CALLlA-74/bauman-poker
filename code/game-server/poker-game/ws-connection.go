package pokergame

import (
	"bauman-poker/config"
	external "bauman-poker/external-services"
	"bauman-poker/utils"
	"container/list"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WSConnection struct {
	ws             *websocket.Conn
	player         *Player
	room           *GameRoom
	tokenValidator *utils.TokenValidator
	expAuthTime    int64
	msgQForSend    list.List //chan any
	respMsgQ       chan *ResponseMessage
	isSending      bool
	isTerminated   bool
	lastPingTime   int64 // момент времени последнего пинга в мс
	lastMsgTime    int64 // момент времени получения последнего сообщения от клиента
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func newWSConnection(rw http.ResponseWriter, r *http.Request, tv *utils.TokenValidator, p *Player, gr *GameRoom) *WSConnection {
	ws, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.WithError(err).Error("error in creating WS-Connection")
		return nil
	}

	wsc := &WSConnection{
		ws:             ws,
		player:         p,
		room:           gr,
		tokenValidator: tv,
		expAuthTime:    0,
		msgQForSend:    *list.New(), //make(chan any, 1000),
		respMsgQ:       make(chan *ResponseMessage),
		isSending:      false,
		isTerminated:   false,
		lastPingTime:   0,
		lastMsgTime:    time.Now().UnixMilli(),
	}
	wsc.wsReader()
	wsc.wsMsgSender()

	return wsc
}

func (wsc *WSConnection) wsReader() {
	go func() {
		defer wsc.close()

		for !wsc.isTerminated {
			log.Info("next read")
			_, bmsg, err := wsc.ws.ReadMessage()
			wsc.lastMsgTime = time.Now().UnixMilli()
			log.Infof("%v", bmsg)

			if err != nil {
				log.WithError(err).Errorf("Error in reading msg")
				return
			} /*else if msgType != websocket.TextMessage {
				log.Errorf("Expected type: Text message!")
				return
			} */

			switch msg := UnpackMsgFromPlayer(bmsg).(type) {
			case (*PongMessage):
				log.Info("Casted to PongMsg")
			case (*AuthMessage):
				log.Info("Casted to AuthMsg")
				if !wsc.updateAuthToken(*msg) {
					log.Error("Auth-error")
				}
			case (*ActionMessage):
				log.Info("Casted to ActionMsg")
				if !wsc.authIsExpired() {
					if msg.UserUid != wsc.player.uid || !wsc.room.pushMsgQ(msg) {
						wsc.sendRespMsg(msg.MessageId, StatusBadReq)
					} else {
						wsc.sendRespMsg(msg.MessageId, StatusOK)
					}
				} else {
					wsc.sendRespMsg(msg.MessageId, StatusUnauthorized)
				}
			default:
				log.Info("Untyped msg")
			}
		}
	}()

	go func() {
		for !wsc.isTerminated {
			if time.Now().UnixMilli()-wsc.lastMsgTime > config.PingPeriodMilli+config.MsgLeewayMilli {
				log.Errorf("PONG TIME ERROR")
				wsc.close()
				return
			}
		}
	}()
}

func NewEventMessage(event any) any {
	switch e := event.(type) {
	case *GameEvent:
		return &EventMessage[GameEvent]{
			MessageType:     EVENT,
			MessageId:       GenId(),
			EventType:       GAME_EVENT,
			EventDescriptor: e,
		}
	case *PrepareEvent:
		return &EventMessage[PrepareEvent]{
			MessageType:     EVENT,
			MessageId:       GenId(),
			EventType:       PREPARE_EVENT,
			EventDescriptor: e,
		}
	case *PlayerActionEvent:
		return &EventMessage[PlayerActionEvent]{
			MessageType:     EVENT,
			MessageId:       GenId(),
			EventType:       PLAYER_ACTION_EVENT,
			EventDescriptor: e,
		}
	}

	return nil
}

func (wsc *WSConnection) authIsExpired() bool {
	return wsc.expAuthTime+config.LeewaySeconds < time.Now().Unix()
}

func (wsc *WSConnection) updateAuthToken(msg AuthMessage) bool {
	if !wsc.tokenValidator.VerifyAccessToken(msg.Token) {
		wsc.sendRespMsg(msg.MessageId, StatusUnauthorized)
		return false
	}
	_, p, _, _ := wsc.tokenValidator.ParseAccessToken(msg.Token)
	payload := p.(*utils.AccessTokenPayload)
	if payload.UserUid != wsc.player.uid {
		wsc.sendRespMsg(msg.MessageId, StatusUnauthorized)
		return false
	}

	wsc.sendRespMsg(msg.MessageId, StatusOK)

	//if wsc.authIsExpired() {
	wsc.player.setCurrentEvent(msg.LastEventId)
	//}
	wsc.expAuthTime = payload.Exp
	return true
}

func (wsc *WSConnection) wsMsgSender() {
	wsc.isSending = true

	send := func(msg any) {
		if wsc.ws == nil {
			wsc.isSending = false
			wsc.isTerminated = true
		}
		if err := wsc.ws.WriteJSON(msg); err != nil {
			log.WithError(err).Errorf("Error in wsMsgSender")
			wsc.close()
			/*if wsc.ws != nil {
				wsc.ws.Close()
			}
			wsc.isSending = false
			wsc.isTerminated = true*/
			return
		}
	}

	go func() {
		for !wsc.isTerminated {
			if time.Now().UnixMilli()-wsc.lastPingTime > config.PingPeriodMilli {
				wsc.lastPingTime = time.Now().UnixMilli()
				send(&PingMessage{
					MessageType: PING,
				})
			}

			select {
			case msg, ok := <-wsc.respMsgQ:
				if ok {
					log.Info("try to send respMsg")
					send(msg)
				}
			default:
			}

			if !wsc.authIsExpired() {
				if event := wsc.player.getNextEvent(); event != nil {
					msg := NewEventMessage(event)
					send(msg)
					if wsc.isSelftCloseEvent(event) {
						log.Infof("prepare to close-conn")
						time.Sleep(time.Second * 2)
						wsc.close()
						return
					}
				}
			}
		}
	}()
}

func (wsc *WSConnection) sendRespMsg(msgId int64, statusCode RespStatusCodeType) {
	wsc.respMsgQ <- &ResponseMessage{
		MessageType:  ACK,
		AckMessageId: msgId,
		StatusCode:   statusCode,
	}
}

func UnpackMsgFromPlayer(bmsg []byte) any {
	pongMsg := &PongMessage{}
	authMsg := &AuthMessage{}
	actionMsg := &ActionMessage{}
	if err2 := external.Unpack(bmsg, authMsg); err2 == nil {
		return authMsg
	} else if err3 := external.Unpack(bmsg, actionMsg); err3 == nil {
		return actionMsg
	} else if err := external.Unpack(bmsg, pongMsg); err == nil {
		return pongMsg
	} else {
		log.WithError(err).WithError(err2).WithError(err3).Errorf("Errors in unpacking msgs from player")
		return nil
	}
}

func (wsc *WSConnection) isSelftCloseEvent(event any) bool {
	switch e := event.(type) {
	case *PlayerActionEvent:
		if e.ActionType == ActType(OUTCOME) {
			return e.UserUid == wsc.player.uid
		}
	}
	return false
}

func (wsc *WSConnection) close() {
	wsc.isTerminated = true
	wsc.isSending = false
	if wsc.ws != nil {
		wsc.ws.Close()
	}
}
