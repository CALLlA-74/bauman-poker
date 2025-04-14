package externalServices

import (
	"bauman-poker/config"
	"bauman-poker/schemas"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type IdentityExterService struct {
	baseUrlApiv1  string
	requestSender *RequestSender
}

func NewIdentityExterService() *IdentityExterService {
	context := NewBreakerContext(config.IdentityExterBaseUrl + config.HealthCheckHandler)
	return &IdentityExterService{
		baseUrlApiv1:  config.IdentityExterBaseUrl + config.IdentityGroupName,
		requestSender: NewRequestSender(context),
	}
}

func (ies IdentityExterService) GetJWKs() []schemas.JWKey {
	url := ies.baseUrlApiv1 + config.JWKsHandler
	newReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.WithError(err).Errorf("Error in IdentityExterService.GetJWKs. (err in making request). url: %s", url)
		return []schemas.JWKey{}
	}

	resp, err2 := ies.requestSender.SendRequest(newReq)
	if err2 != nil {
		log.Errorf("Error in IdentityExterService.GetJWKs. (err in sending request). url: %s", url)
		return []schemas.JWKey{}
	}

	respBody, err3 := ies.requestSender.ReadAll(resp)
	if err3 != nil {
		log.Errorf("Error in IdentityExterService.GetJWKs. (err in readAll resp)")
		return []schemas.JWKey{}
	}

	jwkResp := &schemas.JWKResponse{}
	if err := Unpack(respBody, jwkResp); err != nil {
		log.Errorf("Error in IdentityExterService.GetJWKs. (err in unmarshalling resp)")
		return []schemas.JWKey{}
	}
	log.Infof("List len: %d", len(*jwkResp.Keys))

	return *jwkResp.Keys
}

func (ies IdentityExterService) RegisterUser(req *http.Request) (*schemas.AuthResp, *schemas.ErrorResponse) {
	url := ies.baseUrlApiv1 + config.SignUpHandler
	newReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		log.WithError(err).Errorf("Error in IdentityExterService.RegisterUser. (err in making request)")
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	resp, err2 := ies.requestSender.SendRequest(newReq)
	if err2 != nil {
		log.Errorf("Error in IdentityExterService.RegisterUser. (err in sending request)")
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	respBody, err3 := ies.requestSender.ReadAll(resp)
	if err3 != nil {
		log.Errorf("Error in IdentityExterService.RegisterUser. (err in readAll resp)")
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	signUpResp := &schemas.AuthResp{}
	if err := Unpack(respBody, signUpResp); err != nil {
		errResp := &schemas.ErrorResponse{}
		if err := Unpack(respBody, errResp); err != nil {
			log.Errorf("Error in IdentityExterService.RegisterUser. (err in unmarshalling resp)")
			return nil, &schemas.ErrorResponse{
				StatusCode: 500,
				Message:    "Internal Server error",
			}
		}
		errResp.StatusCode = resp.StatusCode
		return nil, errResp
	}

	return signUpResp, nil
}

func (ies IdentityExterService) AuthUser(req *http.Request) (*schemas.AuthResp, *schemas.ErrorResponse) {
	url := ies.baseUrlApiv1 + config.AuthHandler
	newReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		log.WithError(err).Errorf("Error in IdentityExterService.AuthUser. (err in making request)")
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	resp, err2 := ies.requestSender.SendRequest(newReq)
	if err2 != nil {
		log.Errorf("Error in IdentityExterService.AuthUser. (err in sending request)")
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	respBody, err3 := ies.requestSender.ReadAll(resp)
	if err3 != nil {
		log.Errorf("Error in IdentityExterService.AuthUser. (err in readAll resp)")
		return nil, &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	authResp := &schemas.AuthResp{}
	if err := Unpack(respBody, authResp); err != nil {
		errResp := &schemas.ErrorResponse{}
		if err := Unpack(respBody, errResp); err != nil {
			log.Errorf("Error in IdentityExterService.AuthUser. (err in unmarshalling resp)")
			return nil, &schemas.ErrorResponse{
				StatusCode: 500,
				Message:    "Internal Server error",
			}
		}
		errResp.StatusCode = resp.StatusCode
		return nil, errResp
	}

	return authResp, nil
}

func (ies IdentityExterService) Logout(req *http.Request) *schemas.ErrorResponse {
	url := ies.baseUrlApiv1 + config.LogoutHandler
	newReq, err := http.NewRequest(req.Method, url, req.Body)
	newReq.Header = req.Header.Clone()
	if err != nil {
		log.WithError(err).Errorf("Error in IdentityExterService.Logout. (err in making request)")
		return &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	resp, err2 := ies.requestSender.SendRequest(newReq)
	if err2 != nil {
		log.Errorf("Error in IdentityExterService.Logout. (err in sending request)")
		return &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	respBody, err3 := ies.requestSender.ReadAll(resp)
	if err3 != nil {
		log.Errorf("Error in IdentityExterService.Logout. (err in readAll resp)")
		return &schemas.ErrorResponse{
			StatusCode: 500,
			Message:    "Internal Server error",
		}
	}

	if resp.StatusCode != http.StatusNoContent {
		errResp := &schemas.ErrorResponse{}
		if err := Unpack(respBody, errResp); err != nil {
			log.Errorf("Error in IdentityExterService.Logout. (err in unmarshalling resp)")
			return &schemas.ErrorResponse{
				StatusCode: 500,
				Message:    "Internal Server error",
			}
		}
		errResp.StatusCode = resp.StatusCode
		return errResp
	}
	return nil
}
