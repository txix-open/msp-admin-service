package tests_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/dbt"
	"github.com/integration-system/isp-kit/test/grpct"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
	"msp-admin-service/repository"
	"msp-admin-service/service"
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
	s.db = dbt.New(testInstance, dbx.WithMigration("../migrations"))
	s.httpCli = httpcli.New()

	remote := conf.Remote{
		ExpireSec: 0,
	}
	cfg := assembly.NewLocator(testInstance.Logger(), s.httpCli, s.db).
		Config(context.Background(), emptyLdap, remote)

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

	roleId, err := InsertRole(s.db, entity.Role{
		Name: "admin",
	})
	s.Require().NoError(err)

	err = InsertUserRole(s.db, entity.UserRole{
		UserId: int(id),
		RoleId: int(roleId),
	})
	s.Require().NoError(err)

	response := domain.AdminUserShort{}
	err = s.grpcCli.Invoke("admin/user/get_profile").
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
	id, err := InsertSudirUser(s.db, entity.SudirUser{
		SudirUserId: "sudirUser1",
		FirstName:   "name",
		LastName:    "surname",
		Email:       "a@b.ru",
	})
	s.Require().NoError(err)

	roleId, err := InsertRole(s.db, entity.Role{
		Name: "admin",
	})
	s.Require().NoError(err)

	err = InsertUserRole(s.db, entity.UserRole{
		UserId: int(id),
		RoleId: int(roleId),
	})
	s.Require().NoError(err)

	response := domain.AdminUserShort{}
	err = s.grpcCli.
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
		Permissions: []string{},
	}
	s.Require().Equal(expected, response)
}

func (s *UserTestSuite) TestGetUsers() {
	InsertUser(s.db, entity.User{Email: "a1@a.ru"})
	InsertUser(s.db, entity.User{Email: "b1@a.ru"})
	InsertUser(s.db, entity.User{Email: "a1@b.ru"})
	InsertUser(s.db, entity.User{Email: "a1@c.ru"})

	response := domain.UsersResponse{}
	err := s.grpcCli.Invoke("admin/user/get_users").
		JsonRequestBody(domain.UsersRequest{
			Ids:    []int64{3, 4, 5, 6},
			Offset: 1,
			Limit:  1,
			Email:  "a1@",
		}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)

	s.Require().Equal(1, len(response.Items))
	s.Require().Equal(int64(5), response.Items[0].Id)
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
}

func (s *AuthTestSuite) TestUpdateSudirUserHappyPath() {
	id, err := InsertSudirUser(s.db, entity.SudirUser{Email: "sudir@a.ru"})
	s.Require().NoError(err)
	req := domain.UpdateUserRequest{
		Id:        id,
		FirstName: "name",
		LastName:  "surname",
		Email:     "sudir@a.ru",
	}
	response := domain.User{}
	err = s.grpcCli.
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

	t, err := repository.NewToken(s.db).GetEntity(context.Background(), token)
	s.Require().NoError(err)
	s.EqualValues(entity.TokenStatusRevoked, t.Status)
}
