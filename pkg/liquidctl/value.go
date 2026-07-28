// Package liquidctl provides types and utilities for handling liquidctl device data.
package liquidctl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Value represents a JSON value that can be one of several types (int, string, float, bool).
type Value struct {
	Integer *int64
	String  *string
	Float   *float64
	Bool    *bool
}

func (x *Value) ToString() string {
	if x == nil {
		return ""
	}
	if x.Bool != nil {
		return fmt.Sprintf("%v", *x.Bool)
	} else if x.Float != nil {
		return fmt.Sprintf("%f", *x.Float)
	} else if x.Integer != nil {
		return fmt.Sprintf("%d", *x.Integer)
	} else if x.String != nil {
		return *x.String
	}
	return ""
}

// UnmarshalJSON deserializes JSON data into a Value, supporting multiple value types.
func (x *Value) UnmarshalJSON(data []byte) error {
	object, err := unmarshalUnion(data, &x.Integer, &x.Float, &x.Bool, &x.String, false, nil, false, nil, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

// unmarshalUnion deserializes JSON data into a union type, attempting to match the expected type first.
func unmarshalUnion(data []byte, pi **int64, pf **float64, pb **bool, ps **string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) (bool, error) {
	if pi != nil {
		*pi = nil
	}
	if pf != nil {
		*pf = nil
	}
	if pb != nil {
		*pb = nil
	}
	if ps != nil {
		*ps = nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return false, err
	}

	switch v := tok.(type) {
	case json.Number:
		if pi != nil {
			i, err := v.Int64()
			if err == nil {
				*pi = &i
				return false, nil
			}
		}
		if pf != nil {
			f, err := v.Float64()
			if err == nil {
				*pf = &f
				return false, nil
			}
			return false, errors.New("Unparsable number")
		}
		return false, errors.New("Union does not contain number")
	case float64:
		return false, errors.New("Decoder should not return float64")
	case bool:
		if pb != nil {
			*pb = &v
			return false, nil
		}
		return false, errors.New("Union does not contain bool")
	case string:
		if haveEnum {
			return false, json.Unmarshal(data, pe)
		}
		if ps != nil {
			*ps = &v
			return false, nil
		}
		return false, errors.New("Union does not contain string")
	case nil:
		if nullable {
			return false, nil
		}
		return false, errors.New("Union does not contain null")
	case json.Delim:
		if v == '{' {
			if haveObject {
				return true, json.Unmarshal(data, pc)
			}
			if haveMap {
				return false, json.Unmarshal(data, pm)
			}
			return false, errors.New("Union does not contain object")
		}
		if v == '[' {
			if haveArray {
				return false, json.Unmarshal(data, pa)
			}
			return false, errors.New("Union does not contain array")
		}
		return false, errors.New("Cannot handle delimiter")
	}
	return false, errors.New("Cannot unmarshal union")
}
