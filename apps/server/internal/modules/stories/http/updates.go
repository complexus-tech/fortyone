package storieshttp

import (
	"encoding/json"
	"reflect"
	"time"
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

			// Handle date.Date fields specially - convert to *time.Time
			if field.Type().Implements(reflect.TypeOf((*interface{ TimePtr() *time.Time })(nil)).Elem()) {
				if !field.IsNil() {
					dateValue := field.Interface().(interface{ TimePtr() *time.Time })
					updates[dbTag] = dateValue.TimePtr()
				} else {
					updates[dbTag] = nil
				}
			} else {
				updates[dbTag] = field.Interface()
			}
		}
	}

	return updates, nil
}
