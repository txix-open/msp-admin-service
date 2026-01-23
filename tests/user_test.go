package tests_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/txix-open/isp-kit/grpc/apierrors"
	"golang.org/x/crypto/bcrypt"

	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
	"msp-admin-service/repository"
	"msp-admin-service/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/grpct"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tokenService interface {
	GenerateToken(ctx context.Context, tokenRep service.TokenSaver, id int64) (string, string, error)
}

func TestUserTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &UserTestSuite{})
}

type UserTestSuite struct {
	suite.Suite

	test         *test.Test
	db           *dbt.TestDb
	grpcCli      *client.Client
	httpCli      *httpcli.Client
	tokenService tokenService
}

func (s *UserTestSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))
	s.httpCli = httpcli.New()

	remote := conf.Remote{
		ExpireSec: 0,
	}
	cfg := assembly.NewLocator(testInstance.Logger(), s.httpCli, s.db).
		Config(context.Background(), remote, time.Minute)

	server, apiCli := grpct.TestServer(testInstance, cfg.Handler)
	s.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})

	tokenRep := repository.NewToken(s.db)
	s.tokenService = service.NewToken(tokenRep, 3600)
}

func (s *UserTestSuite) TestGetProfileHappyPath() {
	id := InsertUser(s.db, entity.User{
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Password:  "password",
	})

	roleId := InsertRole(s.db, entity.Role{
		Name: "admin",
	})

	InsertUserRole(s.db, entity.UserRole{
		UserId: int(id),
		RoleId: int(roleId),
	})

	response := domain.AdminUserShort{}
	err := s.grpcCli.Invoke("admin/user/get_profile").
		JsonResponseBody(&response).
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id))).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.AdminUserShort{
		FirstName:   "name",
		LastName:    "surname",
		Email:       "a@a.ru",
		Role:        "admin",
		Roles:       []int{1},
		RoleNames:   []string{"admin"},
		Permissions: []string{},
	}
	s.Require().Equal(expected, response)
}

func (s *UserTestSuite) TestGetProfileNotFound() {
	id := InsertUser(s.db, entity.User{
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@b.ru",
		Password:  "password",
	})

	err := s.grpcCli.Invoke("admin/user/get_profile").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id+1))).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.NotFound, st.Code())
}

func (s *UserTestSuite) TestGetProfileSudir() {
	id := InsertSudirUser(s.db, entity.SudirUser{
		SudirUserId: "sudirUser1",
		FirstName:   "name",
		LastName:    "surname",
		Email:       "a@b.ru",
	})

	roleId := InsertRole(s.db, entity.Role{
		Name: "admin",
	})

	InsertUserRole(s.db, entity.UserRole{
		UserId: int(id),
		RoleId: int(roleId),
	})

	response := domain.AdminUserShort{}
	err := s.grpcCli.
		Invoke("admin/user/get_profile").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id))).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.AdminUserShort{
		FirstName:   "name",
		LastName:    "surname",
		Email:       "a@b.ru",
		Role:        "admin",
		Roles:       []int{1},
		RoleNames:   []string{"admin"},
		Permissions: []string{},
	}
	s.Require().Equal(expected, response)
}

func (s *UserTestSuite) TestGetUsers() {
	InsertUser(s.db, entity.User{Email: "a1@a.ru"})
	InsertUser(s.db, entity.User{Email: "b1@a.ru"})
	userId1 := InsertUser(s.db, entity.User{Email: "a1@b.ru"})
	userId2 := InsertUser(s.db, entity.User{Email: "a1@c.ru"})
	userId3 := InsertUser(s.db, entity.User{Email: "a1@d.ru"})

	roleId1 := InsertRole(s.db, entity.Role{Name: "test_role_1"})
	roleId2 := InsertRole(s.db, entity.Role{Name: "test_role_2"})

	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId1), UserId: int(userId1)})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId1), UserId: int(userId2)})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId1), UserId: int(userId3)})

	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId2), UserId: int(userId1)})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId2), UserId: int(userId2)})

	email := "a1@"

	response := domain.UsersResponse{}
	err := s.grpcCli.Invoke("admin/user/get_users").
		JsonRequestBody(domain.UsersPageRequest{
			LimitOffestParams: domain.LimitOffestParams{
				Limit:  5,
				Offset: 1,
			},
			Order: &domain.OrderParams{
				Field: "email",
				Type:  "asc",
			},
			Query: &domain.UserQuery{
				Email: &email,
				Roles: []int{int(roleId2)},
			},
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	roles := []int{int(roleId1), int(roleId2)}

	s.Require().Len(response.Items, 2)
	s.Require().EqualValues("a1@b.ru", response.Items[0].Email)
	s.Require().EqualValues(roles, response.Items[0].Roles)
	s.Require().EqualValues("a1@c.ru", response.Items[1].Email)
	s.Require().EqualValues(roles, response.Items[1].Roles)

	id := int(userId2)
	email = "test@mail.ru"
	err = s.grpcCli.Invoke("admin/user/get_users").
		JsonRequestBody(domain.UsersPageRequest{
			LimitOffestParams: domain.LimitOffestParams{
				Limit:  10,
				Offset: 0,
			},
			Order: &domain.OrderParams{
				Field: "id",
				Type:  "asc",
			},
			Query: &domain.UserQuery{
				Id:    &id,
				Email: &email,
			},
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)
	s.Require().Empty(response.Items)
}

func (s *UserTestSuite) TestGetUsersFilterByLastActiveAt() {
	userId1 := InsertUser(s.db, entity.User{Email: "test1@a.ru"})
	userId2 := InsertUser(s.db, entity.User{Email: "test2@a.ru"})
	userId3 := InsertUser(s.db, entity.User{Email: "test3@a.ru"})

	userTime1, err := time.Parse("2006-01-02T15:04:05Z", "2018-01-01T00:00:00Z")
	s.Require().NoError(err)

	userTime2, err := time.Parse("2006-01-02T15:04:05Z", "2020-01-01T00:00:00Z")
	s.Require().NoError(err)

	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token1",
		UserId:    userId1,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: userTime1.Add(1 * time.Hour),
		CreatedAt: userTime1,
		UpdatedAt: userTime1,
	})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token2",
		UserId:    userId2,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: userTime1.Add(25 * time.Hour),
		CreatedAt: userTime1.Add(24 * time.Hour),
		UpdatedAt: userTime1.Add(24 * time.Hour),
	})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token3",
		UserId:    userId3,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: userTime2.Add(1 * time.Hour),
		CreatedAt: userTime2,
		UpdatedAt: userTime2,
	})

	response := domain.UsersResponse{}
	err = s.grpcCli.Invoke("admin/user/get_users").
		JsonRequestBody(domain.UsersPageRequest{
			LimitOffestParams: domain.LimitOffestParams{
				Limit:  5,
				Offset: 0,
			},
			Order: &domain.OrderParams{
				Field: "lastSessionCreatedAt",
				Type:  "desc",
			},
			Query: &domain.UserQuery{
				LastSessionCreatedAt: &domain.DateFromToParams{
					From: userTime1.Add(-48 * time.Hour),
					To:   userTime1.Add(48 * time.Hour),
				},
			},
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 2)
	s.Require().EqualValues(userId2, response.Items[0].Id)
	s.Require().EqualValues(userId1, response.Items[1].Id)
}

func (s *UserTestSuite) TestGetUsersSortByName() {
	userIdViser1 := InsertUser(s.db, entity.User{Email: "a1@a.ru", FirstName: "admin_viser", LastName: "admin_viser"})
	userIdViser2 := InsertUser(s.db, entity.User{Email: "a1@b.ru", FirstName: "admin_viser", LastName: "admin_viser2"})
	userIdScripts := InsertUser(s.db, entity.User{Email: "a1@c.ru", FirstName: "admin_scripts", LastName: "admin_scripts"})
	userIdAL := InsertUser(s.db, entity.User{Email: "a1@d.ru", FirstName: "А", LastName: "Л"})

	response := domain.UsersResponse{}
	err := s.grpcCli.Invoke("admin/user/get_users").
		JsonRequestBody(domain.UsersPageRequest{
			LimitOffestParams: domain.LimitOffestParams{
				Limit:  10,
				Offset: 0,
			},
			Order: &domain.OrderParams{
				Field: "userId",
				Type:  "asc",
			},
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 6)
	s.Require().EqualValues(1, response.Items[0].Id)
	s.Require().EqualValues(userIdScripts, response.Items[1].Id)
	s.Require().EqualValues(2, response.Items[2].Id)
	s.Require().EqualValues(userIdViser1, response.Items[3].Id)
	s.Require().EqualValues(userIdViser2, response.Items[4].Id)
	s.Require().EqualValues(userIdAL, response.Items[5].Id)
}

func (s *UserTestSuite) TestCreateUserHappyPath() {
	admin := InsertUser(s.db, entity.User{Email: "admin@a.ru"})

	preCount := 0
	s.db.Must().SelectRow(&preCount, "select count(*) from users")

	response := entity.User{}
	err := s.grpcCli.
		Invoke("admin/user/create_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(admin))).
		JsonRequestBody(domain.CreateUserRequest{
			FirstName: "name",
			LastName:  "surname",
			Email:     "a2@a.ru",
			Password:  "password",
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	postCount := 0
	s.db.Must().SelectRow(&postCount, "select count(*) from users")

	s.Require().Equal(preCount+1, postCount)
	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *UserTestSuite) TestCreateUserWithSameEmailSudirHappyPath() {
	firstSudirId := "user_a"
	admin := InsertSudirUser(s.db, entity.SudirUser{SudirUserId: firstSudirId, Email: "a2@a.ru"})

	preCount := 0
	s.db.Must().SelectRow(&preCount, "select count(*) from users")

	response := entity.User{}
	err := s.grpcCli.
		Invoke("admin/user/create_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(admin))).
		JsonRequestBody(domain.CreateUserRequest{
			FirstName: "name",
			LastName:  "surname",
			Email:     "a2@a.ru",
			Password:  "password",
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	postCount := 0
	s.db.Must().SelectRow(&postCount, "select count(*) from users")

	s.Require().Equal(preCount+1, postCount)
	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *UserTestSuite) TestCreateUserAlreadyExist() {
	id := InsertUser(s.db, entity.User{Email: "exists@a.ru"})

	err := s.grpcCli.
		Invoke("admin/user/create_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id))).
		JsonRequestBody(domain.CreateUserRequest{
			FirstName: "name",
			LastName:  "surname",
			Email:     "exists@a.ru",
			Password:  "password",
		}).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.AlreadyExists, st.Code())
}

func (s *AuthTestSuite) TestUpdateUserHappyPath() {
	id := InsertUser(s.db, entity.User{Email: "update@a.ru", Password: "password"})
	req := domain.UpdateUserRequest{
		Id:        id,
		FirstName: "name",
		LastName:  "surname",
		Email:     "update@a.ru",
	}
	response := domain.User{}
	err := s.grpcCli.
		Invoke("admin/user/update_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id))).
		JsonRequestBody(req).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.UpdateUserRequest{
		Id:        response.Id,
		FirstName: response.FirstName,
		LastName:  response.LastName,
		Email:     response.Email,
	}
	s.Require().Equal(expected, req)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestUpdateSudirUserHappyPath() {
	id := InsertSudirUser(s.db, entity.SudirUser{SudirUserId: "123", Email: "sudir@a.ru"})
	req := domain.UpdateUserRequest{
		Id:        id,
		FirstName: "name",
		LastName:  "surname",
		Email:     "sudir@a.ru",
	}
	response := domain.User{}
	err := s.grpcCli.
		Invoke("admin/user/update_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id))).
		JsonRequestBody(req).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.UpdateUserRequest{
		Id:        response.Id,
		FirstName: response.FirstName,
		LastName:  response.LastName,
		Email:     response.Email,
	}
	s.Require().Equal(expected, req)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestUpdateUserAlreadyExist() {
	admin := InsertUser(s.db, entity.User{Email: "a_exists@a.ru", Password: "password"})
	id := InsertUser(s.db, entity.User{Email: "b_exists@b.ru", Password: "password"})
	req := domain.UpdateUserRequest{
		Id:        id,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a_exists@a.ru",
	}
	err := s.grpcCli.
		Invoke("admin/user/update_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(admin))).
		JsonRequestBody(req).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.AlreadyExists, st.Code())
}

func (s *AuthTestSuite) TestUpdateUserWithSudirAlreadyExist() {
	firstSudirId := "user_a"
	secondSudirId := "user_b"

	admin := InsertSudirUser(s.db, entity.SudirUser{SudirUserId: firstSudirId, Email: "a_exists@a.ru"})
	id := InsertSudirUser(s.db, entity.SudirUser{SudirUserId: secondSudirId, Email: "b_exists@b.ru"})
	req := domain.UpdateUserRequest{
		Id:        id,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a_exists@a.ru",
	}
	err := s.grpcCli.
		Invoke("admin/user/update_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(admin))).
		JsonRequestBody(req).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.AlreadyExists, st.Code())
}

func (s *UserTestSuite) TestDeleteUsers() {
	admin := InsertUser(s.db, entity.User{Email: "a_del@a.ru"})
	InsertUser(s.db, entity.User{Email: "b_del@a.ru"})
	InsertUser(s.db, entity.User{Email: "a_del@b.ru"})
	InsertUser(s.db, entity.User{Email: "a_del@c.ru"})

	response := domain.DeleteResponse{}
	err := s.grpcCli.
		Invoke("admin/user/delete_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(admin))).
		JsonRequestBody(domain.IdentitiesRequest{Ids: []int64{3, 4}}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	s.Require().Equal(2, response.Deleted)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *UserTestSuite) TestBlockUser() {
	id := InsertUser(s.db, entity.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a_block@a.ru",
		Password:  "password",
		Blocked:   false,
	})
	token := uuid.New().String()
	InsertTokenEntity(s.db, entity.Token{
		Token:     token,
		UserId:    id,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().Add(5 * time.Second),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	err := s.grpcCli.Invoke("admin/user/block_user").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(id))).
		JsonRequestBody(domain.IdRequest{UserId: int(id)}).
		Do(context.Background())
	s.Require().NoError(err)

	user, err := repository.NewUser(s.db).GetUserById(context.Background(), id)
	s.Require().NoError(err)
	s.Require().True(user.Blocked)

	t, err := repository.NewToken(s.db).Get(context.Background(), token)
	s.Require().NoError(err)
	s.EqualValues(entity.TokenStatusRevoked, t.Status)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *UserTestSuite) TestChangePasswordUser() {
	// insert user with old password
	adminId := InsertUser(s.db, entity.User{Email: "a_del@a.ru", Password: "password"})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "token-841297641213",
		UserId:    adminId,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Time{},
		ExpiredAt: time.Time{},
	})

	// check for err when invalid data
	invalidReq := domain.ChangePasswordRequest{OldPassword: "invalid", NewPassword: "new_password"}
	err := s.grpcCli.Invoke("admin/user/change_password").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(adminId))).
		JsonRequestBody(invalidReq).
		Do(context.Background())
	s.Require().Error(err)
	err = apierrors.FromError(err)
	s.Require().Contains(err.Error(), strconv.Itoa(domain.ErrCodeInvalidPassword))

	// change password
	changePswReq := domain.ChangePasswordRequest{OldPassword: "password", NewPassword: "new_password"}
	err = s.grpcCli.Invoke("admin/user/change_password").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(adminId))).
		JsonRequestBody(changePswReq).
		Do(context.Background())

	s.Require().NoError(err)

	var newPassword string
	s.db.Must().SelectRow(&newPassword, "select password from users where id = $1", adminId)

	notEqualErr := bcrypt.CompareHashAndPassword([]byte(newPassword), []byte(changePswReq.NewPassword))
	s.Require().NoError(notEqualErr)

	tokenInfo := SelectTokenEntityByToken(s.db, "token-841297641213")
	s.Require().Equal(entity.TokenStatusRevoked, tokenInfo.Status)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}
