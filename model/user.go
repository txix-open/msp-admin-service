package model

import (
	"admin-service/structure"
	"gitlab8.alx/msp2.0/msp-lib/database"
	"github.com/go-pg/pg"
	"admin-service/utils"
)

const DELETE_USERS = "DELETE FROM " + utils.DB_SCHEME + ".users WHERE id IN (?)"

func GetUserByEmail(email string) (*structure.User, error) {
	var user structure.User
	err := database.GetDBManager().Db.Model(&user).
		Where("email = ?", email).
		First()
	if err != nil && err == pg.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func GetUserById(identity int64) (*structure.User, error) {
	var user structure.User
	err := database.GetDBManager().Db.Model(&user).
		Where("id = ?", identity).
		First()
	if err != nil && err == pg.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func GetUsers(usersRequest structure.UsersRequest) (*[]structure.User, error) {
	var users []structure.User
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

func CreateUser(user structure.User) (structure.User, error) {
	_, err := database.GetDBManager().Db.Model(&user).
		Returning("id").
		Returning("created_at").
		Returning("updated_at").
		Insert()
	return user, err
}

func UpdateUser(user structure.User) (structure.User, error) {
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
