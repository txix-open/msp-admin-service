package tests_test

import (
	"strconv"
	"testing"
	"time"

	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
	"msp-admin-service/service/session_worker"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/bgjobx"
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
	config  assembly.Config
	grpcCli *client.Client
}

func (s *SessionSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))

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
	config, err := assembly.NewLocator(testInstance.Logger(), httpcli.New(), s.db).
		Config(s.T().Context(), remote, 500*time.Millisecond)
	s.Require().NoError(err)

	s.config = config

	server, apiCli := grpct.TestServer(testInstance, s.config.Handler)
	s.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})
}

//nolint:funlen
func (s *SessionSuite) Test_All_Session() {
	userId := InsertUser(s.db, entity.User{Email: "test_1@aa.ru"})
	timeNow := time.Now().UTC()

	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token_1",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: timeNow.Add(1 * time.Hour),
		CreatedAt: timeNow.Add(-1 * time.Hour),
		UpdatedAt: timeNow,
	})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token_2",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: timeNow.Add(1 * time.Hour),
		CreatedAt: timeNow.Add(-2 * time.Hour),
		UpdatedAt: timeNow,
	})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token_3",
		UserId:    userId,
		Status:    entity.TokenStatusExpired,
		ExpiredAt: timeNow.Add(-1 * time.Hour),
		CreatedAt: timeNow.Add(-3 * time.Hour),
		UpdatedAt: timeNow,
	})

	// Дефолт сортировка, лимит и оффсет
	request := domain.SessionPageRequest{
		LimitOffestParams: domain.LimitOffestParams{
			Limit:  5,
			Offset: 1,
		},
	}

	var response *domain.SessionResponse
	err := s.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 2)
	s.Require().EqualValues(3, response.TotalCount)
	s.Require().EqualValues(2, response.Items[0].Id)
	s.Require().EqualValues(3, response.Items[1].Id)

	// Дефолт сортировка, поиск по expired_at
	request.Offset = 0
	request.Query = &domain.SessionQuery{
		ExpiredAt: &domain.DateFromToParams{
			From: timeNow,
			To:   timeNow.Add(24 * time.Hour),
		},
	}
	err = s.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 2)
	s.Require().EqualValues(2, response.TotalCount)
	s.Require().EqualValues(1, response.Items[0].Id)
	s.Require().EqualValues(2, response.Items[1].Id)

	// Сортировка по статусу, пустой запрос
	request.Query = nil
	request.Order = &domain.OrderParams{
		Field: "status",
		Type:  "asc",
	}

	err = s.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 3)
	s.Require().EqualValues(3, response.TotalCount)
	s.Require().EqualValues(1, response.Items[0].Id)
	s.Require().EqualValues(3, response.Items[1].Id)
	s.Require().EqualValues(2, response.Items[2].Id)

	// Сортировка по статусу, поиск по userId & status
	resUserId := int(userId)
	request.Query = &domain.SessionQuery{
		UserId: []int{resUserId},
		Status: []string{entity.TokenStatusExpired},
	}

	err = s.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 1)
	s.Require().EqualValues(1, response.TotalCount)
	s.Require().EqualValues(3, response.Items[0].Id)

	// Сортировка по id, поиск по id
	for i := range 20 {
		userId = InsertUser(s.db, entity.User{Email: "test_11" + strconv.Itoa(i) + "@aa.ru"})
		InsertTokenEntity(s.db, entity.Token{
			Token:     "test_token_1" + strconv.Itoa(i),
			UserId:    userId,
			Status:    entity.TokenStatusAllowed,
			ExpiredAt: timeNow.Add(1 * time.Hour),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

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
			Id: new(2),
		},
	}
	err = s.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 5)
	s.Require().EqualValues(6, response.TotalCount)
	s.Require().EqualValues(23, response.Items[0].Id)
	s.Require().EqualValues(22, response.Items[1].Id)
	s.Require().EqualValues(21, response.Items[2].Id)
	s.Require().EqualValues(20, response.Items[3].Id)
	s.Require().EqualValues(12, response.Items[4].Id)
}

func (s *SessionSuite) Test_Session_Expired_Worker() {
	userId := InsertUser(s.db, entity.User{Email: "a@test"})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "token_allowed",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().UTC().Add(24 * time.Hour)})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "token_expired",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().UTC().Add(-2 * time.Hour)})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "token_expired2",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().UTC().Add(-24 * time.Hour)})

	bgjobCli := bgjobx.NewClient(s.db, s.test.Logger())

	err := session_worker.EnqueueSeedJob(s.T().Context(), bgjobCli)
	s.Require().NoError(err)

	err = bgjobCli.Upgrade(s.T().Context(), s.config.BgJobCfg)
	s.Require().NoError(err)

	time.Sleep(2 * time.Second)

	tokens := make([]entity.Token, 0)
	s.db.Must().Select(&tokens, "SELECT * FROM tokens ORDER BY status ASC")

	s.Require().EqualValues(entity.TokenStatusAllowed, tokens[0].Status)
	s.Require().EqualValues(entity.TokenStatusExpired, tokens[1].Status)
	s.Require().EqualValues(entity.TokenStatusExpired, tokens[2].Status)
}
