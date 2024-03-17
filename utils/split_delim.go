package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Marshal marshals a struct into a string using dynamic field positions
func Marshal(data interface{}) (string, error) {
	// Get the type of the struct
	dataType := reflect.TypeOf(data)

	// Initialize a slice to store field values
	var values []string

	// Iterate over the fields of the struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		// Get the pos tag of the field
		_, exists := field.Tag.Lookup("pos")
		if !exists {
			continue // Skip fields without pos tag
		}

		// Get the value of the field
		fieldValue := reflect.ValueOf(data).Field(i)

		// Convert the field value to a string
		value := fmt.Sprintf("%v", fieldValue.Interface())

		// Append the value to the slice
		values = append(values, value)
	}

	// Join field values with the delimiter
	return strings.Join(values, "|"), nil
}

// Unmarshal unmarshals a string into a struct using dynamic field positions
func Unmarshal(data string, v interface{}) error {
	// Split the data by the delimiter
	parts := strings.Split(data, "|")

	// Get the type of the struct
	dataType := reflect.TypeOf(v).Elem()

	// Iterate over the fields of the struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		// Get the pos tag of the field
		positionTag, exists := field.Tag.Lookup("pos")
		if !exists {
			continue // Skip fields without pos tag
		}

		// Convert the pos tag to an integer
		pos, err := strconv.Atoi(positionTag)
		if err != nil {
			return err
		}

		// Check if the pos is within the bounds of the parts slice
		if pos-1 < len(parts) {
			// Set the value of the field
			fieldValue := reflect.ValueOf(v).Elem().Field(i)
			fieldValue.SetString(parts[pos-1]) // Subtract 1 to account for zero-based indexing
		}
	}

	return nil
}
