package mongodb

import (
	"time"

	"github.com/ieshan/timi"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// BSON Helper Functions for timi.Time
// Users can use these functions or copy this pattern to their own code

// MarshalTimiBSON marshals a timi.Time to BSON
func MarshalTimiBSON(t timi.Time) (bson.Type, []byte, error) {
	if !t.Valid {
		return bson.TypeNull, nil, nil
	}
	return bson.MarshalValue(t.Time)
}

// UnmarshalTimiBSON unmarshals BSON data to a timi.Time
func UnmarshalTimiBSON(bType bson.Type, data []byte) (timi.Time, error) {
	if bType == bson.TypeNull {
		return timi.NilTime, nil
	}
	var tv time.Time
	if err := bson.UnmarshalValue(bType, data, &tv); err != nil {
		return timi.NilTime, err
	}
	return timi.Time{Time: tv.UTC(), Valid: true}, nil
}

// TimiBSONWrapper provides a wrapper type that implements BSON marshaling
type TimiBSONWrapper struct {
	timi.Time
}

func (t TimiBSONWrapper) MarshalBSONValue() (bson.Type, []byte, error) {
	return MarshalTimiBSON(t.Time)
}

func (t *TimiBSONWrapper) UnmarshalBSONValue(bType bson.Type, data []byte) error {
	timiTime, err := UnmarshalTimiBSON(bType, data)
	if err != nil {
		return err
	}
	t.Time = timiTime
	return nil
}
