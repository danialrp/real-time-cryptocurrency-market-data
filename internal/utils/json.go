package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func RemoveKeysRecursive(data []byte, keysToRemove []string) ([]byte, error) {
	var jsonData interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&jsonData); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	cleaned := removeKeys(jsonData, toKeySet(keysToRemove))

	result, err := json.Marshal(cleaned)
	if err != nil {
		return nil, fmt.Errorf("json encode error: %w", err)
	}

	return result, nil
}

func removeKeys(data interface{}, keys map[string]struct{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key := range v {
			if _, exists := keys[key]; exists {
				delete(v, key)
				continue
			}
			v[key] = removeKeys(v[key], keys)
		}
		return v

	case []interface{}:
		for i := range v {
			v[i] = removeKeys(v[i], keys)
		}
		return v

	default:
		return v
	}
}

func toKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		set[k] = struct{}{}
	}
	return set
}
