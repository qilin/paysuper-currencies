package utils

import "github.com/paysuper/paysuper-currencies-rates/config"

func GetMongoUrl(cfg *config.Config) string {
    mongoUrl := ""
    if cfg.MongoUser != "" {
        mongoUrl += cfg.MongoUser
        if cfg.MongoPassword != "" {
            mongoUrl += ":" + cfg.MongoPassword
        }
        mongoUrl += "@"
    }
    mongoUrl += cfg.MongoHost
    return mongoUrl
}

