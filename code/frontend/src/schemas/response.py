from pydantic import BaseModel
from uuid import UUID

class AuthenticationResponse(BaseModel):
    userUid: UUID
    refreshToken: str
    accessToken: str
    expiresIn: int
    scope: str

class ErrorResponse(BaseModel):
    message: str