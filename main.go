package main

import (
    "fmt"
	"flag"
	"log"
    "gobank/storage"
    "gobank/api"
    "gobank/types"
)

func seedAccount(store storage.Storage, firstName, lastName, pw string) *types.Account {
    acc, err := types.NewAccount(firstName, lastName, pw)
    if err != nil {
        log.Fatal(err)
    }

    if err := store.CreateAccount(acc); err != nil {
       log.Fatal(err) 
    }

    fmt.Println("new account => ", acc)

    return acc
}

func seedAccounts(s storage.Storage) {
    seedAccount(s, "lolname", "lollastname", "hunter999")
}

func main()  {
    seed := flag.Bool("seed", false, "seed the db")
    flag.Parse()

    store, err := storage.NewPostgresStore()
    if err != nil {
        log.Fatal(err)
    }

    if err := store.Init(); err != nil  {
        log.Fatal(err)
    }

    if *seed {
        fmt.Println("seeding the database")
        seedAccounts(store)
    }


    server := api.NewApiServer(":3000", store)
    if  err := server.Run(); err != nil {
        log.Fatal(err)
    }
}
