package controller

import "admin-service/structure"

func Auth() (*structure.Auth, error) {
	return &structure.Auth{"afasdaf", "asdfgsad", "x-afdasf-asd"}, nil
}
