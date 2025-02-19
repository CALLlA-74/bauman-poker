package usecases

import (
	repo "bauman-poker/repo"
	tokenUtils "bauman-poker/utils"
)

type IdentityProvUsecase struct {
	repo        *repo.GormIdentityProvRepo
	tokenMaster *tokenUtils.TokenMaster
}

func NewIdentityProvUsecase(repo *repo.GormIdentityProvRepo,
	tokenMaster *tokenUtils.TokenMaster) *IdentityProvUsecase {

	return &IdentityProvUsecase{
		repo:        repo,
		tokenMaster: tokenMaster,
	}
}

func (uc IdentityProvUsecase) GetJWKs() *repo.JWKResponse {
	return &repo.JWKResponse{
		Keys: uc.repo.GetJWKs(),
	}
}

func (uc IdentityProvUsecase) SignUpUser(req *repo.SignUpReq) (*repo.AuthResp, *repo.ErrorResponse) {
	if isValid(req.Username) && isValid(req.Password) {
		user, err := uc.repo.CreateUser(req.Username, req.Password)
		if err != nil {
			if err.Error() == "400" {
				return nil, &repo.ErrorResponse{
					StatusCode: 400,
					Message:    "Username is not unique",
				}
			} else {
				return nil, &repo.ErrorResponse{
					StatusCode: 500,
					Message:    "Internal server error",
				}
			}
		}
		return uc.tokenMaster.GenerateSignedTokens(user, req.Scope)
	} else {
		return nil, &repo.ErrorResponse{
			StatusCode: 400,
			Message:    "Username or password is empty",
		}
	}
}

func (uc IdentityProvUsecase) AuthUser(req *repo.AuthReq) (*repo.AuthResp, *repo.ErrorResponse) {
	switch req.GrantType {
	case repo.PASSWORD:
		user, err := uc.repo.GetUserByNameAndPassw(req.Username, req.Password)
		if err != nil {
			if err.Error() == "401" {
				return nil, &repo.ErrorResponse{
					StatusCode: 401,
					Message:    "Unauthorized",
				}
			} else {
				return nil, &repo.ErrorResponse{
					StatusCode: 500,
					Message:    "Internal server error",
				}
			}
		}
		return uc.tokenMaster.GenerateSignedTokens(user, req.Scope)
	case repo.REFRESH_TOKEN:
		return uc.tokenMaster.UpdateTokens(req.RefreshToken, req.Scope)
	}
	return nil, &repo.ErrorResponse{
		StatusCode: 401,
		Message:    "Unauthorized",
	}
}

func (uc IdentityProvUsecase) Logout(refreshToken string) *repo.ErrorResponse {
	return uc.tokenMaster.RevokeToken(refreshToken)
}

func isValid(field string) bool {
	return len(field) > 0
}
