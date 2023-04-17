package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"

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
		ExpireSec: 3600,
		AntiBruteforce: conf.AntiBruteforce{
			MaxInFlightLoginRequests: 3,
			DelayLoginRequestInSec:   1,
		},
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

func (s *AuthTestSuite) TestLoginHappyPath() {
	id := InsertUser(s.db, entity.User{
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

	tokenInfo := SelectTokenEntityByToken(s.db, response.Token)
	s.Require().Equal(tokenInfo.UserId, id)
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

func (s *AuthTestSuite) TestBlockedUser() {
	InsertUser(s.db, entity.User{
		RoleId:    1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
		Blocked:   true,
	})

	err := s.grpcCli.Invoke("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a@a.ru",
			Password: "password",
		}).
		Do(context.Background())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())
}

func (s *AuthTestSuite) TestLoginWrongPassword() {
	InsertUser(s.db, entity.User{
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
	s.Require().Equal(3, user.RoleId)
	s.Require().Equal("sudir@email.ru", user.Email)

	tokenInfo := SelectTokenEntityByToken(s.db, response.Token)
	s.Require().Equal(tokenInfo.UserId, user.Id)
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

func (s *AuthTestSuite) Test_Logout_HappyPath() {
	userId := InsertUser(s.db, entity.User{RoleId: 1})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "token-841297641213",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Time{},
		ExpiredAt: time.Time{},
	})
	err := s.grpcCli.Invoke("admin/auth/logout").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(userId))).
		Do(context.Background())
	s.Require().NoError(err)

	tokenInfo := SelectTokenEntityByToken(s.db, "token-841297641213")
	s.Require().Equal(entity.TokenStatusRevoked, tokenInfo.Status)
}

func (s *AuthTestSuite) Test_Logout_NotFound() {
	err := s.grpcCli.Invoke("admin/auth/logout").
		AppendMetadata(domain.AdminAuthIdHeader, "0143218411981").
		Do(context.Background())
	s.Require().NoError(err)
}

func (s *AuthTestSuite) Test_Logout_AlreadyRevoke() {
	userId := InsertUser(s.db, entity.User{RoleId: 1})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "token-148623719462",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		CreatedAt: time.Time{},
		ExpiredAt: time.Time{},
	})
	err := s.grpcCli.Invoke("admin/auth/logout").
		AppendMetadata(domain.AdminAuthIdHeader, strconv.Itoa(int(userId))).
		Do(context.Background())
	s.Require().NoError(err)

	tokenInfo := SelectTokenEntityByToken(s.db, "token-148623719462")
	s.Require().Equal(entity.TokenStatusRevoked, tokenInfo.Status)
}

func (s *AuthTestSuite) TestBruteForceLogin() {
	_ = InsertUser(s.db, entity.User{
		RoleId:    1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
	})

	tooManyRequestsErrorCount := &atomic.Int32{}
	unauthorizedErrorCount := &atomic.Int32{}
	group, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < 100; i++ {
		index := i
		group.Go(func() error {
			start := time.Now()
			response := domain.LoginResponse{}
			err := s.grpcCli.Invoke("admin/auth/login").
				JsonRequestBody(domain.LoginRequest{
					Email:    "a@a.ru",
					Password: fmt.Sprintf("password %s", strconv.Itoa(index)),
				}).
				ReadJsonResponse(&response).
				Do(ctx)
			s.Require().Error(err)

			switch status.Code(err) {
			case codes.ResourceExhausted:
				tooManyRequestsErrorCount.Add(1)
				s.Require().True(time.Since(start) < time.Second)
			case codes.Unauthenticated:
				unauthorizedErrorCount.Add(1)
				s.Require().True(time.Since(start) > time.Second)
			default:
				s.Require().NoError(errors.New("never happen"))
			}

			return nil
		})
	}

	err := group.Wait()
	s.Require().NoError(err)

	s.Require().EqualValues(97, tooManyRequestsErrorCount.Load())
	s.Require().EqualValues(3, unauthorizedErrorCount.Load())
}
