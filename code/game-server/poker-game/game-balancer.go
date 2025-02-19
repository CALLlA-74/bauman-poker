package pokergame

import (
	"bauman-poker/repo"
	"container/list"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type GameBalancer struct {
	freeRooms          *list.List
	allRooms           map[string]*GameRoom // uid комнаты -> указатель на комнату
	playerToRoom       map[string]*GameRoom // распределение: uid игрока -> указатель на комнату
	unallocatedPlayers chan string
	repo               *repo.GormPlayerRepo
}

const (
	lenOfFreeRoomsQ       = 10
	lenOfAllocatedPlayerQ = 100
)

func NewGameBalancer(rp *repo.GormPlayerRepo) *GameBalancer {
	gb := &GameBalancer{
		freeRooms:          list.New(),
		allRooms:           make(map[string]*GameRoom),
		playerToRoom:       make(map[string]*GameRoom),
		unallocatedPlayers: make(chan string, lenOfAllocatedPlayerQ),
		repo:               rp,
	}

	go func() { // горутина распределения игроков по комнатам
		for {
			playerUid := <-gb.unallocatedPlayers
			if gb.playerToRoom[playerUid] != nil {
				continue
			}

			for {
				if gb.freeRooms.Len() <= 0 {
					room := newRoom(rp, gb)
					gb.freeRooms.PushBack(room)
					gb.allRooms[room.roomUid] = room
				}
				room := gb.freeRooms.Front().Value.(*GameRoom)

				if !room.addPlayer(playerUid) {
					gb.freeRooms.Remove(gb.freeRooms.Front())
				} else {
					gb.playerToRoom[playerUid] = room
					break
				}
			}
		}
	}()

	go func() { // горутина, удаляющая комнаты с завершенной игрой
		for {
			for u, r := range gb.allRooms {
				if r.roomState == DISSOLUTION {
					delete(gb.allRooms, u)
				}
			}
		}
	}()

	return gb
}

func (gb *GameBalancer) GetRoomUidByPlayer(playerUid string) string {
	if room := gb.playerToRoom[playerUid]; room != nil {
		return room.roomUid
	}
	return ""
}

func (gb *GameBalancer) MatchingRoom(playerUid string) *RoomInfo {
	if room := gb.playerToRoom[playerUid]; room != nil {
		return room.GetRoomInfo(playerUid)
	}
	gb.unallocatedPlayers <- playerUid

	room := gb.playerToRoom[playerUid]
	for room == nil {
		time.Sleep(10 * time.Millisecond)
		room = gb.playerToRoom[playerUid]
	}
	return room.GetRoomInfo(playerUid)
}

func (gb *GameBalancer) GetRoom(roomUid, playerUid string) *RoomInfo {
	return gb.allRooms[roomUid].GetRoomInfo(playerUid)
}

func (gb *GameBalancer) ConnectToRoom(wsReq WSRequest) bool {
	room := gb.allRooms[wsReq.RoomUid]

	if room == nil {
		log.Errorf("Error in GameBalancer.ConnectToRoom(). No such room with Uid: %s", wsReq.RoomUid)
		return false
	}

	return room.connectToRoom(wsReq)
}

func (gb *GameBalancer) GetPlayerInfo(playerUid, pUidRelateOf string) (*PlayerInfo, error) {
	if r := gb.playerToRoom[pUidRelateOf]; r != nil {
		if res := r.GetPlayerInfo(playerUid, pUidRelateOf); res != nil {
			return res, nil
		}
		return nil, fmt.Errorf("500")
	}
	return nil, fmt.Errorf("404")
}

func (gb *GameBalancer) removePlayerFromRoom(playerUid string) {
	delete(gb.playerToRoom, playerUid)
}
