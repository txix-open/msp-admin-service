package model

import (
	"gitlab.alx/msp2.0/msp-lib/database"
	libStr "gitlab.alx/msp2.0/msp-lib/structure"
	"gitlab.alx/msp2.0/msp-lib/utils"
	"time"
)

const DELETE_TOKENS = "DELETE FROM " + utils.DB_SCHEME + ".tokens WHERE user_id=?"
const DELETE_TOKEN = "DELETE FROM " + utils.DB_SCHEME + ".tokens WHERE token=?"

func InvalidateOldTokens(userId int64) (int, error) {
	result, err := database.GetDBManager().Db.Exec(DELETE_TOKENS, userId)
	return result.RowsAffected(), err
}

func InvalidateOldToken(token string) (int, error) {
	result, err := database.GetDBManager().Db.Exec(DELETE_TOKEN, token)
	return result.RowsAffected(), err
}

func CreateNewToken(userId int64, tokenString string, expiredTime *time.Time) (*libStr.AdminToken, error) {
	token := libStr.AdminToken{UserId: userId, Token: tokenString}
	if expiredTime != nil {
		token.ExpiredAt = expiredTime
	}
	_, err := database.GetDBManager().Db.Model(&token).
		Returning("id").
		Returning("token").
		Returning("expired_at").
		Insert()
	return &token, err
}
