package storiesgrp

import (
	"encoding/json"
	"reflect"
)

// getUpdates returns a map of database field names and their values from the given request data.
func getUpdates(requestData map[string]json.RawMessage) (map[string]any, error) {
	story := AppUpdateStory{}
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
