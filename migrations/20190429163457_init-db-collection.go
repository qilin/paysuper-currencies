package migrations

import (
    migrate "github.com/eminetto/mongo-migrate"
    "github.com/globalsign/mgo"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
)

func init() {
    migrate.Register(func(db *mgo.Database) error { //Up
        db.Session.SetMode(mgo.Monotonic, true)
        return db.C(pkg.CollectionRate).Create(&mgo.CollectionInfo{})
    }, func(db *mgo.Database) error { //Down
        return nil
    })
}
