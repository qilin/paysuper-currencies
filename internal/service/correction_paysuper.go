package service

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
	"time"
)

const (
	errorGetCorrection = "can't get correction value"
)

type paysuperCorrection struct {
	Pair      string    `bson:"pair"`
	CreatedAt time.Time `bson:"created_at"`
	Value     float64   `bson:"value"`
}

// GetPaysuperCorrection - returns paysuper correction value for passed pair of currencies
func (s *Service) GetPaysuperCorrection(pair string) (float64, error) {
	if !s.isPairExists(pair) {
		zap.S().Errorw(errorGetCorrection, "error", errorCurrencyPairNotExists, "pair", pair)
		return 0, errors.New(errorCurrencyPairNotExists)
	}

	query := bson.M{"pair": pair}

	res := &paysuperCorrection{}
	err := s.db.Collection(collectionNamePaysuperCorrections).Find(query).Sort("-_id").Limit(1).One(res)
	if err != nil {
		zap.S().Errorw(errorGetCorrection, "error", err, "pair", pair)
		return 0, err
	}

	return res.Value, nil
}
