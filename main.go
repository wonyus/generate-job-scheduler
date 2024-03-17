package main

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/wonyus/generate-job-scheduler/context"
	"github.com/wonyus/generate-job-scheduler/handler"
	mssql "github.com/wonyus/generate-job-scheduler/pkg/sql"
	"github.com/wonyus/generate-job-scheduler/repository"
)

func main() {
	db1, err := mssql.NewDefaultClient("iot")
	if err != nil {
		panic(err)
	}
	defer db1.Close()

	db2, err := mssql.NewDefaultClient("taskmanager")
	if err != nil {
		panic(err)
	}
	defer db2.Close()

	h := handler.Handler{
		Context:     context.NewDefaultContext(),
		Repository1: repository.NewSQLRepository(db1),
		Repository2: repository.NewSQLRepository(db2),
	}

	err = h.Handler(20)
	if err != nil {
		fmt.Println(err)
	}
}
