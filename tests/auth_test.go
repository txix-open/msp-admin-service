package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/json"
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
)

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, &AuthTestSuite{})
}

type AuthTestSuite struct {
	suite.Suite
	test    *test.Test
	db      *dbt.TestDb
	grpcCli *client.Client
	httpCli *httpcli.Client
}

func (s *AuthTestSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigration("../migrations"))
	s.httpCli = httpcli.New()

	mocksrv, host := s.initMockSudir()

	cfg := conf.Remote{
		SudirAuth: &conf.SudirAuth{
			ClientId:     "admin",
			ClientSecret: "admin",
			Host:         host,
			RedirectURI:  "http://localhost",
		},
		SecretKey: "admin",
		ExpireSec: 0,
	}

	locator := assembly.NewLocator(testInstance.Logger(), s.httpCli, s.db)
	handler := locator.Handler(cfg)

	server, apiCli := grpct.TestServer(testInstance, handler)
	s.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
		mocksrv.Close()
	})
}

type customClaims struct {
	Id int64
	jwt.StandardClaims
}

func (s *AuthTestSuite) TestLoginHappyPath() {
	id := InsertUser(s.db, entity.CreateUser{
		RoleId:    1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
	})

	response := domain.LoginResponse{}
	err := s.grpcCli.Invoke("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a@a.ru",
			Password: "password",
		}).
		ReadJsonResponse(&response).
		Do(context.Background())
	s.Require().NoError(err)
	claims := customClaims{Id: id}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("admin"))
	s.Require().NoError(err)
	s.Require().Equal(token, response.Token)
}

func (s *AuthTestSuite) TestLoginNotFound() {
	err := s.grpcCli.Invoke("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a1@a.ru",
			Password: "password",
		}).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())
}

func (s *AuthTestSuite) TestLoginWrongPassword() {
	InsertUser(s.db, entity.CreateUser{
		RoleId:    1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
	})

	err := s.grpcCli.Invoke("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a@a.ru",
			Password: "WrongPassword",
		}).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())
}

func (s *AuthTestSuite) TestSudirLoginHappyPath() {
	response := domain.LoginResponse{}

	err := s.grpcCli.Invoke("admin/auth/login_with_sudir").
		JsonRequestBody(domain.LoginSudirRequest{
			AuthCode: "code",
		}).
		ReadJsonResponse(&response).
		Do(context.Background())
	s.Require().NoError(err)
	user := entity.User{}
	s.db.Must().SelectRow(&user, "select id, role_id, email from users where sudir_user_id = $1", "sudirUser1")
	s.Require().Equal(1, user.RoleId)
	s.Require().Equal("sudir@email.ru", user.Email)

	claims := customClaims{Id: user.Id}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("admin"))
	s.Require().NoError(err)
	s.Require().Equal(token, response.Token)
}

func (s *AuthTestSuite) initMockSudir() (*httptest.Server, string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blitz/oauth/te", func(writer http.ResponseWriter, request *http.Request) {
		res := entity.SudirTokenResponse{
			SudirAuthError: nil,
			IdToken:        "1",
			AccessToken:    "token",
		}
		data, err := json.Marshal(res)
		s.Require().NoError(err)
		_, err = writer.Write(data)
		s.Require().NoError(err)
	})
	mux.HandleFunc("/blitz/oauth/me", func(writer http.ResponseWriter, request *http.Request) {
		res := entity.SudirUserResponse{
			SudirAuthError: nil,
			Email:          "sudir@email.ru",
			Groups:         []string{"DIT-KKD-Admins"},
			Sub:            "sudirUser1",
			GivenName:      "name",
			FamilyName:     "surname",
		}
		data, err := json.Marshal(res)
		s.Require().NoError(err)
		_, err = writer.Write(data)
		s.Require().NoError(err)
	})
	srv := httptest.NewServer(mux)
	return srv, srv.URL
}
