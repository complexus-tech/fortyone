package storiesgrp

import (
	"encoding/json"
	"net/url"
	"reflect"
)

// getFilters returns a map of database field names and their values from the given request data.
func getFilters(requestData map[string]json.RawMessage) (map[string]any, error) {
	story := AppFilters{}
	updates := make(map[string]any)
	v := reflect.ValueOf(&story).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		jsonTag := t.Field(i).Tag.Get("json")
		dbTag := t.Field(i).Tag.Get("db")

		if rawValue, ok := requestData[jsonTag]; ok {
			err := json.Unmarshal(rawValue, field.Addr().Interface())
			if err != nil {
				return nil, err
			}
			updates[dbTag] = field.Interface()
		}
	}

	return updates, nil
}

// Convert url.Values to map[string]json.RawMessage
func queryMapToRawMessage(queryMap url.Values) (map[string]json.RawMessage, error) {
	rawMessageMap := make(map[string]json.RawMessage)
	for key, values := range queryMap {
		if len(values) > 0 {
			jsonValue, err := json.Marshal(values[0])
			if err != nil {
				return nil, err
			}
			rawMessageMap[key] = json.RawMessage(jsonValue)
		}
	}
	return rawMessageMap, nil
}
