package controller

import (
	"errors"
	"fmt"

	"github.com/integration-system/isp-lib/v2/config"
	"github.com/integration-system/isp-lib/v2/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
	"msp-admin-service/invoker"
	"msp-admin-service/model"
	"msp-admin-service/service"
	"msp-admin-service/structure"
)

const adminAuthHeaderName = "x-auth-admin"

const (
	ValidationError = "Validation errors"
)

// @Tags auth
// @Summary Выход из авторизованной сессии
// @Description Выход из авторизованной сессии администрирования
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200
// @Failure 400 {object} structure.GrpcError "Невалидный токен"
// @Failure 500 {object} structure.GrpcError
// @Router /auth/logout [POST]
func Logout(_ metadata.MD) error {
	return nil
}

// @Tags user
// @Summary Получить профиль
// @Description Получить данные профиля
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {object} structure.AdminUserShort
// @Failure 400 {object} structure.GrpcError "Невалидный токен"
// @Failure 401 {object} structure.GrpcError "Токен не соответствует ни одному пользователю"
// @Failure 500 {object} structure.GrpcError
// @Router /user/get_profile [POST]
func GetProfile(metadata metadata.MD) (*structure.AdminUserShort, error) {
	token := metadata.Get(adminAuthHeaderName)
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

// @Tags auth
// @Summary Авторизация по логину и паролю
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body structure.AuthRequest true "Тело запроса"
// @Success 200 {object} structure.Auth
// @Failure 400 {object} structure.GrpcError
// @Failure 401 {object} structure.GrpcError "Данные для авторизации не верны"
// @Failure 404 {object} structure.GrpcError "Пользователь не найден"
// @Failure 500 {object} structure.GrpcError
// @Router /auth/login [POST]
func Login(authRequest structure.AuthRequest) (*structure.Auth, error) {
	user, err := model.GetUserByEmail(authRequest.Email)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	} else if user.SudirUserId != "" {
		return nil, status.Error(codes.InvalidArgument, "User is authorized only with SUDIR")
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

// @Tags auth
// @Summary Авторизация по авторизационному коду от СУДИР
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body structure.SudirAuthRequest true "Тело запроса"
// @Success 200 {object} structure.Auth
// @Failure 401 {object} structure.GrpcError "Некорректный код для авторизации"
// @Failure 412 {object} structure.GrpcError "Авторизация СУДИР не настроена на сервере"
// @Failure 500 {object} structure.GrpcError
// @Router /auth/login_with_sudir [POST]
func LoginWithSudir(request structure.SudirAuthRequest) (*structure.Auth, error) {
	remoteConfig := config.GetRemote().(*conf.RemoteConfig)
	if remoteConfig.SudirAuth == nil {
		return nil, status.Error(codes.FailedPrecondition, "SUDIR authorization is not configured on the server")
	}

	sudirUser, err := service.AuthSudir(*remoteConfig.SudirAuth, request.AuthCode)
	var authErr *invoker.SudirAuthError
	if errors.As(err, &authErr) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	} else if err != nil {
		return nil, err
	}

	if sudirUser.SudirUserId == "" || sudirUser.Email == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing sudirUserId '%s' or email '%s'", sudirUser.SudirUserId, sudirUser.Email)
	}

	existedUser, err := model.GetUserBySudirUserId(sudirUser.SudirUserId)
	if err != nil {
		return nil, err
	}
	if existedUser != nil {
		sudirUser = *existedUser
	} else {
		sudirUser, err = model.CreateUser(sudirUser)
		if err != nil {
			return nil, err
		}
	}

	tokenString, expired, err := service.GenerateToken(sudirUser.Id)
	if err != nil {
		return nil, err
	}
	return &structure.Auth{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: adminAuthHeaderName,
	}, nil
}

// @Tags user
// @Summary Список пользователей
// @Description Получить список пользователей
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body structure.UsersRequest true "Тело запроса"
// @Success 200 {object} structure.UsersResponse
// @Failure 400 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /user/get_users [POST]
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

// @Tags user
// @Summary Создать/Обновить пользователя
// @Description Создать пользователя или обновить данные существующего
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body entity.AdminUser true "Тело запроса"
// @Success 200 {object} entity.AdminUser
// @Failure 400 {object} structure.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} structure.GrpcError
// @Router /user/create_update_user [POST]
func CreateUpdateUser(user entity.AdminUser) (*entity.AdminUser, error) {
	if user.Id == 0 {
		if user.Password == "" {
			validationErrors := map[string]string{"password": "Required"}
			return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument,
				ValidationError, validationErrors)
		} else if user.SudirUserId != "" {
			validationErrors := map[string]string{"sudirUserId": "Could not create user with sudirUserId"}
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
		if err != nil {
			return nil, err
		}
	} else {
		userExists, err := model.GetUserById(user.Id)
		if err != nil {
			return nil, err
		}
		if userExists.Id == 0 {
			validationErrors := map[string]string{
				"id": fmt.Sprintf("User with id: %d not found", user.Id),
			}
			return nil, utils.CreateValidationErrorDetails(codes.NotFound,
				ValidationError, validationErrors)
		} else if userExists.SudirUserId != user.SudirUserId {
			validationErrors := map[string]string{"sudirUserId": "Could not change sudirUserId of an existing user"}
			return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument,
				ValidationError, validationErrors)
		} else if user.Password != "" && user.SudirUserId != "" {
			validationErrors := map[string]string{"password": "Could not set password of a user logged in with SUDIR"}
			return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument,
				ValidationError, validationErrors)
		}

		if user.Password != "" {
			err = cryptPassword(&user)
			if err != nil {
				return nil, err
			}
		}

		user, err = model.UpdateUser(user)
		if err != nil {
			return nil, err
		}

		user.CreatedAt = userExists.CreatedAt
		user.UpdatedAt = userExists.UpdatedAt
	}

	user.Password = ""

	return &user, nil
}

// @Tags user
// @Summary Удалить пользователя
// @Description Удалить существующего пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body structure.IdentitiesRequest true "Тело запроса"
// @Success 200 {object} structure.DeleteResponse
// @Failure 400 {object} structure.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} structure.GrpcError
// @Router /user/delete_user [POST]
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
