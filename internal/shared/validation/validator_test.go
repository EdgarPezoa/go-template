package validation

import (
	"reflect"
	"strings"
	"testing"
)

type validationFixture struct {
	Required string `required:"true"`
	Number   int    `min:"1" max:"3"`
	Default  string `default:"ok"`
}

type nestedValidationFixture struct {
	Name string `required:"true" string:"true"`
}

type recursiveValidationFixture struct {
	Pointer *nestedValidationFixture
	Items   []nestedValidationFixture
	Array   [1]*nestedValidationFixture
	Enabled bool   `boolean:"true"`
	Status  string `enum:"ready|not_ready"`
}

type flagValidationFixture struct {
	Required string `required:"true"`
	String   int    `string:"true"`
	Boolean  string `boolean:"true"`
	Bool     string `bool:"true"`
	Number   string `number:"true"`
	Struct   string `struct:"true"`
	Array    string `array:"true"`
}

type noOpValidationFixture struct {
	Required string `required:"false"`
	String   int    `string:"false"`
	Boolean  string `boolean:"false"`
	Bool     string `bool:"false"`
	Number   string `number:"false"`
	Struct   string `struct:"false"`
	Array    string `array:"false"`
	NoTag    int
}

type enumValidationFixture struct {
	Status string `enum:" ready | not ready "`
}

type malformedEnumValidationFixture struct {
	Blank       string `enum:""`
	Leading     string `enum:"|a"`
	Trailing    string `enum:"a|"`
	Consecutive string `enum:"a||b"`
	Whitespace  string `enum:"a|   |b"`
}

type nonStringEnumValidationFixture struct {
	Status int `enum:"ready"`
}

type shapeValidationFixture struct {
	StructValue struct{} `struct:"true"`
	SliceValue  []string `array:"true"`
	ArrayValue  [1]string `array:"true"`
}

func TestValidatorDefaultsAndValidation(t *testing.T) {
	validator := NewValidator()
	fixture := validationFixture{Number: 4}
	validator.ApplyDefaults(&fixture)
	if fixture.Default != "ok" {
		t.Fatal("default missing")
	}
	if validator.Validate(fixture) == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatorFlagsFalseAndAbsentAreInactive(t *testing.T) {
	validator := NewValidator()
	if err := validator.Validate(flagValidationFixture{}); err == nil {
		t.Fatal("expected validation errors")
	}
	err := validator.Validate(noOpValidationFixture{})
	if err != nil {
		t.Fatalf("false and absent tags must be inactive: %v", err)
	}
}

func TestValidatorFlagsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{name: "required", tag: `required:""`, want: "Field has an invalid required tag"},
		{name: "string", tag: `string:"yes"`, want: "Field has an invalid string tag"},
		{name: "boolean", tag: `boolean:"1"`, want: "Field has an invalid boolean tag"},
		{name: "bool", tag: `bool:"no"`, want: "Field has an invalid bool tag"},
		{name: "number", tag: `number:" "` , want: "Field has an invalid number tag"},
		{name: "struct", tag: `struct:"TRUE"`, want: "Field has an invalid struct tag"},
		{name: "array", tag: `array:"invalid"`, want: "Field has an invalid array tag"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixtureType := reflect.StructOf([]reflect.StructField{
				{Name: "Field", Type: reflect.TypeOf(""), Tag: reflect.StructTag(test.tag)},
			})
			err := NewValidator().Validate(reflect.New(fixtureType).Elem().Interface())
			if err == nil || err.Error() != test.want {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidatorFlagErrorsPreserveOrder(t *testing.T) {
	type fixture struct {
		First  int    `required:"" string:"true" boolean:"bad" number:"true" struct:"true" array:"true"`
		Second string `required:"true"`
	}
	err := NewValidator().Validate(fixture{})
	want := "First has an invalid required tag; First must be a string; First has an invalid boolean tag; First must be a struct; First must be an array; Second is required"
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestValidatorBooleanAliases(t *testing.T) {
	tests := []struct {
		name    string
		boolean string
		bool    string
		want    string
	}{
		{name: "both active", boolean: "true", bool: "true", want: "Field must be a boolean"},
		{name: "boolean active", boolean: "true", bool: "false", want: "Field must be a boolean"},
		{name: "bool active", boolean: "false", bool: "true", want: "Field must be a boolean"},
		{name: "both inactive", boolean: "false", bool: "false"},
		{name: "invalid boolean with active bool", boolean: "bad", bool: "true", want: "Field has an invalid boolean tag; Field must be a boolean"},
		{name: "active boolean with invalid bool", boolean: "true", bool: "bad", want: "Field has an invalid bool tag; Field must be a boolean"},
		{name: "both invalid", boolean: "bad", bool: "bad", want: "Field has an invalid boolean tag; Field has an invalid bool tag"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			type fixture struct {
				Field string
			}
			fixtureType := reflect.StructOf([]reflect.StructField{
				{Name: "Field", Type: reflect.TypeOf(""), Tag: reflect.StructTag(`boolean:"` + test.boolean + `" bool:"` + test.bool + `"` )},
			})
			err := NewValidator().Validate(reflect.New(fixtureType).Elem().Interface())
			if test.want == "" {
				if err != nil {
					t.Fatalf("error = %v, want nil", err)
				}
			} else if err == nil || err.Error() != test.want {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidatorStructAndArrayTags(t *testing.T) {
	if err := NewValidator().Validate(shapeValidationFixture{}); err != nil {
		t.Fatalf("valid struct and array values rejected: %v", err)
	}
	type invalid struct {
		StructValue string   `struct:"true"`
		ArrayValue string   `array:"true"`
	}
	if err := NewValidator().Validate(invalid{}); err == nil {
		t.Fatal("expected invalid struct and array errors")
	}
}

func TestValidatorEnum(t *testing.T) {
	if err := NewValidator().Validate(enumValidationFixture{Status: "not ready"}); err != nil {
		t.Fatalf("trimmed enum values should be accepted: %v", err)
	}
	if err := NewValidator().Validate(enumValidationFixture{Status: "unknown"}); err == nil {
		t.Fatal("expected enum validation error")
	}
}

func TestValidatorRejectsMalformedAndNonStringEnums(t *testing.T) {
	if err := NewValidator().Validate(malformedEnumValidationFixture{}); err == nil {
		t.Fatal("expected malformed enum errors")
	}
	if err := NewValidator().Validate(nonStringEnumValidationFixture{}); err == nil {
		t.Fatal("expected non-string enum error")
	}
}

func TestValidatorRecursesThroughPointersSlicesAndArrays(t *testing.T) {
	validator := NewValidator()
	fixture := recursiveValidationFixture{
		Pointer: &nestedValidationFixture{},
		Items:   []nestedValidationFixture{{}},
		Array:   [1]*nestedValidationFixture{{}},
		Status:  "unknown",
	}
	err := validator.Validate(fixture)
	if err == nil {
		t.Fatal("expected recursive validation errors")
	}
	for _, expected := range []string{"Pointer.Name", "Items[0].Name", "Array[0].Name", "Status must be one of ready|not_ready"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("error = %q, missing %q", err, expected)
		}
	}
}
