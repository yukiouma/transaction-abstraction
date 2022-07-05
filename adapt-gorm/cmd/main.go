package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AkiOuma/transaction-abstraction/adapt-gorm/datasource"
	"github.com/AkiOuma/transaction-abstraction/adapt-gorm/usecase"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	addr := "root:000000@tcp(127.0.0.1:3306)/tx-demo?parseTime=True"
	gormDB := datasource.NewGormDB(addr)
	ds := datasource.NewDatasource(gormDB)
	uc := usecase.NewUsecase(ds)

	// with out transaction
	if err := uc.CreateUser(ctx, "yuki"); err != nil {
		log.Println(err)
	}

	// with concurrency and transaction
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			if err := uc.CreateAdminUser(ctx, fmt.Sprintf("user %d", i)); err != nil {
				log.Println(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
