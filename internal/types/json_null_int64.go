package types

import (
	"database/sql"
	"encoding/json"
	"strconv"
)

type JsonNullInt64 struct {
	sql.NullInt64
}

func (j JsonNullInt64) MarshalJSON() ([]byte, error) {
	if j.Valid {
		return json.Marshal(strconv.FormatInt(j.Int64, 10))
	} else {
		return json.Marshal(nil)
	}
}

func (j *JsonNullInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		j.Valid = false
		return nil
	}

	var i int64 = 0
	var str string = ""

	if err := json.Unmarshal(data, &str); err != nil {
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
	}

	j.Valid = true
	if str != "" {
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		j.Int64 = i
	} else if i != 0 {
		j.Int64 = i
	}

	return nil
}

func (j JsonNullInt64) String() string {
	return strconv.FormatInt(j.Int64, 10)
}

func NewJsonNullInt64(i int64) JsonNullInt64 {
	return JsonNullInt64{
		sql.NullInt64{
			Int64: i,
			Valid: true,
		},
	}
}

func NewJsonNullInt64Null() JsonNullInt64 {
	return JsonNullInt64{
		sql.NullInt64{
			Int64: 0,
			Valid: false,
		},
	}
}
