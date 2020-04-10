package controller

import (
	"fmt"

	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"msp-admin-service/entity"
	"msp-admin-service/model"
	"msp-admin-service/service"
	"msp-admin-service/structure"
)

const adminAuthHeaderName = "x-auth-admin"

const (
	ServiceError    = "Service is not available now, please try later"
	ValidationError = "Validation errors"
)

func Logout(metadata metadata.MD) error {
	token := metadata.Get(adminAuthHeaderName)

	if len(token) == 0 || token[0] == "" {
		log.Errorf(0, "Admin AUTH header: %s, not found, received: %v", adminAuthHeaderName, metadata)
		return status.Error(codes.InvalidArgument, ServiceError)
	}
	return nil
}

func GetProfile(metadata metadata.MD) (*structure.AdminUserShort, error) {
	token := metadata.Get(adminAuthHeaderName)

	if len(token) == 0 || token[0] == "" {
		log.Errorf(0, "Admin AUTH header: %s, not found, received: %v", adminAuthHeaderName, metadata)
		return nil, status.Error(codes.InvalidArgument, ServiceError)
	}

	userId, err := service.GetUserId(token[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid auth token")
	}

	user, err := model.GetUserById(userId)
	if err != nil {
		return nil, err
	}
	return &structure.AdminUserShort{
		Image:     user.Image,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
	}, nil
}

func Login(authRequest structure.AuthRequest) (*structure.Auth, error) {
	user, err := model.GetUserByEmail(authRequest.Email)
	if user == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authRequest.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Email or password is incorrect")
	}

	tokenString, expired, err := service.GenerateToken(user.Id)
	if err != nil {
		return nil, err
	}
	return &structure.Auth{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: adminAuthHeaderName,
	}, nil
}

func GetUsers(identities structure.UsersRequest) (*structure.UsersResponse, error) {
	users, err := model.GetUsers(identities)
	if err != nil {
		return nil, err
	}
	for i := range *users {
		(*users)[i].Password = ""
	}
	return &structure.UsersResponse{Items: users}, err
}

func CreateUpdateUser(user entity.AdminUser) (*entity.AdminUser, error) {
	var err error
	if user.Id == 0 {
		if user.Password == "" {
			validationErrors := map[string]string{"password": "Requered"}
			return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument,
				ValidationError, validationErrors)
		}

		userExists, err := model.GetUserByEmail(user.Email)
		if err != nil {
			return nil, err
		}
		if userExists != nil && userExists.Id != 0 {
			validationErrors := map[string]string{
				"email": fmt.Sprintf("User with email: %s already exists", user.Email),
			}
			return nil, utils.CreateValidationErrorDetails(codes.AlreadyExists,
				ValidationError, validationErrors)
		}

		err = cryptPassword(&user)
		if err != nil {
			return nil, err
		}

		user, err = model.CreateUser(user)
	} else {
		var userExists *entity.AdminUser
		userExists, err = model.GetUserById(user.Id)
		if err != nil {
			return nil, err
		}
		if userExists.Id == 0 {
			validationErrors := map[string]string{
				"id": fmt.Sprintf("User with id: %d not found", user.Id),
			}
			return nil, utils.CreateValidationErrorDetails(codes.NotFound,
				ValidationError, validationErrors)
		}

		if user.Password != "" {
			err = cryptPassword(&user)
			if err != nil {
				return nil, err
			}
		}

		user, err = model.UpdateUser(user)
		user.CreatedAt = userExists.CreatedAt
		user.UpdatedAt = userExists.UpdatedAt
	}
	if err != nil {
		return nil, err
	}

	user.Password = ""

	return &user, err
}

func DeleteUser(identities structure.IdentitiesRequest) (*structure.DeleteResponse, error) {
	if len(identities.Ids) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument,
			ValidationError, validationErrors)
	}
	count, err := model.DeleteUser(identities)
	if err != nil {
		return nil, err
	}
	return &structure.DeleteResponse{Deleted: count}, err
}

func cryptPassword(user *entity.AdminUser) error {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return err
	}
	user.Password = string(passwordBytes)
	return nil
}
