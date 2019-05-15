package currencyrates

import (
    "errors"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "time"
)

const (
    errorInvalidObjectId = "invalid bson object id"
)

type MgoRateData struct {
    Id         bson.ObjectId `bson:"_id"`
    CreatedAt  time.Time     `bson:"created_at"`
    CreateDate string        `bson:"create_date"`
    Pair       string        `bson:"pair"`
    Rate       float64       `bson:"rate"`
    Source     string        `bson:"source"`
}

type MgoCorrectionRule struct {
    Id               bson.ObjectId      `bson:"_id"`
    MerchantId       bson.ObjectId      `bson:"merchant_id"`
    RateType         string             `bson:"rate_type"`
    CommonCorrection float64            `bson:"common_correction"`
    PairCorrection   map[string]float64 `bson:"pair_correction"`
    CreatedAt        time.Time          `bson:"created_at"`
}

func (p *RateData) SetBSON(raw bson.Raw) error {
    decoded := new(MgoRateData)
    err := raw.Unmarshal(decoded)

    if err != nil {
        return err
    }

    p.Id = decoded.Id.Hex()
    p.Pair = decoded.Pair
    p.Rate = decoded.Rate
    p.Source = decoded.Source

    p.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

    if err != nil {
        return err
    }
    return nil
}

func (p *RateData) GetBSON() (interface{}, error) {
    st := &MgoRateData{
        Pair:   p.Pair,
        Rate:   p.Rate,
        Source: p.Source,
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

    st.CreateDate = st.CreatedAt.Format("2006-01-02")

    return st, nil
}

func (p *CorrectionRule) SetBSON(raw bson.Raw) error {
    decoded := new(MgoCorrectionRule)
    err := raw.Unmarshal(decoded)

    if err != nil {
        return err
    }

    p.Id = decoded.Id.Hex()
    p.MerchantId = decoded.MerchantId.Hex()
    p.RateType = decoded.RateType
    p.CommonCorrection = decoded.CommonCorrection
    p.PairCorrection = decoded.PairCorrection

    p.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)

    if err != nil {
        return err
    }
    return nil
}

func (p *CorrectionRule) GetBSON() (interface{}, error) {
    st := &MgoCorrectionRule{
        RateType:         p.RateType,
        CommonCorrection: p.CommonCorrection,
        PairCorrection:   p.PairCorrection,
    }

    if len(p.Id) <= 0 {
        st.Id = bson.NewObjectId()
    } else {
        if bson.IsObjectIdHex(p.Id) == false {
            return nil, errors.New(errorInvalidObjectId)
        }

        st.Id = bson.ObjectIdHex(p.Id)
    }

    if len(p.MerchantId) > 0 {
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
