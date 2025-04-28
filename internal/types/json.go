package types

import (
	"database/sql/driver"
	"encoding/json"
)

type JSON struct {
	json.RawMessage
}

func (j JSON) String() string {
	if j.RawMessage == nil {
		return ""
	}
	if len(j.RawMessage) == 0 {
		return "{}"
	}
	return string(j.RawMessage)
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if data == nil {
		j.RawMessage = nil
		return nil
	}
	if len(data) == 0 {
		j.RawMessage = json.RawMessage{}
		return nil
	}
	RawMessage := json.RawMessage(data)
	j.RawMessage = RawMessage
	return nil
}

func (j *JSON) MarshalJSON() ([]byte, error) {
	if j.RawMessage == nil {
		return nil, nil
	}
	if len(j.RawMessage) == 0 {
		return []byte{}, nil
	}
	return j.RawMessage, nil
}

func (j *JSON) Scan(value any) error {
	if value == nil {
		j.RawMessage = nil
		return nil
	}

	switch value.(type) {
	case []byte:
		j.RawMessage = value.([]byte)
		return nil
	case string:
		j.RawMessage = []byte(value.(string))
		return nil
	default:
		return nil
	}
}

func (j JSON) Value() (driver.Value, error) {
	if j.RawMessage == nil {
		return nil, nil
	}
	if len(j.RawMessage) == 0 {
		return []byte{}, nil
	}

	rawBytes, err := json.Marshal(j.RawMessage)
	if err != nil {
		return nil, err
	}
	return rawBytes, nil
}
