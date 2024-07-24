package storiesrepo

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

func getUpdates(story dbStory) (map[string]any, error) {
	updates := make(map[string]any)
	v := reflect.ValueOf(story)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := field.Type()

		dbTag := t.Field(i).Tag.Get("db")
		if dbTag == "" {
			return nil, fmt.Errorf("missing db tag for field %s", t.Field(i).Name)
		}

		// Handle different types
		switch fieldType.Kind() {
		case reflect.Ptr:
			if !field.IsNil() {
				elem := field.Elem()
				if !isZeroValue(elem) {
					updates[dbTag] = elem.Interface()
				}
			}
		case reflect.String:
			if field.String() != "" {
				updates[dbTag] = field.Interface()
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 {
				updates[dbTag] = field.Interface()
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 {
				updates[dbTag] = field.Interface()
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() != 0 {
				updates[dbTag] = field.Interface()
			}
		case reflect.Bool:
			if field.Bool() {
				updates[dbTag] = field.Interface()
			}
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Time{}) {
				if !field.Interface().(time.Time).IsZero() {
					updates[dbTag] = field.Interface()
				}
			} else if fieldType == reflect.TypeOf(uuid.UUID{}) {
				if field.Interface().(uuid.UUID) != uuid.Nil {
					updates[dbTag] = field.Interface()
				}
			} else if !isZeroValue(field) {
				updates[dbTag] = field.Interface()
			}
		default:
			if !isZeroValue(field) {
				updates[dbTag] = field.Interface()
			}
		}
	}

	return updates, nil
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// func getUpdates(story dbStory) (map[string]any, error) {
// 	updates := make(map[string]any)
// 	v := reflect.ValueOf(story)
// 	t := v.Type()

// 	for i := 0; i < v.NumField(); i++ {
// 		field := v.Field(i)
// 		fieldType := field.Type()

// 		dbTag := t.Field(i).Tag.Get("db")
// 		if dbTag == "" {
// 			return nil, fmt.Errorf("missing db tag for field %s", t.Field(i).Name)
// 		}

// 		// Handle pointer types
// 		if fieldType.Kind() == reflect.Ptr {
// 			if !field.IsNil() {
// 				updates[dbTag] = field.Elem().Interface()
// 			}
// 		} else {
// 			// For non-pointer types, add them directly
// 			updates[dbTag] = field.Interface()
// 		}
// 	}

// 	return updates, nil
// }
