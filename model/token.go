package model

import (
	"gitlab8.alx/msp2.0/msp-lib/database"
	"admin-service/structure"
	"time"
	"admin-service/utils"
)

const DELETE_TOKENS = "DELETE FROM " + utils.DB_SCHEME + ".tokens WHERE user_id=?"

func InvalidateOldTokens(userId int64) (int, error) {
	result, err := database.GetDBManager().Db.Exec(DELETE_TOKENS, userId)
	return result.RowsAffected(), err
}

func CreateNewToken(userId int64, tokenString string, expiredTime *time.Time) (*structure.Token, error) {
	token := structure.Token{UserId: userId, Token: tokenString}
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
