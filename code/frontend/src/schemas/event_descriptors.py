from pydantic import BaseModel
from uuid import UUID
from typing import Literal, List

# from info import PlayingCard

class PlayingCard(BaseModel):
    cardSuit: Literal['DIAMONDS', 'HEARTS', 'CLUBS', 'SPADES']
    index: Literal['2', '3', '4', '5', '6', '7', '8', '9', '10', 'ACE', 'KING', 'QUEEN', 'JACK']


class BoutVariant(BaseModel):
    variantType: Literal['FOLD', 'CHECK', 'CALL', 'RAISE']
    callValue: int | None = None
    raiseVariants: Literal['X1_5', 'X2', 'ALL-IN'] | None = None

class PlayerActionEvent(BaseModel):
    eventId: int
    userUid: UUID
    actionType: Literal['INCOME', 'OUTCOME', 'BOUT', 'FOLD', 'CHECK', 'CALL', 'RAISE', 'SET-DEALER', 'MIN-BLIN-IN', 'MAX-BLIN-IN']
    boutVariants: List[BoutVariant]
    bestCombName: str | None = None
    timeStartBout: int | None = None
    newStack: int | None = None
    newBet: int | None = None
    blindSize: int | None = None
    newDeposit: int | None = None

class GameEvent(BaseModel):
    eventId: int
    eventType: Literal['ROOM_STATE_UPDATE', 'PERSONAL_CARDS', 'CARDS_ON_TABLE', 'WINNER_RESULT']
    newRoomState: Literal['FORMING', 'GAMING', 'DISSOLUTION'] | None = None
    playingCardsList: List[PlayingCard] | None = None
    closedCards: List[PlayingCard] | None = None
    playerUids: List[UUID] | None = None
    winnerUids: List[UUID] | None = None
    bestCombinations: List[PlayingCard] | None = None
    bestCombName: str | None = None
    winnerDeposits: int | None = None
    newStack: int | None = None