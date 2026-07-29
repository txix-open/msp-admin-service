package tests_test

import (
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

func (s *AuditSuite) SetupTest() {
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
	cfg, err := assembly.NewLocator(testInstance.Logger(), httpcli.New(), s.db).
		Config(s.T().Context(), remote, time.Minute)
	s.Require().NoError(err)
	s.insertAuditLogs()

	server, apiCli := grpct.TestServer(testInstance, cfg.Handler)
	s.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})
}

func (s *AuditSuite) Test_Events_DefaultEvents() {
	response := make([]domain.AuditEvent, 0)
	err := s.grpcCli.
		Invoke("admin/log/events").
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

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
		s.Require().Equal(found, event.Enabled)
		s.Require().Equal(name, event.Name)
		delete(expectedEventList, event.Event)
	}
	s.Require().Empty(expectedEventList)
}

func (s *AuditSuite) Test_Events_SortEvents() {
	eventRep := repository.NewAuditEvent(s.db)
	err := eventRep.Upsert(s.T().Context(), []entity.AuditEvent{
		{Event: "новый#1", Enable: false},
		{Event: entity.EventSuccessLogin, Enable: false},
		{Event: entity.EventErrorLogin, Enable: true},
		{Event: entity.EventSuccessLogout, Enable: true},
		{Event: entity.EventRoleChanged, Enable: false},
		{Event: entity.EventUserChanged, Enable: true},
		{Event: entity.EventUserBlocked, Enable: true},
		{Event: "новый#2", Enable: false},
	})
	s.Require().NoError(err)

	response := make([]domain.AuditEvent, 0)
	err = s.grpcCli.
		Invoke("admin/log/events").
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	expectedSort := []bool{
		true, true, true, true, false, false, false, false,
	}
	s.Require().Equal(len(expectedSort), len(response)) // nolint:testifylint
	for i, event := range response {
		s.Require().Equal(expectedSort[i], event.Enabled)
	}
}

func (s *AuditSuite) Test_SetEvents_HappyPath() {
	err := s.grpcCli.
		Invoke("admin/log/set_events").
		JsonRequestBody([]domain.SetAuditEvent{
			{Event: entity.EventUserChanged, Enabled: true},
			{Event: entity.EventRoleChanged, Enabled: false},
		}).
		Do(s.T().Context())
	s.Require().NoError(err)

	expectedEventList := map[string]bool{
		entity.EventSuccessLogin:  true,
		entity.EventErrorLogin:    true,
		entity.EventSuccessLogout: true,
		entity.EventRoleChanged:   false,
		entity.EventUserChanged:   true,
		entity.EventUserBlocked:   true,
	}
	eventRep := repository.NewAuditEvent(s.db)
	eventList, err := eventRep.All(s.T().Context())
	s.Require().NoError(err)
	for _, event := range eventList {
		enable, found := expectedEventList[event.Event]
		s.Require().True(found)
		s.Require().Equal(enable, event.Enable)
		delete(expectedEventList, event.Event)
	}
	s.Require().Empty(expectedEventList)
}

func (s *AuditSuite) Test_SetEvents_InvalidEvent() {
	err := s.grpcCli.Invoke("admin/log/set_events").
		JsonRequestBody([]domain.SetAuditEvent{
			{Event: entity.EventUserChanged, Enabled: true},
			{Event: "новый#2", Enabled: false},
		}).
		Do(s.T().Context())
	s.Require().Error(err)
	status, isStatus := status.FromError(err)
	s.Require().True(isStatus)
	s.Require().Equal(codes.InvalidArgument, status.Code())
}

func (s *AuditSuite) Test_All_Logs() {
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
	err := s.grpcCli.
		Invoke("admin/log/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 3)
	s.Require().EqualValues(5, response.Items[0].UserId)
	s.Require().EqualValues(4, response.Items[1].UserId)
	s.Require().EqualValues(3, response.Items[2].UserId)

	request.Query = &domain.AuditQuery{
		Message: new("Выход"),
	}
	request.Limit = 10
	request.Offset = 0

	err = s.grpcCli.
		Invoke("admin/log/all").
		JsonRequestBody(request).
		JsonResponseBody(&response).
		Do(s.T().Context())
	s.Require().NoError(err)

	s.Require().Len(response.Items, 6)
	s.Require().EqualValues(6, response.TotalCount)
	s.Require().EqualValues(10, response.Items[0].UserId)
	s.Require().EqualValues(9, response.Items[1].UserId)
	s.Require().EqualValues(8, response.Items[2].UserId)
	s.Require().EqualValues(7, response.Items[3].UserId)
	s.Require().EqualValues(6, response.Items[4].UserId)
	s.Require().EqualValues(5, response.Items[5].UserId)
}

func (s *AuditSuite) insertAuditLogs() {
	s.db.Must().Exec(`INSERT INTO audit (user_id, message, created_at, event)
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
