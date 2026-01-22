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
	t.config = assembly.NewLocator(testInstance.Logger(), httpcli.New(), t.db).
		Config(context.Background(), remote, 500*time.Millisecond)

	server, apiCli := grpct.TestServer(testInstance, t.config.Handler)
	t.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})
}

//nolint:funlen
func (t *SessionSuite) Test_All_Session() {
	userId := InsertUser(t.db, entity.User{Email: "test_1@aa.ru"})
	timeNow := time.Now().UTC()

	InsertTokenEntity(t.db, entity.Token{
		Token:     "test_token_1",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: timeNow.Add(1 * time.Hour),
		CreatedAt: timeNow.Add(-1 * time.Hour),
		UpdatedAt: timeNow,
	})
	InsertTokenEntity(t.db, entity.Token{
		Token:     "test_token_2",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: timeNow.Add(1 * time.Hour),
		CreatedAt: timeNow.Add(-2 * time.Hour),
		UpdatedAt: timeNow,
	})
	InsertTokenEntity(t.db, entity.Token{
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
	err := t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 2)
	t.Require().EqualValues(3, response.TotalCount)
	t.Require().EqualValues(2, response.Items[0].Id)
	t.Require().EqualValues(3, response.Items[1].Id)

	// Дефолт сортировка, поиск по expired_at
	request.Offset = 0
	request.Query = &domain.SessionQuery{
		ExpiredAt: &domain.DateFromToParams{
			From: timeNow,
			To:   timeNow.Add(24 * time.Hour),
		},
	}
	err = t.grpcCli.
		Invoke("admin/session/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 2)
	t.Require().EqualValues(2, response.TotalCount)
	t.Require().EqualValues(1, response.Items[0].Id)
	t.Require().EqualValues(2, response.Items[1].Id)

	// Сортировка по статусу, пустой запрос
	request.Query = nil
	request.Order = &domain.OrderParams{
		Field: "status",
		Type:  "asc",
	}

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

	// Сортировка по статусу, поиск по userId & status
	resUserId := int(userId)
	reqStatus := entity.TokenStatusExpired
	request.Query = &domain.SessionQuery{
		UserId: &resUserId,
		Status: &reqStatus,
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

	// Сортировка по id, поиск по id
	for i := range 20 {
		userId = InsertUser(t.db, entity.User{Email: "test_11" + strconv.Itoa(i) + "@aa.ru"})
		InsertTokenEntity(t.db, entity.Token{
			Token:     "test_token_1" + strconv.Itoa(i),
			UserId:    userId,
			Status:    entity.TokenStatusAllowed,
			ExpiredAt: timeNow.Add(1 * time.Hour),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
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

func (t *SessionSuite) Test_Session_Expired_Worker() {
	userId := InsertUser(t.db, entity.User{Email: "a@test"})

	InsertTokenEntity(t.db, entity.Token{
		Token:     "token_allowed",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().UTC().Add(24 * time.Hour)})

	InsertTokenEntity(t.db, entity.Token{
		Token:     "token_expired",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().UTC().Add(-2 * time.Hour)})

	InsertTokenEntity(t.db, entity.Token{
		Token:     "token_expired2",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: time.Now().UTC().Add(-24 * time.Hour)})

	bgjobCli := bgjobx.NewClient(t.db, t.test.Logger())

	err := session_worker.EnqueueSeedJob(t.T().Context(), bgjobCli)
	t.Require().NoError(err)

	err = bgjobCli.Upgrade(t.T().Context(), t.config.BgJobCfg)
	t.Require().NoError(err)

	time.Sleep(2 * time.Second)

	tokens := make([]entity.Token, 0)
	t.db.Must().Select(&tokens, "SELECT * FROM tokens ORDER BY status ASC")

	t.Require().EqualValues(entity.TokenStatusAllowed, tokens[0].Status)
	t.Require().EqualValues(entity.TokenStatusExpired, tokens[1].Status)
	t.Require().EqualValues(entity.TokenStatusExpired, tokens[2].Status)
}
