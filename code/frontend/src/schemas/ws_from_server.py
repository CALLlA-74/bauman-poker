from pydantic import BaseModel
from uuid import UUID
from typing import Literal, List

from event_descriptors import PlayerActionEvent, PrepareEvent, GameEvent

class ResponseMessage(BaseModel):
    messageType: str = 'ACK'
    ackMessageId: int
    statusCode: 200 | 401 | 400

class EventMessage(BaseModel):
    MessageType: str = 'EVENT'
    MessageId: int
    EventDescriptor: PlayerActionEvent | PrepareEvent | GameEvent