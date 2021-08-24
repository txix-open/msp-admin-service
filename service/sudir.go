package service

import (
	"msp-admin-service/conf"
	"msp-admin-service/entity"
	"msp-admin-service/invoker"
)

func AuthSudir(cfg conf.SudirAuth, authCode string) (entity.AdminUser, error) {
	tokenResponse, err := invoker.GetSudirTokens(cfg, authCode)
	if err != nil {
		return entity.AdminUser{}, err
	} else if tokenResponse.SudirAuthError != nil {
		return entity.AdminUser{}, tokenResponse.SudirAuthError
	}

	user, err := invoker.GetSudirUser(cfg.Host, tokenResponse.AccessToken)
	if err != nil {
		return entity.AdminUser{}, err
	} else if user.SudirAuthError != nil {
		return entity.AdminUser{}, user.SudirAuthError
	}

	return entity.AdminUser{
		SudirUserId: user.UserId,
		FirstName:   user.Firstname,
		LastName:    user.Surname,
		Email:       user.Email,
		Password:    "",
	}, nil
}
