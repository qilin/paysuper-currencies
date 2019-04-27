package migrations

import (
    migrate "github.com/eminetto/mongo-migrate"
    "github.com/globalsign/mgo"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
)

func init() {
    migrate.Register(func(db *mgo.Database) error { //Up
        db.Session.SetMode(mgo.Monotonic, true)

        err := db.C(pkg.CollectionRate).EnsureIndex(mgo.Index{
            Key: []string{"pair", "is_cb_rate"},
            Name: "pair-type",
        })
        if err != nil {
            return err
        }

        return db.C(pkg.CollectionRate).EnsureIndex(mgo.Index{
            Key: []string{"created_at"},
            Name: "created-at",
        })

    }, func(db *mgo.Database) error { //Down
        return nil
    })
}
