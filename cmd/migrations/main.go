package main

import (
    "fmt"
    migrate "github.com/eminetto/mongo-migrate"
    "github.com/globalsign/mgo"
    "github.com/paysuper/paysuper-currencies-rates/config"
    _ "github.com/paysuper/paysuper-currencies-rates/migrations"
    "go.uber.org/zap"
    "io"
    "log"
    "os"
    "time"
)

func main() {

    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatal("Config init failed with error", zap.Error(err))
    }

    if len(os.Args) == 1 {
        log.Fatal("Missing options: up or down")
    }
    option := os.Args[1]

    mongoUrl := ""
    if cfg.MongoUser != "" {
        mongoUrl += cfg.MongoUser
        if cfg.MongoPassword != "" {
            mongoUrl += ":" + cfg.MongoPassword
        }
        mongoUrl += "@"
    }
    mongoUrl += cfg.MongoHost
    session, err := mgo.Dial(mongoUrl)
    if err != nil {
        log.Fatal(err.Error())
    }
    defer session.Close()
    db := session.DB(cfg.MongoDatabase)
    migrate.SetDatabase(db)
    migrate.SetMigrationsCollection("migrations")
    migrate.SetLogger(log.New(os.Stdout, "INFO: ", 0))
    switch option {
    case "new":
        if len(os.Args) != 3 {
            log.Fatal("Should be: new description-of-migration")
        }
        fName := fmt.Sprintf("./migrations/%s_%s.go", time.Now().Format("20060102150405"), os.Args[2])
        from, err := os.Open("./migrations/template.go")
        if err != nil {
            log.Fatal("Should be: new description-of-migration")
        }
        defer from.Close()

        to, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0666)
        if err != nil {
            log.Fatal(err.Error())
        }
        defer to.Close()

        _, err = io.Copy(to, from)
        if err != nil {
            log.Fatal(err.Error())
        }
        log.Printf("New migration created: %s\n", fName)
    case "up":
        err = migrate.Up(migrate.AllAvailable)
    case "down":
        err = migrate.Down(migrate.AllAvailable)
    }
    if err != nil {
        log.Fatal(err.Error())
    }
}
