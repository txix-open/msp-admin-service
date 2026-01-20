package tests_test

import (
	"context"
	"testing"
	"time"

	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/grpct"
)

func TestSessionSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SessionSuite{})
}

type SessionSuite struct {
	suite.Suite

	test    *test.Test
	db      *dbt.TestDb
	grpcCli *client.Client
}

func (t *SessionSuite) SetupTest() {
	testInstance, _ := test.New(t.T())
	t.test = testInstance
	t.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))

	remote := conf.Remote{
		Audit: conf.Audit{
			EventSettings: []conf.AuditEventSetting{
				{
					Event: entity.EventSuccessLogin,
					Name:  "успешный вход",
				},
				{
					Event: entity.EventErrorLogin,
					Name:  "неуспешный вход",
				},
				{
					Event: entity.EventSuccessLogout,
					Name:  "успешный выход",
				},
				{
					Event: entity.EventRoleChanged,
					Name:  "изменение роли",
				},
				{
					Event: entity.EventUserChanged,
					Name:  "изменение пользователя",
				},
				{
					Event: entity.EventUserBlocked,
					Name:  "изменение статуса блокировки пользователя",
				},
			},
			AuditTTl: conf.AuditTTlSetting{},
		},
	}
	cfg := assembly.NewLocator(testInstance.Logger(), httpcli.New(), t.db).
		Config(context.Background(), remote, time.Minute)

	server, apiCli := grpct.TestServer(testInstance, cfg.Handler)
	t.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})
}

func (t *SessionSuite) Test_All_Session() {
	userId := InsertUser(t.db, entity.User{Email: "test_1@aa.ru"})

	userTime1, err := time.Parse("2006-01-02T15:04:05Z", "2018-01-01T00:00:00Z")
	t.Require().NoError(err)
	userTime2, err := time.Parse("2006-01-02T15:04:05Z", "2018-02-01T00:00:00Z")
	t.Require().NoError(err)
	userTime3, err := time.Parse("2006-01-02T15:04:05Z", "2019-01-01T00:00:00Z")
	t.Require().NoError(err)

	InsertTokenEntity(t.db, entity.Token{
		Token:     "test_token_1",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: userTime1.Add(1 * time.Hour),
		CreatedAt: userTime1,
		UpdatedAt: userTime1,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "test_token_2",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: userTime2.Add(1 * time.Hour),
		CreatedAt: userTime2,
		UpdatedAt: userTime2,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "test_token_3",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: userTime3.Add(1 * time.Hour),
		CreatedAt: userTime3,
		UpdatedAt: userTime3,
	})

	request := domain.SessionPageRequest{
		LimitOffestParams: domain.LimitOffestParams{
			Limit:  5,
			Offset: 1,
		},
		Order: &domain.OrderParams{
			Field: "created_at",
			Type:  "asc",
		},
	}

	var response *domain.SessionResponse
	err = t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 2)
	t.Require().EqualValues(2, response.Items[0].Id)
	t.Require().EqualValues(3, response.Items[1].Id)

	request.Offset = 0
	request.Query = &domain.SessionQuery{
		ExpiredAt: &domain.DateFromToParams{
			From: userTime3.Add(-24 * time.Hour),
			To:   userTime3.Add(24 * time.Hour),
		},
	}

	err = t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 1)
	t.Require().EqualValues(3, response.Items[0].Id)
}
