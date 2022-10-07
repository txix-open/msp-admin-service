package tests

import (
	"context"
	"testing"

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
	GenerateToken(ctx context.Context, id int64) (string, string, error)
}

func TestUserTestSuite(t *testing.T) {
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

	cfg := conf.Remote{
		SecretKey: "secret",
		ExpireSec: 0,
	}
	locator := assembly.NewLocator(testInstance.Logger(), s.httpCli, s.db)
	handler := locator.Handler(cfg)

	server, apiCli := grpct.TestServer(testInstance, handler)
	s.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})

	tokenRep := repository.NewToken(s.db)
	s.tokenService = service.NewToken(tokenRep, 3600)
}

func (s *UserTestSuite) TestGetProfileHappyPath() {
	id := InsertUser(s.db, entity.CreateUser{
		RoleId:    1,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Password:  "password",
	})
	token, _, err := s.tokenService.GenerateToken(context.Background(), id)
	s.Require().NoError(err)

	response := domain.AdminUserShort{}
	err = s.grpcCli.Invoke("admin/user/get_profile").
		ReadJsonResponse(&response).
		AppendMetadata(domain.AdminAuthHeaderName, token).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.AdminUserShort{
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Role:      "admin",
	}
	s.Require().Equal(expected, response)
}

func (s *UserTestSuite) TestGetProfileUnauthorized() {
	id := InsertUser(s.db, entity.CreateUser{
		RoleId:    1,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Password:  "password",
	})
	token, _, err := s.tokenService.GenerateToken(context.Background(), id+1)
	s.Require().NoError(err)

	err = s.grpcCli.Invoke("admin/user/get_profile").
		AppendMetadata(domain.AdminAuthHeaderName, token).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())
}

func (s *UserTestSuite) TestGetProfileSudir() {
	id, err := InsertSudirUser(s.db, entity.SudirUser{
		RoleId:      1,
		SudirUserId: "sudirUser1",
		FirstName:   "name",
		LastName:    "surname",
		Email:       "a@a.ru",
	})
	s.Require().NoError(err)

	token, _, err := s.tokenService.GenerateToken(context.Background(), id)
	s.Require().NoError(err)

	response := domain.AdminUserShort{}
	err = s.grpcCli.Invoke("admin/user/get_profile").
		AppendMetadata(domain.AdminAuthHeaderName, token).
		ReadJsonResponse(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.AdminUserShort{
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Role:      "admin",
	}
	s.Require().Equal(expected, response)
}

func (s *UserTestSuite) TestGetUsers() {
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@a.ru"})
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "b@a.ru"})
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@b.ru"})
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@c.ru"})

	response := domain.UsersResponse{}
	err := s.grpcCli.Invoke("admin/user/get_users").
		JsonRequestBody(domain.UsersRequest{
			Ids:    []int64{3, 4, 5, 6},
			Offset: 1,
			Limit:  1,
			Email:  "a@",
		}).ReadJsonResponse(&response).Do(context.Background())
	s.Require().NoError(err)

	s.Require().Equal(1, len(response.Items))
	s.Require().Equal(int64(5), response.Items[0].Id)
}

func (s *UserTestSuite) TestCreateUserHappyPath() {
	preCount := 0
	s.db.Must().SelectRow(&preCount, "select count(*) from users")

	response := entity.User{}
	err := s.grpcCli.
		Invoke("admin/user/create_user").
		JsonRequestBody(domain.CreateUserRequest{
			RoleId:    1,
			FirstName: "name",
			LastName:  "surname",
			Email:     "a@a.ru",
			Password:  "password",
		}).
		ReadJsonResponse(&response).
		Do(context.Background())
	s.Require().NoError(err)

	postCount := 0
	s.db.Must().SelectRow(&postCount, "select count(*) from users")

	s.Require().Equal(preCount+1, postCount)
}

func (s *UserTestSuite) TestCreateUserAlreadyExist() {
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@a.ru"})

	err := s.grpcCli.
		Invoke("admin/user/create_user").
		JsonRequestBody(domain.CreateUserRequest{
			RoleId:    1,
			FirstName: "name",
			LastName:  "surname",
			Email:     "a@a.ru",
			Password:  "password",
		}).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.AlreadyExists, st.Code())
}

func (s *AuthTestSuite) TestUpdateUserHappyPath() {
	id := InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@a.ru", Password: "password"})
	req := domain.UpdateUserRequest{
		Id:        id,
		RoleId:    1,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Password:  "password",
	}
	response := domain.User{}
	err := s.grpcCli.
		Invoke("admin/user/update_user").
		JsonRequestBody(req).
		ReadJsonResponse(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.UpdateUserRequest{
		Id:        response.Id,
		RoleId:    response.RoleId,
		FirstName: response.FirstName,
		LastName:  response.LastName,
		Email:     response.Email,
		Password:  req.Password,
	}
	s.Require().Equal(expected, req)
}

func (s *AuthTestSuite) TestUpdateSudirUserHappyPath() {
	id, err := InsertSudirUser(s.db, entity.SudirUser{RoleId: 1, Email: "a@a.ru"})
	s.Require().NoError(err)
	req := domain.UpdateUserRequest{
		Id:        id,
		RoleId:    1,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
	}
	response := domain.User{}
	err = s.grpcCli.
		Invoke("admin/user/update_user").
		JsonRequestBody(req).
		ReadJsonResponse(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := domain.UpdateUserRequest{
		Id:        response.Id,
		RoleId:    response.RoleId,
		FirstName: response.FirstName,
		LastName:  response.LastName,
		Email:     response.Email,
	}
	s.Require().Equal(expected, req)
}

func (s *AuthTestSuite) TestUpdateSudirUserInvalidRequest() {
	id, err := InsertSudirUser(s.db, entity.SudirUser{RoleId: 1, Email: "a@a.ru"})
	s.Require().NoError(err)
	req := domain.UpdateUserRequest{
		Id:        id,
		RoleId:    1,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Password:  "password",
	}
	err = s.grpcCli.
		Invoke("admin/user/update_user").
		JsonRequestBody(req).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
}

func (s *AuthTestSuite) TestUpdateUserAlreadyExist() {
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@a.ru", Password: "password"})
	id := InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "b@b.ru", Password: "password"})
	req := domain.UpdateUserRequest{
		Id:        id,
		RoleId:    1,
		FirstName: "name",
		LastName:  "surname",
		Email:     "a@a.ru",
		Password:  "password",
	}
	err := s.grpcCli.
		Invoke("admin/user/update_user").
		JsonRequestBody(req).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.AlreadyExists, st.Code())
}

func (s *UserTestSuite) TestDeleteUsers() {
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@a.ru"})
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "b@a.ru"})
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@b.ru"})
	InsertUser(s.db, entity.CreateUser{RoleId: 1, Email: "a@c.ru"})

	response := domain.DeleteResponse{}
	err := s.grpcCli.Invoke("admin/user/delete_user").
		JsonRequestBody(domain.IdentitiesRequest{Ids: []int64{3, 4}}).ReadJsonResponse(&response).Do(context.Background())
	s.Require().NoError(err)

	s.Require().Equal(2, response.Deleted)
}
