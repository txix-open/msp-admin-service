package model

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
	"msp-admin-service/entity"
)

type RoleRepository struct {
	DB     orm.DB
	client *database.RxDbClient
}

func (r *RoleRepository) GetRoleById(id int) (*entity.Role, error) {
	role := new(entity.Role)
	err := r.getDb().Model(role).Where("id = ?", id).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) GetRoleByName(name string) (*entity.Role, error) {
	role := new(entity.Role)
	err := r.getDb().Model(role).Where("name = ?", name).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}
	return r.client.Unsafe()
}
