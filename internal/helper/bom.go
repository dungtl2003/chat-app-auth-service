package helper

import "encoding/json"

func ParseAsJson(data []byte, target any) error {
	return json.Unmarshal(data, target)
}
