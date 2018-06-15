package model

import (
	libStr "gitlab8.alx/msp2.0/msp-lib/structure"
	"gitlab8.alx/msp2.0/msp-lib/database"
	"github.com/go-pg/pg"
	"gitlab8.alx/msp2.0/msp-lib/utils"
	"admin-service/structure"
)

const DELETE_USERS = "DELETE FROM " + utils.DB_SCHEME + ".users WHERE id IN (?)"

func GetUserByToken(token string) (*structure.AdminUserShort, error) {
	var user structure.AdminUserShort
	_, err := database.GetDBManager().Db.Model(&user).
		Query(&user, `SELECT
			u.image,
				u.email,
				u.phone,
				u.first_name,
				u.last_name
			FROM
				` + utils.DB_SCHEME + `.tokens
			t LEFT JOIN ` + utils.DB_SCHEME + `.users u ON t.user_id = u.ID
			WHERE
				token = ?`, token)
	if err != nil && err == pg.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func GetUserByEmail(email string) (*libStr.AdminUser, error) {
	var user libStr.AdminUser
	err := database.GetDBManager().Db.Model(&user).
		Where("email = ?", email).
		First()
	if err != nil && err == pg.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func GetUserById(identity int64) (*libStr.AdminUser, error) {
	var user libStr.AdminUser
	err := database.GetDBManager().Db.Model(&user).
		Where("id = ?", identity).
		First()
	if err != nil && err == pg.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func GetUsers(usersRequest structure.UsersRequest) (*[]libStr.AdminUser, error) {
	var users []libStr.AdminUser
	query := database.GetDBManager().Db.Model(&users)
	if len(usersRequest.Ids) > 0 {
		query.Where("id IN (?)", pg.In(usersRequest.Ids))
	}
	if usersRequest.Email != "" {
		query.Where("email LIKE ?", "%"+usersRequest.Email+"%")
	}
	if usersRequest.Phone != "" {
		query.Where("phone LIKE ?", "%"+usersRequest.Phone+"%")
	}
	err := query.
		
		Order("created_at DESC").
		Limit(usersRequest.Limit).
		Offset(usersRequest.Offset).
		Select()
	if err != nil && err == pg.ErrNoRows {
		return nil, nil
	}
	return &users, err
}

func CreateUser(user libStr.AdminUser) (libStr.AdminUser, error) {
	_, err := database.GetDBManager().Db.Model(&user).
		Returning("id").
		Returning("created_at").
		Returning("updated_at").
		Insert()
	return user, err
}

func UpdateUser(user libStr.AdminUser) (libStr.AdminUser, error) {
	_, err := database.GetDBManager().Db.Model(&user).
		WherePK().
		Returning("id").
		Returning("created_at").
		Returning("updated_at").
		Update()
	return user, err
}

func DeleteUser(identities structure.IdentitiesRequest) (int, error) {
	result, err := database.GetDBManager().Db.
		Exec(DELETE_USERS, pg.In(identities.Ids))
	return result.RowsAffected(), err
}
