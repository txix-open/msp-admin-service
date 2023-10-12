package tests_test

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
	t.db = dbt.New(testInstance, dbx.WithMigration("../migrations"))

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
			},
			AuditTTl: conf.AuditTTlSetting{},
		},
	}
	cfg := assembly.NewLocator(testInstance.Logger(), httpcli.New(), t.db).
		Config(context.Background(), remote)

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
	}
	for _, event := range response {
		name, found := expectedEventList[event.Event]
		t.Require().Equal(found, event.Enabled)
		t.Require().Equal(name, event.Name)
		delete(expectedEventList, event.Event)
	}
	t.Require().Equal(0, len(expectedEventList))
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
		{Event: "новый#2", Enable: false},
	})
	t.NoError(err)

	response := make([]domain.AuditEvent, 0)
	err = t.grpcCli.
		Invoke("admin/log/events").
		JsonResponseBody(&response).
		Do(context.Background())
	t.Require().NoError(err)

	expectedSort := []bool{
		true, true, true, false, false, false, false,
	}
	t.Require().Equal(len(expectedSort), len(response))
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
	t.Require().Equal(0, len(expectedEventList))
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
