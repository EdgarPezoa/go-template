package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator { return &Validator{} }

func (validator *Validator) ApplyDefaults(target any) {
	reflectedValue := reflect.ValueOf(target)
	if reflectedValue.Kind() != reflect.Pointer || reflectedValue.Elem().Kind() != reflect.Struct {
		return
	}
	applyDefaults(reflectedValue.Elem())
}

func applyDefaults(value reflect.Value) {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field, structField := value.Field(i), valueType.Field(i)
		if field.CanSet() && field.IsZero() && structField.Tag.Get("default") != "" {
			_ = setString(field, structField.Tag.Get("default"))
		}
		applyDefaultsValue(field)
	}
}

func applyDefaultsValue(value reflect.Value) {
	if !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Pointer {
		if !value.IsNil() && value.Elem().Kind() == reflect.Struct {
			applyDefaults(value.Elem())
		}
		return
	}
	if value.Kind() == reflect.Struct {
		applyDefaults(value)
	}
}

func (validator *Validator) Validate(value any) error {
	var validationErrors []string
	validateValue(reflect.ValueOf(value), "", &validationErrors)
	if len(validationErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(validationErrors, "; "))
	}
	return nil
}

func validateValue(value reflect.Value, path string, validationErrors *[]string) {
	if !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Pointer {
		if !value.IsNil() {
			validateValue(value.Elem(), path, validationErrors)
		}
		return
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		for i := 0; i < value.Len(); i++ {
			indexedPath := fmt.Sprintf("%s[%d]", path, i)
			validateValue(value.Index(i), indexedPath, validationErrors)
		}
		return
	}
	if value.Kind() != reflect.Struct {
		return
	}
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field, structField := value.Field(i), valueType.Field(i)
		name := structField.Name
		if path != "" {
			name = path + "." + name
		}
		tag := structField.Tag
		required, requiredValid := parseTagFlag(tag, "required", name, validationErrors)
		if required && requiredValid && field.IsZero() {
			*validationErrors = append(*validationErrors, name+" is required")
		}
		stringTag, stringValid := parseTagFlag(tag, "string", name, validationErrors)
		if stringTag && stringValid && field.Kind() != reflect.String {
			*validationErrors = append(*validationErrors, name+" must be a string")
		}
		booleanTag, booleanValid := parseTagFlag(tag, "boolean", name, validationErrors)
		boolTag, boolValid := parseTagFlag(tag, "bool", name, validationErrors)
		if (booleanTag && booleanValid || boolTag && boolValid) && field.Kind() != reflect.Bool {
			*validationErrors = append(*validationErrors, name+" must be a boolean")
		}
		number, numberValid := parseTagFlag(tag, "number", name, validationErrors)
		if number && numberValid && !isNumber(field.Kind()) {
			*validationErrors = append(*validationErrors, name+" must be a number")
		}
		structTag, structValid := parseTagFlag(tag, "struct", name, validationErrors)
		if structTag && structValid && field.Kind() != reflect.Struct {
			*validationErrors = append(*validationErrors, name+" must be a struct")
		}
		array, arrayValid := parseTagFlag(tag, "array", name, validationErrors)
		if array && arrayValid && field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
			*validationErrors = append(*validationErrors, name+" must be an array")
		}
		if rawEnum, present := tag.Lookup("enum"); present {
			allowed, valid := parseEnum(rawEnum)
			if !valid {
				*validationErrors = append(*validationErrors, name+" has an invalid enum tag")
			} else if field.Kind() != reflect.String {
				*validationErrors = append(*validationErrors, name+" must be a string for enum")
			} else if !isAllowedValue(field, allowed) {
				*validationErrors = append(*validationErrors, name+" must be one of "+strings.Join(allowed, "|"))
			}
		}
		if limit := tag.Get("min"); limit != "" {
			checkNumber(field, name, limit, true, validationErrors)
		}
		if limit := tag.Get("max"); limit != "" {
			checkNumber(field, name, limit, false, validationErrors)
		}
		validateValue(field, name, validationErrors)
	}
}

func parseTagFlag(tag reflect.StructTag, key, name string, validationErrors *[]string) (bool, bool) {
	raw, present := tag.Lookup(key)
	if !present {
		return false, true
	}
	switch raw {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		*validationErrors = append(*validationErrors, name+" has an invalid "+key+" tag")
		return false, false
	}
}

func parseEnum(raw string) ([]string, bool) {
	if raw == "" {
		return nil, false
	}
	parts := strings.Split(raw, "|")
	allowed := make([]string, len(parts))
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, false
		}
		allowed[i] = trimmed
	}
	return allowed, true
}

func isAllowedValue(field reflect.Value, allowed []string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	for _, value := range allowed {
		if field.String() == value {
			return true
		}
	}
	return false
}

func isNumber(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func checkNumber(field reflect.Value, name, rawLimit string, minimum bool, validationErrors *[]string) {
	limit, err := strconv.ParseFloat(rawLimit, 64)
	if err != nil {
		return
	}
	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return
	}
	if (minimum && value < limit) || (!minimum && value > limit) {
		operator := "at most"
		if minimum {
			operator = "at least"
		}
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s must be %s %s", name, operator, rawLimit))
	}
}

func setString(value reflect.Value, text string) error {
	switch value.Kind() {
	case reflect.String:
		value.SetString(text)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, err := strconv.ParseInt(text, 10, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetInt(number)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, err := strconv.ParseUint(text, 10, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetUint(number)
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(text, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetFloat(number)
	default:
		return fmt.Errorf("unsupported default")
	}
	return nil
}
