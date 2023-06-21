package integration_tests

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/testcontainers/testcontainers-go"
	postgrestc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetCleanContainerizedPsqlDb() (*postgresdb.PostgresDatabase, *postgrestc.PostgresContainer) {
	ctx := context.Background()

	files, err := os.ReadDir("/home/anto/GolandProjects/splid_backend_clone/migrations/")
	if err != nil {
		panic(err.Error())
	}

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
	if err != nil {
		panic(err.Error())
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")

	db, err := postgresdb.NewDatabase(connStr)
	if err != nil {
		panic(err.Error())
	}

	// this is where golang-migrate should apply migrations

	return db, container

	// containers can be terminated explicitly but to keep code simpler I rely on
	// testcontainers "garbage collection" with Ryuk https://golang.testcontainers.org/features/garbage_collector/#ryuk
	// that automatically deletes containers when not in use after a while

	// err2 := container.Terminate(context.Background())
	// if err2 != nil {
	// 	panic(err2.Error())
	// }
}
