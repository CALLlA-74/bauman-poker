package repo

import (
	"gorm.io/gorm"
)

type RankType string

const (
	RECREUIT RankType = "РЕКРУТ"
	SOLDIER  RankType = "РЯДОВОЙ"
	SERGEANT RankType = "СЕРЖАНТ"
	CAPTAIN  RankType = "КАПИТАН"
	MAJOR    RankType = "МАЙОР"
	COLONEL  RankType = "ПОЛКОВНИК"
	GENERAL  RankType = "ГЕНЕРАЛ"
)

type UserStateType string

const (
	IN_GAME UserStateType = "IN-GAME"
	MENU    UserStateType = "MENU"
)

type PlayerAccount struct {
	gorm.Model

	ID         int64    `gorm:"column:id;primaryKey;autoIncrement"`
	Uid        string   `gorm:"column:uid;type:varchar(100);not null;unique"`
	Username   string   `gorm:"column:username;type:varchar(100);not null;unique"`
	NumOfGames int64    `gorm:"column:num_of_games;type:integer;not null;default:0"`
	NumOfWins  int64    `gorm:"column:num_of_wins;type:integer;not null;default:0"`
	UserRank   RankType `gorm:"column:user_rank;type:varchar;not null;default:'РЕКРУТ'"`
	//UserState  UserStateType `gorm:"column:user_state;type:varchar;not null;default:'MENU'"`
	//RoomUid    string        `gorm:"column:room_uid;type:varchar;not null;default:''"`
}
