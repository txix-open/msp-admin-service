package controller

import (
	"admin-service/structure"
	"admin-service/model"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"gitlab8.alx/msp2.0/msp-lib/validate"
	libUtils "gitlab8.alx/msp2.0/msp-lib/utils"
	libStr "gitlab8.alx/msp2.0/msp-lib/structure"
	"gitlab8.alx/msp2.0/msp-lib/token-gen"
	"golang.org/x/crypto/bcrypt"
	"gitlab8.alx/msp2.0/msp-lib/utils"
	"fmt"
	"google.golang.org/grpc/metadata"
	"gitlab8.alx/msp2.0/msp-lib/logger"
)

func Logout(metadata metadata.MD) error {
	token := metadata.Get(utils.ADMIN_AUTH_HEADER_NAME)
	
	if len(token) == 0 || token[0] == "" {
		logger.Errorf("Admin AUTH header: %s, not found, received: %v", utils.ADMIN_AUTH_HEADER_NAME, metadata)
		st := status.New(codes.InvalidArgument, utils.ServiceError)
		return st.Err()
	}
	model.InvalidateOldToken(token[0])
	return nil
}

func GetProfile(metadata metadata.MD) (*structure.AdminUserShort, error) {
	token := metadata.Get(utils.ADMIN_AUTH_HEADER_NAME)
	
	if len(token) == 0 || token[0] == "" {
		logger.Errorf("Admin AUTH header: %s, not found, received: %v", utils.ADMIN_AUTH_HEADER_NAME, metadata)
		st := status.New(codes.InvalidArgument, utils.ServiceError)
		return nil, st.Err()
	}
	user, err := model.GetUserByToken(token[0])
	if err != nil {
		return nil, validate.CreateUnknownError(err)
	}
	return user, nil
}

func Login(authRequest structure.AuthRequest) (*structure.Auth, error) {
	user, err := model.GetUserByEmail(authRequest.Email)
	if user == nil {
		return nil, status.New(codes.NotFound, "User not found").Err()
	}
	if err != nil {
		return nil, validate.CreateUnknownError(err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authRequest.Password))
	if err != nil {
		return nil, status.New(codes.Unauthenticated, "Email or password is incorrect").Err()
	}
	
	_, err = model.InvalidateOldTokens(user.Id)
	if err != nil {
		return nil, validate.CreateUnknownError(err)
	}
	token, err := model.CreateNewToken(user.Id, tg.Default.NextDefault(), nil)
	if err != nil {
		return nil, validate.CreateUnknownError(err)
	}
	
	return &structure.Auth{
		Token:      token.Token,
		Expired:    token.ExpiredAt.Format(libUtils.FullDateFormat),
		HeaderName: utils.ADMIN_AUTH_HEADER_NAME,
	}, nil
}

func GetUsers(identities structure.UsersRequest) (*structure.UsersResponse, error) {
	users, err := model.GetUsers(identities)
	if err != nil {
		return nil, validate.CreateUnknownError(err)
	}
	for i := range *users {
		(*users)[i].Password = ""
	}
	return &structure.UsersResponse{Items: users}, err
}

func CreateUpdateUser(user libStr.AdminUser) (*libStr.AdminUser, error) {
	var err error
	if user.Id == 0 {
		if user.Password == "" {
			validationErrors := map[string]string{"password": "Requered"}
			return nil, libUtils.CreateValidationErrorDetails(codes.InvalidArgument,
				libUtils.ValidationError, validationErrors)
		}
		
		userExists, err := model.GetUserByEmail(user.Email)
		if err != nil {
			return nil, validate.CreateUnknownError(err)
		}
		if userExists != nil && userExists.Id != 0 {
			validationErrors := map[string]string{
				"email": fmt.Sprintf("User with email: %s already exists", user.Email),
			}
			return nil, libUtils.CreateValidationErrorDetails(codes.AlreadyExists,
				libUtils.ValidationError, validationErrors)
		}
		
		err = cryptPassword(&user)
		if err != nil {
			return nil, err
		}
		
		user, err = model.CreateUser(user)
	} else {
		userExists, err := model.GetUserById(user.Id)
		if err != nil {
			return nil, validate.CreateUnknownError(err)
		}
		if userExists.Id == 0 {
			validationErrors := map[string]string{
				"id": fmt.Sprintf("User with id: %d not found", user.Id),
			}
			return nil, libUtils.CreateValidationErrorDetails(codes.NotFound,
				libUtils.ValidationError, validationErrors)
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
		return nil, validate.CreateUnknownError(err)
	}
	
	user.Password = ""
	
	return &user, err
}

func DeleteUser(identities structure.IdentitiesRequest) (*structure.DeleteResponse, error) {
	if len(identities.Ids) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, libUtils.CreateValidationErrorDetails(codes.InvalidArgument,
			libUtils.ValidationError, validationErrors)
	}
	count, err := model.DeleteUser(identities)
	if err != nil {
		return nil, validate.CreateUnknownError(err)
	}
	return &structure.DeleteResponse{Deleted: count}, err
}

func cryptPassword(user *libStr.AdminUser) error {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return validate.CreateUnknownError(err)
	}
	user.Password = string(passwordBytes)
	return nil
}
