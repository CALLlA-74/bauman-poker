from pydantic import BaseModel
from uuid import UUID
from typing import Literal, List

class PlayingCard(BaseModel):
    cardSuit: Literal['DIAMONDS', 'HEARTS', 'CLUBS', 'SPADES']
    index: Literal['2', '3', '4', '5', '6', '7', '8', '9', '10', 'ACE', 'KING', 'QUEEN', 'JACK']

class UserInfo(BaseModel):
    userUid: UUID
    username: str
    numOfGames: int
    numOfWins: int
    userRank: Literal['РЕКРУТ', 'РЯДОВОЙ', 'СЕРЖАНТ', 'КАПИТАН', 'МАЙОР', 'ПОЛКОВНИК', 'ГЕНЕРАЛ']
    userState: Literal['IN-GAME', 'MENU']
    roomUid: UUID

class PlayerInfo(BaseModel):
    userUid: UUID
    username: str
    bet: int
    deposit: int
    lastActionLabel: Literal['NONE', 'FOLD', 'CHECK', 'CALL', 'RAISE', 'ALL-IN']
    userRank: Literal['РЕКРУТ', 'РЯДОВОЙ', 'СЕРЖАНТ', 'КАПИТАН', 'МАЙОР', 'ПОЛКОВНИК', 'ГЕНЕРАЛ']
    personalCardList: List[PlayingCard]

class RoomInfo(BaseModel):
    roomUid: UUID
    roomState: Literal['FORMING', 'GAMING', 'DISSOLUTION']
    playerList: List[PlayerInfo]
    tableCardList: List[PlayingCard]
    stack: int
    bout: UUID
    lastEventId: int
    roundNumber: int
