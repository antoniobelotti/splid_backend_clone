//go:build integration
// +build integration

package person_test

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	postgrestc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type PersonTestSuite struct {
	suite.Suite
	psqlContainer *postgrestc.PostgresContainer
	personService person.Service
}

func TestPersonTestSuite(t *testing.T) {
	suite.Run(t, new(PersonTestSuite))
}

func (suite *PersonTestSuite) TearDownTest() {
	ctx := context.Background()

	suite.Require().NoError(suite.psqlContainer.Terminate(ctx))
}

func (suite *PersonTestSuite) SetupTest() {
	ctx := context.Background()

	files, err := os.ReadDir("/home/anto/GolandProjects/splid_backend_clone/migrations/")
	suite.Require().NoError(err)

	/* I was unable to make golang-migrate work in this setting.
	as a workaround, pass all *.up.sql migration files to the container. They will be executed after start up,
	hopefully in the correct order
	*/
	var migrationFiles []string
	for _, file := range files {
		if strings.Contains(file.Name(), "up") {
			migrationFiles = append(migrationFiles, filepath.Join("/home/anto/GolandProjects/splid_backend_clone/migrations/", file.Name()))
		}
	}

	container, err := postgrestc.RunContainer(ctx,
		testcontainers.WithImage("postgres:12-alpine"),
		postgrestc.WithInitScripts(migrationFiles...),
		postgrestc.WithDatabase("postgres"),
		postgrestc.WithUsername("postgres"),
		postgrestc.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	suite.Require().NoError(err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")

	suite.psqlContainer = container

	db, err := postgresdb.NewDatabase(connStr)
	suite.Require().NoError(err)

	// this is where golang-migrate should apply migrations

	suite.personService = person.NewService(db)
}

func (suite *PersonTestSuite) TestGetPersonByEmailFails() {
	p, err := suite.personService.GetPersonByEmail(context.Background(), "nonExistentEmail")

	suite.Equal(person.ErrPersonNotFound, err)
	suite.Equal(person.Person{}, p)
}
