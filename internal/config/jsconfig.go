package config

import (
	"encoding/json"
	"os"
)

func ReadJSONConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil

}

func AsString(v interface{}) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

func AsInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case int64:
		return int(val), true
	case int32:
		return int(val), true
	default:
		return 0, false
	}
}

func AsBool(v interface{}) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}
