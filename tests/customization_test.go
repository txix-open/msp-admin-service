package tests_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/grpct"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
)

func TestCustomizationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CustomizationTestSuite{})
}

type CustomizationTestSuite struct {
	suite.Suite
	test    *test.Test
	db      *dbt.TestDb
	grpcCli *client.Client
}

func (s *CustomizationTestSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))

	remote := conf.Remote{
		UiDesign: conf.UIDesign{
			Name:         "test",
			PrimaryColor: "#ff4d4f",
		},
		ExpireSec: 0,
	}
	cfg := assembly.NewLocator(testInstance.Logger(), nil, s.db).
		Config(context.Background(), emptyLdap, remote, time.Minute)

	server, apiCli := grpct.TestServer(testInstance, cfg.Handler)
	s.grpcCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Shutdown()
	})
}

func (s *CustomizationTestSuite) TestGetDesign() {
	response := conf.UIDesign{}
	err := s.grpcCli.
		Invoke("admin/user/get_design").
		JsonRequestBody(struct{}{}).
		JsonResponseBody(&response).
		Do(context.Background())
	s.Require().NoError(err)
	expected := conf.UIDesign{
		Name:         "test",
		PrimaryColor: "#ff4d4f",
	}
	s.Require().Equal(expected, response)
}
