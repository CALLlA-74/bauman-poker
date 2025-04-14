from pydantic import BaseModel
from uuid import UUID
from typing import Literal, List

class PongMessage(BaseModel):
    MessageType: str = 'PONG'

class AuthMessage(BaseModel):
    MessageType: str = 'AUTH'
    MessageId: int
    RoomUid: UUID
    Token: str
    LastEventId: int

class ActionMessage(BaseModel):
    MessageType: Literal['GAME-ACTION', 'VOTE']
    MessageId: int
    RoomUid: UUID
    UserUid: UUID
    ActionType: Literal['FOLD', 'CHECK', 'CALL', 'RAISE', 'OUTCOME'] | None = None
    Coef: Literal['X1_5', 'X2', 'ALL-IN'] | None = None
    VoteType: Literal['START', 'WAIT'] | None = None