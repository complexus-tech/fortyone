package web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
)

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

// GetFilters returns a map of database field names and their values from the given request data.
func GetFilters(queryMap url.Values, data any) (map[string]any, error) {

	requestData, err := queryMapToRawMessage(queryMap)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]any)
	// v := reflect.ValueOf(&data).Elem()
	// t := v.Type()

	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("data must be a pointer to a struct")
	}
	v = v.Elem()
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
