package tests_test

import (
	"context"
	"testing"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/dbt"
	"github.com/integration-system/isp-kit/test/grpct"
	"github.com/stretchr/testify/suite"
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
	s.db = dbt.New(testInstance, dbx.WithMigration("../migrations"))

	remote := conf.Remote{
		UiDesign: conf.UIDesign{
			Name:         "test",
			PrimaryColor: "#ff4d4f",
		},
		ExpireSec: 0,
	}
	cfg := assembly.NewLocator(testInstance.Logger(), nil, s.db).
		Config(context.Background(), emptyLdap, remote)

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
