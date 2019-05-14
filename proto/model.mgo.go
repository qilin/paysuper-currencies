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
    Id            bson.ObjectId `bson:"_id"`
    CreatedAt     time.Time     `bson:"created_at"`
    CreateDate    string        `bson:"create_date"`
    Pair          string        `bson:"pair"`
    Rate          float64       `bson:"rate"`
    Source        string        `bson:"source"`
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
        Pair:          p.Pair,
        Rate:          p.Rate,
        Source:        p.Source,
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
