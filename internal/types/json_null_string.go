package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JsonNullString struct {
	sql.NullString
}

func NewJsonNullString(s string) JsonNullString {
	return JsonNullString{
		sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

func NewJsonNullStringFromNull() JsonNullString {
	return JsonNullString{
		sql.NullString{
			String: "",
			Valid:  false,
		},
	}
}

func (j JsonNullString) MarshalJSON() ([]byte, error) {
	if j.Valid {
		return json.Marshal(j.String)
	} else {
		return json.Marshal(nil)
	}
}

func (j *JsonNullString) UnmarshalJSON(data []byte) error {
	var str *string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if str != nil {
		j.Valid = true
		j.String = *str
	} else {
		j.Valid = false
	}
	return nil
}

func (j *JsonNullString) Scan(value any) error {
	if value == nil {
		j.Valid = false
		j.String = ""
		return nil
	}

	switch value := value.(type) {
	case string:
		j.Valid = true
		j.String = value
		return nil
	case []byte:
		j.Valid = true
		j.String = string(value)
		return nil
	default:
		return fmt.Errorf("jsonNullString: unsupported type: %T", value)
	}
}

func (j JsonNullString) Value() (driver.Value, error) {
	if j.Valid {
		return j.String, nil
	}
	return nil, nil
}
