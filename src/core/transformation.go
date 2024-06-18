package core

import (
	"reflect"
)

func ConvertJSONMapToDBMap(jsonMap map[string]interface{}, model interface{}) map[string]interface{} {
	dbMap := make(map[string]interface{})

	// Get the type of the struct
	structType := reflect.TypeOf(model)

	// Iterate through the input map
	for jsonKey, value := range jsonMap {
		// Find the corresponding struct field based on JSON tag
		field, found := FindFieldByJSONTag(structType, jsonKey)

		// If found, use the db tag as the new key
		if found {
			dbMap[field.Tag.Get("db")] = value
		} else {
			// If not found, use the original key
			dbMap[jsonKey] = value
		}
	}

	return dbMap
}

// FindFieldByJSONTag finds a struct field based on its JSON tag.
func FindFieldByJSONTag(structType reflect.Type, jsonTag string) (reflect.StructField, bool) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("json")
		if tag == jsonTag {
			return field, true
		}
	}
	return reflect.StructField{}, false
}
