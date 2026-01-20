package tests_test

import (
	"context"
	"strconv"
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

//nolint:funlen
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
	t.Require().EqualValues(3, response.TotalCount)
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
	t.Require().EqualValues(1, response.TotalCount)
	t.Require().EqualValues(3, response.Items[0].Id)

	for i := range 20 {
		userId = InsertUser(t.db, entity.User{Email: "test_11" + strconv.Itoa(i) + "@aa.ru"})
		InsertTokenEntity(t.db, entity.Token{
			Token:     "test_token_1" + strconv.Itoa(i),
			UserId:    userId,
			Status:    entity.TokenStatusAllowed,
			ExpiredAt: userTime3.Add(1 * time.Hour),
			CreatedAt: userTime3,
			UpdatedAt: userTime3,
		})
	}

	tokenId := 2
	request = domain.SessionPageRequest{
		LimitOffestParams: domain.LimitOffestParams{
			Limit:  5,
			Offset: 0,
		},
		Order: &domain.OrderParams{
			Field: "id",
			Type:  "desc",
		},
		Query: &domain.SessionQuery{
			Id: &tokenId,
		},
	}
	err = t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 5)
	t.Require().EqualValues(5, response.TotalCount)
	t.Require().EqualValues(23, response.Items[0].Id)
	t.Require().EqualValues(22, response.Items[1].Id)
	t.Require().EqualValues(21, response.Items[2].Id)
	t.Require().EqualValues(20, response.Items[3].Id)
	t.Require().EqualValues(2, response.Items[4].Id)
}

func (t *SessionSuite) Test_All_Session_Expired_Revoked() {
	userId := InsertUser(t.db, entity.User{Email: "test_1@aa.ru"})

	userTime, err := time.Parse("2006-01-02T15:04:05Z", "2025-01-01T00:00:00Z")
	t.Require().NoError(err)
	expiredTime, err := time.Parse("2006-01-02T15:04:05Z", "2426-01-01T00:00:00Z")
	t.Require().NoError(err)

	InsertTokenEntity(t.db, entity.Token{
		Token:     "allowed_token",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: expiredTime,
		CreatedAt: userTime,
		UpdatedAt: userTime,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "revoked_token",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: expiredTime,
		CreatedAt: userTime,
		UpdatedAt: userTime,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "expired_token",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: userTime.Add(1 * time.Hour),
		CreatedAt: userTime,
		UpdatedAt: userTime,
	})

	status := "EXPIRED"
	request := domain.SessionPageRequest{
		LimitOffestParams: domain.LimitOffestParams{
			Limit:  5,
			Offset: 0,
		},
		Order: &domain.OrderParams{
			Field: "expired_at",
			Type:  "desc",
		},
		Query: &domain.SessionQuery{
			Status: &status,
		},
	}

	var response *domain.SessionResponse
	err = t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 1)
	t.Require().EqualValues(1, response.TotalCount)
	t.Require().EqualValues(3, response.Items[0].Id)

	status = "REVOKED"
	request.Query.Status = &status

	err = t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 1)
	t.Require().EqualValues(1, response.TotalCount)
	t.Require().EqualValues(2, response.Items[0].Id)
}

func (t *SessionSuite) Test_All_Session_Status() {
	userId := InsertUser(t.db, entity.User{Email: "test_1@aa.ru"})

	userTime, err := time.Parse("2006-01-02T15:04:05Z", "2025-01-01T00:00:00Z")
	t.Require().NoError(err)
	expiredTime, err := time.Parse("2006-01-02T15:04:05Z", "2426-01-01T00:00:00Z")
	t.Require().NoError(err)

	InsertTokenEntity(t.db, entity.Token{
		Token:     "allowed_token",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: expiredTime,
		CreatedAt: userTime,
		UpdatedAt: userTime,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "revoked_token",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: expiredTime,
		CreatedAt: userTime,
		UpdatedAt: userTime,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "expired_token",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: userTime.Add(1 * time.Hour),
		CreatedAt: userTime,
		UpdatedAt: userTime,
	})

	request := domain.SessionPageRequest{
		LimitOffestParams: domain.LimitOffestParams{
			Limit:  5,
			Offset: 0,
		},
		Order: &domain.OrderParams{
			Field: "status",
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

	t.Require().Len(response.Items, 3)
	t.Require().EqualValues(3, response.TotalCount)
	t.Require().EqualValues(1, response.Items[0].Id)
	t.Require().EqualValues(3, response.Items[1].Id)
	t.Require().EqualValues(2, response.Items[2].Id)
}
