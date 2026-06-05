package stream

import "encoding/json"

func encodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
