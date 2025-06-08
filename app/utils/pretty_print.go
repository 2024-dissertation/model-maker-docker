package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	colorRed     = "\033[1;31m" // Bold Red
	colorGreen   = "\033[1;32m" // Bold Green
	colorYellow  = "\033[1;33m" // Bold Yellow
	colorBlue    = "\033[1;34m" // Bold Blue
	colorMagenta = "\033[1;35m" // Bold Magenta
	colorCyan    = "\033[1;36m" // Bold Cyan
	colorReset   = "\033[0m"
)

// PrettyPrint prints any object in a colorful, formatted way
func PrettyPrint(obj interface{}) {
	// Get the type and value of the object
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	// Print type name as header
	typeName := t.String()
	fmt.Printf("\n%s╔═ %s ═╗%s\n", colorCyan, strings.Repeat("═", len(typeName)), colorReset)
	fmt.Printf("%s║ %s%s%s ║%s\n", colorCyan, colorYellow, typeName, colorCyan, colorReset)
	fmt.Printf("%s╚%s%s\n", colorCyan, strings.Repeat("═", len(typeName)+4), colorReset)

	// If it's a struct, print each field
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)

			// Format the value based on its kind
			var valueStr string
			switch value.Kind() {
			case reflect.String:
				valueStr = fmt.Sprintf("%s\"%v\"%s", colorGreen, value.Interface(), colorReset)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				valueStr = fmt.Sprintf("%s%v%s", colorBlue, value.Interface(), colorReset)
			case reflect.Bool:
				if value.Bool() {
					valueStr = fmt.Sprintf("%strue%s", colorGreen, colorReset)
				} else {
					valueStr = fmt.Sprintf("%sfalse%s", colorRed, colorReset)
				}
			case reflect.Slice, reflect.Array:
				bytes, err := json.MarshalIndent(value.Interface(), "", "  ")
				if err == nil {
					valueStr = fmt.Sprintf("%s%s%s", colorMagenta, string(bytes), colorReset)
				} else {
					valueStr = fmt.Sprintf("%s%v%s", colorMagenta, value.Interface(), colorReset)
				}
			default:
				valueStr = fmt.Sprintf("%v", value.Interface())
			}

			fmt.Printf("%s%s:%s %s\n", colorYellow, field.Name, colorReset, valueStr)
		}
	} else {
		// For non-struct types, just print the value
		bytes, err := json.MarshalIndent(obj, "", "  ")
		if err == nil {
			fmt.Printf("%s%s%s\n", colorMagenta, string(bytes), colorReset)
		} else {
			fmt.Printf("%v\n", obj)
		}
	}
	fmt.Println()
}
