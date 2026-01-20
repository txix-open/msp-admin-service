package tests_test

import (
	"context"
	"testing"
	"time"

	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
	"msp-admin-service/repository"

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

func TestAuditSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &AuditSuite{})
}

type AuditSuite struct {
	suite.Suite

	test    *test.Test
	db      *dbt.TestDb
	grpcCli *client.Client
}

func (t *AuditSuite) SetupTest() {
	testInstance, _ := test.New(t.T())
	t.test = testInstance
	t.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))
	insertAuditLogs(t.db)

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

func (t *AuditSuite) Test_Events_DefaultEvents() {
	response := make([]domain.AuditEvent, 0)
	err := t.grpcCli.
		Invoke("admin/log/events").
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	expectedEventList := map[string]string{
		entity.EventSuccessLogin:  "успешный вход",
		entity.EventErrorLogin:    "неуспешный вход",
		entity.EventSuccessLogout: "успешный выход",
		entity.EventRoleChanged:   "изменение роли",
		entity.EventUserChanged:   "изменение пользователя",
		entity.EventUserBlocked:   "изменение статуса блокировки пользователя",
	}
	for _, event := range response {
		name, found := expectedEventList[event.Event]
		t.Require().Equal(found, event.Enabled)
		t.Require().Equal(name, event.Name)
		delete(expectedEventList, event.Event)
	}
	t.Require().Empty(expectedEventList)
}

func (t *AuditSuite) Test_Events_SortEvents() {
	eventRep := repository.NewAuditEvent(t.db)
	err := eventRep.Upsert(context.Background(), []entity.AuditEvent{
		{Event: "новый#1", Enable: false},
		{Event: entity.EventSuccessLogin, Enable: false},
		{Event: entity.EventErrorLogin, Enable: true},
		{Event: entity.EventSuccessLogout, Enable: true},
		{Event: entity.EventRoleChanged, Enable: false},
		{Event: entity.EventUserChanged, Enable: true},
		{Event: entity.EventUserBlocked, Enable: true},
		{Event: "новый#2", Enable: false},
	})
	t.Require().NoError(err)

	response := make([]domain.AuditEvent, 0)
	err = t.grpcCli.
		Invoke("admin/log/events").
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	expectedSort := []bool{
		true, true, true, true, false, false, false, false,
	}
	t.Require().Equal(len(expectedSort), len(response)) // nolint:testifylint
	for i, event := range response {
		t.Require().Equal(expectedSort[i], event.Enabled)
	}
}

func (t *AuditSuite) Test_SetEvents_HappyPath() {
	err := t.grpcCli.
		Invoke("admin/log/set_events").
		JsonRequestBody([]domain.SetAuditEvent{
			{Event: entity.EventUserChanged, Enabled: true},
			{Event: entity.EventRoleChanged, Enabled: false},
		}).
		Do(context.Background())
	t.Require().NoError(err)

	expectedEventList := map[string]bool{
		entity.EventSuccessLogin:  true,
		entity.EventErrorLogin:    true,
		entity.EventSuccessLogout: true,
		entity.EventRoleChanged:   false,
		entity.EventUserChanged:   true,
		entity.EventUserBlocked:   true,
	}
	eventRep := repository.NewAuditEvent(t.db)
	eventList, err := eventRep.All(context.Background())
	t.Require().NoError(err)
	for _, event := range eventList {
		enable, found := expectedEventList[event.Event]
		t.Require().True(found)
		t.Require().Equal(enable, event.Enable)
		delete(expectedEventList, event.Event)
	}
	t.Require().Empty(expectedEventList)
}

func (t *AuditSuite) Test_SetEvents_InvalidEvent() {
	err := t.grpcCli.
		Invoke("admin/log/set_events").
		JsonRequestBody([]domain.SetAuditEvent{
			{Event: entity.EventUserChanged, Enabled: true},
			{Event: "новый#2", Enabled: false},
		}).
		Do(context.Background())
	t.Require().Error(err)
	s, isStatus := status.FromError(err)
	t.Require().True(isStatus)
	t.Require().Equal(codes.InvalidArgument, s.Code())
}

func (t *AuditSuite) Test_All_Logs() {
	request := domain.AuditPageRequest{
		LimitOffestParams: domain.LimitOffestParams{
			Limit:  3,
			Offset: 5,
		},
		Order: &domain.OrderParams{
			Field: "user_id",
			Type:  "desc",
		},
	}

	var response *domain.AuditResponse
	err := t.grpcCli.
		Invoke("admin/log/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 3)
	t.Require().EqualValues(5, response.Items[0].UserId)
	t.Require().EqualValues(4, response.Items[1].UserId)
	t.Require().EqualValues(3, response.Items[2].UserId)

	msg := "Успешный выход"
	request.Query = &domain.AuditQuery{
		Message: &msg,
	}
	request.Limit = 10
	request.Offset = 0

	err = t.grpcCli.
		Invoke("admin/log/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	t.Require().Len(response.Items, 2)
	t.Require().EqualValues(6, response.Items[0].UserId)
	t.Require().EqualValues(5, response.Items[1].UserId)
}

func insertAuditLogs(testDb *dbt.TestDb) {
	testDb.Must().Exec(`INSERT INTO audit (user_id, message, created_at, event)
	VALUES (1, 'Успешный вход', NOW(), 'success_login'),
	       (2, 'Успешный вход', NOW(), 'success_login'),
	       (3, 'Неуспешный вход', NOW(), 'unsuccess_login'),
	       (4, 'Неуспешный вход', NOW(), 'unsuccess_login'),
	       (5, 'Успешный выход', NOW(), 'success_logout'),
	       (6, 'Успешный выход', NOW(), 'success_logout'),
	       (7, 'Неуспешный выход', NOW(), 'unsuccess_logout'),
	       (8, 'Неуспешный выход', NOW(), 'unsuccess_logout'),
	       (9, 'Выход', NOW(), 'logout'),
	       (10, 'Выход', NOW(), 'logout')`)
}
