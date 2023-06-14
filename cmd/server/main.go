package main

import (
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"os"
)

func Run() error {
	db, err := postgresdb.NewDatabase()
	if err != nil {
		return err
	}
	fmt.Println("successfully connected to db")

	ps := person.NewService(db)
	gs := group.NewService(db)

	restServer := http.NewRESTServer(ps, gs)
	err = restServer.Run(os.Getenv("HTTP_PORT"))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println("main running")
	err := Run()
	if err != nil {
		fmt.Println(err)
	}
}
