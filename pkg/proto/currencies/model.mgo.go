package currencies

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"time"
)

const (
	errorInvalidObjectId = "invalid bson object id"
)

type mgoRateData struct {
	Id         bson.ObjectId `bson:"_id"`
	CreatedAt  time.Time     `bson:"created_at"`
	CreateDate string        `bson:"create_date"`
	Pair       string        `bson:"pair"`
	Rate       float64       `bson:"rate"`
	Source     string        `bson:"source"`
	Volume     float64       `bson:"volume"`
}

// RateData struct for mongoDb
type mgoCorrectionRule struct {
	Id                bson.ObjectId      `bson:"_id"`
	MerchantId        string             `bson:"merchant_id"`
	RateType          string             `bson:"rate_type"`
	ExchangeDirection string             `bson:"exchange_direction"`
	CommonCorrection  float64            `bson:"common_correction"`
	PairCorrection    map[string]float64 `bson:"pair_correction"`
	CreatedAt         time.Time          `bson:"created_at"`
}

// RateData.SetBSON
func (p *RateData) SetBSON(raw bson.Raw) error {
	decoded := new(mgoRateData)
	err := raw.Unmarshal(decoded)

	if err != nil {
		return err
	}

	p.Id = decoded.Id.Hex()
	p.Pair = decoded.Pair
	p.Rate = decoded.Rate
	p.Source = decoded.Source
	p.Volume = decoded.Volume

	p.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

// RateData.GetBSON
func (p *RateData) GetBSON() (interface{}, error) {
	st := &mgoRateData{
		Pair:   p.Pair,
		Rate:   p.Rate,
		Source: p.Source,
		Volume: p.Volume,
	}

	st.Id = bson.NewObjectId()
	if p.CreatedAt != nil {
		t, err := ptypes.Timestamp(p.CreatedAt)

		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	} else {
		st.CreatedAt = time.Now()
	}
	st.CreateDate = st.CreatedAt.Format("2006-01-02")

	return st, nil
}

// CorrectionRule.SetBSON
func (p *CorrectionRule) SetBSON(raw bson.Raw) error {
	decoded := new(mgoCorrectionRule)
	err := raw.Unmarshal(decoded)

	if err != nil {
		return err
	}

	p.Id = decoded.Id.Hex()
	p.MerchantId = decoded.MerchantId
	p.RateType = decoded.RateType
	p.ExchangeDirection = decoded.ExchangeDirection
	p.CommonCorrection = decoded.CommonCorrection
	p.PairCorrection = decoded.PairCorrection

	p.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

// CorrectionRule.GetBSON
func (p *CorrectionRule) GetBSON() (interface{}, error) {
	st := &mgoCorrectionRule{
		RateType:          p.RateType,
		ExchangeDirection: p.ExchangeDirection,
		CommonCorrection:  p.CommonCorrection,
		PairCorrection:    p.PairCorrection,
		MerchantId:        p.MerchantId,
	}

	if len(p.Id) <= 0 {
		st.Id = bson.NewObjectId()
	} else {
		if bson.IsObjectIdHex(p.Id) == false {
			return nil, errors.New(errorInvalidObjectId)
		}

		st.Id = bson.ObjectIdHex(p.Id)
	}

	if p.CreatedAt != nil {
		t, err := ptypes.Timestamp(p.CreatedAt)

		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	} else {
		st.CreatedAt = time.Now()
	}

	return st, nil
}
