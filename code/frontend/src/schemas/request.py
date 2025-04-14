from pydantic import BaseModel
from typing import Literal

class SignUpRequest(BaseModel):
    scope: Literal['OPENID']
    username: str
    password: str

class AuthenticationRequest(BaseModel):
    scope: Literal['OPENID']
    grantType: Literal['PASSWORD', 'REFRESH-TOKEN']
    username: str | None = None
    password: str | None = None
    refreshToken: str | None = None