package group_test

import (
	"context"
	"errors"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type GroupTestSuite struct {
	suite.Suite
	psqlContainer *psqlcont.PostgresContainer
	groupService  group.Service
}

func (suite *GroupTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()
	suite.psqlContainer = cont
	suite.groupService = group.NewService(db)
}

func (suite *GroupTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func TestPersonTestSuite(t *testing.T) {
	suite.Run(t, new(GroupTestSuite))
}

func (suite *GroupTestSuite) TestGetGroupByIdFail() {
	g, err := suite.groupService.GetGroupById(context.Background(), 999)
	suite.Assert().True(errors.Is(err, group.ErrGroupNotFound))
	suite.Assert().Equal(group.Group{}, g)
}
