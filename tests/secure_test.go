package tests_test

import (
	"context"
	"testing"
	"time"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/dbt"
	"github.com/integration-system/isp-kit/test/grpct"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
	"msp-admin-service/tests"
)

func TestSecureSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SecureSuite{})
}

type SecureSuite struct {
	suite.Suite
	test    *test.Test
	require *require.Assertions
	db      *dbt.TestDb
	grpcCli *client.Client
}

func (s *SecureSuite) SetupTest() {
	s.test, s.require = test.New(s.T())
	s.db = dbt.New(s.test, dbx.WithMigration("../migrations"))
	httpCli := httpcli.New()
	remote := conf.Remote{
		ExpireSec: 3600,
	}
	cfg := assembly.NewLocator(s.test.Logger(), httpCli, s.db).
		Config(context.Background(), remote)

	server, apiCli := grpct.TestServer(s.test, cfg.Handler)
	s.test.T().Cleanup(func() {
		server.Shutdown()
	})
	s.grpcCli = apiCli
}

func (s *SecureSuite) Test_Authenticate_HappyPath() {
	tests.InsertTokenEntity(s.db, entity.Token{
		Token:     "happy_path",
		UserId:    1,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Now().UTC(),
		ExpiredAt: time.Now().UTC().Add(time.Hour),
	})

	result := domain.SecureAuthResponse{}
	err := s.grpcCli.Invoke("admin/secure/authenticate").
		JsonRequestBody(domain.SecureAuthRequest{
			Token: "happy_path",
		}).
		JsonResponseBody(&result).
		Do(context.Background())
	s.require.NoError(err)
	s.require.Equal(domain.SecureAuthResponse{
		Authenticated: true,
		ErrorReason:   "",
		AdminId:       1,
	}, result)
}

func (s *SecureSuite) Test_Authenticate_StatusRevoked() {
	tests.InsertTokenEntity(s.db, entity.Token{
		Token:     "revoked",
		UserId:    1,
		Status:    entity.TokenStatusRevoked,
		CreatedAt: time.Now().UTC(),
		ExpiredAt: time.Now().UTC().Add(time.Hour),
	})

	result := domain.SecureAuthResponse{}
	err := s.grpcCli.Invoke("admin/secure/authenticate").
		JsonRequestBody(domain.SecureAuthRequest{
			Token: "revoked",
		}).
		JsonResponseBody(&result).
		Do(context.Background())
	s.require.NoError(err)
	s.require.Equal(domain.SecureAuthResponse{
		Authenticated: false,
		ErrorReason:   domain.ErrTokenExpired.Error(),
		AdminId:       0,
	}, result)
}

func (s *SecureSuite) Test_Authenticate_Expired() {
	tests.InsertTokenEntity(s.db, entity.Token{
		Token:     "expired",
		UserId:    1,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Now().UTC(),
		ExpiredAt: time.Now().UTC().Add(-2 * time.Hour),
	})

	result := domain.SecureAuthResponse{}
	err := s.grpcCli.Invoke("admin/secure/authenticate").
		JsonRequestBody(domain.SecureAuthRequest{
			Token: "expired",
		}).
		JsonResponseBody(&result).
		Do(context.Background())
	s.require.NoError(err)
	s.require.Equal(domain.SecureAuthResponse{
		Authenticated: false,
		ErrorReason:   domain.ErrTokenExpired.Error(),
		AdminId:       0,
	}, result)
}

func (s *SecureSuite) Test_Authenticate_NotFound() {
	result := domain.SecureAuthResponse{}
	err := s.grpcCli.Invoke("admin/secure/authenticate").
		JsonRequestBody(domain.SecureAuthRequest{
			Token: "not_found",
		}).
		JsonResponseBody(&result).
		Do(context.Background())
	s.require.NoError(err)
	s.require.Equal(domain.SecureAuthResponse{
		Authenticated: false,
		ErrorReason:   domain.ErrTokenNotFound.Error(),
		AdminId:       0,
	}, result)
}
