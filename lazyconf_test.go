package lazyconf

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// Define a struct to use for testing
type Config struct {
	StringField string `env:"STRING_FIELD"`

	IntField    int    `env:"INT_FIELD"`
	Int8Field   int8   `env:"INT_FIELD"`
	Int16Field  int16  `env:"INT_FIELD"`
	Int32Field  int32  `env:"INT_FIELD"`
	Int64Field  int64  `env:"INT_FIELD"`
	UintField   uint   `env:"INT_FIELD"`
	Uint8Field  uint8  `env:"INT_FIELD"`
	Uint16Field uint16 `env:"INT_FIELD"`
	Uint32Field uint32 `env:"INT_FIELD"`
	Uint64Field uint64 `env:"INT_FIELD"`

	BoolField     bool     `env:"BOOL_FIELD"`
	FloatField    float64  `env:"FLOAT_FIELD"`
	Float32Field  float32  `env:"FLOAT_FIELD"`
	DefaultField  string   `env:"DEFAULT_FIELD,default=defaultValue"`
	RequiredField string   `env:"REQUIRED_FIELD,required"`
	SliceField    []int    `env:"SLICE_FIELD"`
	SliceField8   []int8   `env:"SLICE_FIELD"`
	SliceField16  []int16  `env:"SLICE_FIELD"`
	SliceField32  []int32  `env:"SLICE_FIELD"`
	SliceField64  []int64  `env:"SLICE_FIELD"`
	SliceFieldU   []uint   `env:"SLICE_FIELD"`
	SliceFieldU8  []uint8  `env:"SLICE_FIELD"`
	SliceFieldU16 []uint16 `env:"SLICE_FIELD"`
	SliceFieldU32 []uint32 `env:"SLICE_FIELD"`
	SliceFieldU64 []uint64 `env:"SLICE_FIELD"`

	TimeDuration  time.Duration   `env:"TIME_DURATION_FIELD"`
	TimeDurations []time.Duration `env:"TIME_DURATIONS_FIELD"`

	Time  time.Time   `env:"TIME_FIELD"`
	Times []time.Time `env:"TIMES_FIELD"`

	CustomField      CustomType   `env:"INT_FIELD"`
	CustomFieldSlice []CustomType `env:"SLICE_FIELD"`
}

type CustomType struct {
	Val int64
}

func (c *CustomType) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("invalid CustomType value: %v", value)
	}

	str, ok := value.(string)
	if !ok {
		if b, ok := value.([]byte); ok {
			str = string(b)
		} else {
			return fmt.Errorf("invalid status value: %v", value)
		}
	}

	val, _ := strconv.ParseInt(str, 10, 32)
	*c = CustomType{Val: val}
	return nil
}

// TextUnmarshalType implements encoding.TextUnmarshaler
type TextUnmarshalType struct {
	Value string
}

func (t *TextUnmarshalType) UnmarshalText(text []byte) error {
	t.Value = "text:" + string(text)
	return nil
}

// JSONUnmarshalType implements json.Unmarshaler
type JSONUnmarshalType struct {
	Data map[string]interface{}
}

func (j *JSONUnmarshalType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}

// BothUnmarshalType implements both interfaces
type BothUnmarshalType struct {
	TextValue string
	JSONData  map[string]interface{}
}

func (b *BothUnmarshalType) UnmarshalText(text []byte) error {
	b.TextValue = "text:" + string(text)
	return nil
}

func (b *BothUnmarshalType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.JSONData)
}

// TestParseEnv tests the successful parsing of environment variables into a struct.
func TestParseEnv(t *testing.T) {
	// Set up environment variables
	_ = os.Setenv("STRING_FIELD", "test")
	_ = os.Setenv("INT_FIELD", "42")
	_ = os.Setenv("BOOL_FIELD", "true")
	_ = os.Setenv("FLOAT_FIELD", "3.14")
	_ = os.Setenv("REQUIRED_FIELD", "requiredValue")
	_ = os.Setenv("SLICE_FIELD", "1,2,3")
	_ = os.Setenv("TIME_DURATION_FIELD", "5m")
	_ = os.Setenv("TIME_DURATIONS_FIELD", "5m,10h,1h5m")
	_ = os.Setenv("TIME_FIELD", "2023-07-19T15:30:45Z")
	_ = os.Setenv("TIMES_FIELD", "2023-07-19T15:30:45Z,2023-07-20T10:15:30Z")

	// Create a config instance
	cfg := &Config{}

	// Call ParseEnv
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	// Validate the results
	if cfg.StringField != "test" {
		t.Errorf("expected StringField to be 'test', got '%s'", cfg.StringField)
	}
	if cfg.IntField != 42 {
		t.Errorf("expected IntField to be 42, got %d", cfg.IntField)
	}
	if cfg.BoolField != true {
		t.Errorf("expected BoolField to be true, got %v", cfg.BoolField)
	}
	if cfg.FloatField != 3.14 {
		t.Errorf("expected FloatField to be 3.14, got %f", cfg.FloatField)
	}
	if cfg.DefaultField != "defaultValue" {
		t.Errorf("expected DefaultField to be 'defaultValue', got '%s'", cfg.DefaultField)
	}
	if cfg.RequiredField != "requiredValue" {
		t.Errorf("expected RequiredField to be 'requiredValue', got '%s'", cfg.RequiredField)
	}
	expectedSlice := []int{1, 2, 3}
	if !reflect.DeepEqual(cfg.SliceField, expectedSlice) {
		t.Errorf("expected SliceField to be %v, got %v", expectedSlice, cfg.SliceField)
	}

	// Validate time.Time field
	expectedTime, _ := time.Parse(time.RFC3339, "2023-07-19T15:30:45Z")
	if !cfg.Time.Equal(expectedTime) {
		t.Errorf("expected Time to be %v, got %v", expectedTime, cfg.Time)
	}

	// Validate []time.Time field
	expectedTime1, _ := time.Parse(time.RFC3339, "2023-07-19T15:30:45Z")
	expectedTime2, _ := time.Parse(time.RFC3339, "2023-07-20T10:15:30Z")
	expectedTimes := []time.Time{expectedTime1, expectedTime2}
	if len(cfg.Times) != len(expectedTimes) {
		t.Errorf("expected Times length to be %d, got %d", len(expectedTimes), len(cfg.Times))
	} else {
		for i, expectedT := range expectedTimes {
			if !cfg.Times[i].Equal(expectedT) {
				t.Errorf("expected Times[%d] to be %v, got %v", i, expectedT, cfg.Times[i])
			}
		}
	}
}

// TestParseEnvDefault tests the use of default values for environment variables.
func TestParseEnvDefault(t *testing.T) {
	// Unset environment variables to test default values
	_ = os.Unsetenv("DEFAULT_FIELD")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.DefaultField != "defaultValue" {
		t.Errorf("expected DefaultField to be 'defaultValue', got '%s'", cfg.DefaultField)
	}
}

// TestParseEnvMissingRequired tests the error returned when a required environment variable is missing.
func TestParseEnvMissingRequired(t *testing.T) {
	// Unset the required environment variable
	_ = os.Unsetenv("REQUIRED_FIELD")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when required environment variable is missing, but got none")
	}
}

// TestParseEnvInvalidInt tests the error handling for invalid integer values.
func TestParseEnvInvalidInt(t *testing.T) {
	_ = os.Setenv("INT_FIELD", "notanint")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when INT_FIELD is not a valid integer, but got none")
	}
}

// TestParseEnvInvalidBool tests the error handling for invalid boolean values.
func TestParseEnvInvalidBool(t *testing.T) {
	_ = os.Setenv("BOOL_FIELD", "notabool")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when BOOL_FIELD is not a valid boolean, but got none")
	}
}

// TestParseEnvInvalidFloat tests the error handling for invalid float values.
func TestParseEnvInvalidFloat(t *testing.T) {
	_ = os.Setenv("FLOAT_FIELD", "notafloat")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when FLOAT_FIELD is not a valid float, but got none")
	}
}

// TestParseEnvInvalidSlice tests the error handling for invalid slice values.
func TestParseEnvInvalidSlice(t *testing.T) {
	_ = os.Setenv("SLICE_FIELD", "1,notanint,3")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when SLICE_FIELD contains invalid integers, but got none")
	}
}

// TestParseEnvInvalidTime tests the error handling for invalid time values.
func TestParseEnvInvalidTime(t *testing.T) {
	_ = os.Setenv("TIME_FIELD", "not-a-time")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when TIME_FIELD is not a valid time, but got none")
	}
}

// TestParseEnvInvalidTimeSlice tests the error handling for invalid time slice values.
func TestParseEnvInvalidTimeSlice(t *testing.T) {
	_ = os.Setenv("TIMES_FIELD", "2023-07-19T15:30:45Z,not-a-time,2023-07-20T10:15:30Z")

	cfg := &Config{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when TIMES_FIELD contains invalid time, but got none")
	}
}

// TestParseEnvUnsupportedType tests the error handling for unsupported field types.
func TestParseEnvUnsupportedType(t *testing.T) {
	type UnsupportedType map[string]string

	type UnsupportedConfig struct {
		ComplexField UnsupportedType `env:"UNSUPPORTED_TYPE"`
	}

	_ = os.Setenv("UNSUPPORTED_TYPE", "test")

	cfg := &UnsupportedConfig{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when parsing unsupported complex type, but got none")
	}
}

// TestParseEnvUnexported tests the error handling for unexported fields.
func TestParseEnvUnexported(t *testing.T) {
	type unexported struct {
		field string `env:"UNEXPORTED_FIELD"`
	}

	_ = os.Setenv("UNEXPORTED_FIELD", "test")

	cfg := &unexported{}

	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when parsing unexported field, but got none")
	}
}

// TestParseEnvComplexTypes tests parsing of complex number types.
func TestParseEnvComplexTypes(t *testing.T) {
	type ComplexConfig struct {
		Complex64Field  complex64  `env:"COMPLEX64_FIELD"`
		Complex128Field complex128 `env:"COMPLEX128_FIELD"`
	}

	_ = os.Setenv("COMPLEX64_FIELD", "1+2i")
	_ = os.Setenv("COMPLEX128_FIELD", "3.5+4.2i")

	cfg := &ComplexConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expected64 := complex64(1 + 2i)
	expected128 := complex128(3.5 + 4.2i)

	if cfg.Complex64Field != expected64 {
		t.Errorf("expected Complex64Field to be %v, got %v", expected64, cfg.Complex64Field)
	}
	if cfg.Complex128Field != expected128 {
		t.Errorf("expected Complex128Field to be %v, got %v", expected128, cfg.Complex128Field)
	}
}

// TestParseEnvInvalidComplex tests error handling for invalid complex values.
func TestParseEnvInvalidComplex(t *testing.T) {
	type ComplexConfig struct {
		ComplexField complex128 `env:"COMPLEX_FIELD"`
	}

	_ = os.Setenv("COMPLEX_FIELD", "notacomplex")

	cfg := &ComplexConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when COMPLEX_FIELD is not a valid complex number, but got none")
	}
}

// SetterConfig for testing custom setter methods
type SetterConfig struct {
	CustomField string `env:"CUSTOM_FIELD,setter=SetCustomField"`
	value       string
}

// SetCustomField is a custom setter method
func (c *SetterConfig) SetCustomField(val string) error {
	c.value = "custom_" + val
	return nil
}

// SetterConfigNotFound for testing missing setter methods
type SetterConfigNotFound struct {
	CustomField string `env:"CUSTOM_FIELD,setter=NonExistentMethod"`
}

// SetterConfigError for testing failing setter methods
type SetterConfigError struct {
	CustomField string `env:"CUSTOM_FIELD,setter=FailingSetter"`
}

// FailingSetter is a setter method that always fails
func (c *SetterConfigError) FailingSetter(val string) error {
	return fmt.Errorf("setter failed")
}

// ErrorCustomType for testing Setter interface error handling
type ErrorCustomType struct {
	Val int64
}

// Scan method that always fails
func (c *ErrorCustomType) Scan(value interface{}) error {
	return fmt.Errorf("scan error")
}

// ErrorSliceCustomType for testing custom type slice error handling
type ErrorSliceCustomType struct {
	Val int64
}

// Scan method that always fails for slice elements
func (c *ErrorSliceCustomType) Scan(value interface{}) error {
	return fmt.Errorf("slice scan error")
}

// TestParseEnvCustomSetter tests custom setter method functionality.
func TestParseEnvCustomSetter(t *testing.T) {
	_ = os.Setenv("CUSTOM_FIELD", "test")

	cfg := &SetterConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.value != "custom_test" {
		t.Errorf("expected value to be 'custom_test', got '%s'", cfg.value)
	}
}

// TestParseEnvCustomSetterNotFound tests error when custom setter method is not found.
func TestParseEnvCustomSetterNotFound(t *testing.T) {
	_ = os.Setenv("CUSTOM_FIELD", "test")

	cfg := &SetterConfigNotFound{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when custom setter method is not found, but got none")
	}
}

// TestParseEnvCustomSetterError tests error handling when custom setter method fails.
func TestParseEnvCustomSetterError(t *testing.T) {
	_ = os.Setenv("CUSTOM_FIELD", "test")

	cfg := &SetterConfigError{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when custom setter method fails, but got none")
	}
}

// TestParseEnvSpecialEnvKey tests handling of special envKey="_".
func TestParseEnvSpecialEnvKey(t *testing.T) {
	type SpecialConfig struct {
		EmptyField string `env:"_"`
	}

	cfg := &SpecialConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.EmptyField != "" {
		t.Errorf("expected EmptyField to be empty, got '%s'", cfg.EmptyField)
	}
}

// TestParseEnvNestedStruct tests recursive parsing of nested structs.
func TestParseEnvNestedStruct(t *testing.T) {
	type NestedConfig struct {
		NestedField string `env:"NESTED_FIELD"`
	}

	type ParentConfig struct {
		ParentField string `env:"PARENT_FIELD"`
		Nested      NestedConfig
	}

	_ = os.Setenv("PARENT_FIELD", "parent_value")
	_ = os.Setenv("NESTED_FIELD", "nested_value")

	cfg := &ParentConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.ParentField != "parent_value" {
		t.Errorf("expected ParentField to be 'parent_value', got '%s'", cfg.ParentField)
	}
	if cfg.Nested.NestedField != "nested_value" {
		t.Errorf("expected Nested.NestedField to be 'nested_value', got '%s'", cfg.Nested.NestedField)
	}
}

// TestParseEnvAllIntegerTypes tests all integer type variants.
func TestParseEnvAllIntegerTypes(t *testing.T) {
	type IntConfig struct {
		Int8Field  int8  `env:"INT8_FIELD"`
		Int16Field int16 `env:"INT16_FIELD"`
		Int32Field int32 `env:"INT32_FIELD"`
		Int64Field int64 `env:"INT64_FIELD"`
	}

	_ = os.Setenv("INT8_FIELD", "127")
	_ = os.Setenv("INT16_FIELD", "32767")
	_ = os.Setenv("INT32_FIELD", "2147483647")
	_ = os.Setenv("INT64_FIELD", "9223372036854775807")

	cfg := &IntConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.Int8Field != 127 {
		t.Errorf("expected Int8Field to be 127, got %d", cfg.Int8Field)
	}
	if cfg.Int16Field != 32767 {
		t.Errorf("expected Int16Field to be 32767, got %d", cfg.Int16Field)
	}
	if cfg.Int32Field != 2147483647 {
		t.Errorf("expected Int32Field to be 2147483647, got %d", cfg.Int32Field)
	}
	if cfg.Int64Field != 9223372036854775807 {
		t.Errorf("expected Int64Field to be 9223372036854775807, got %d", cfg.Int64Field)
	}
}

// TestParseEnvAllUnsignedIntegerTypes tests all unsigned integer type variants.
func TestParseEnvAllUnsignedIntegerTypes(t *testing.T) {
	type UintConfig struct {
		UintField   uint   `env:"UINT_FIELD"`
		Uint8Field  uint8  `env:"UINT8_FIELD"`
		Uint16Field uint16 `env:"UINT16_FIELD"`
		Uint32Field uint32 `env:"UINT32_FIELD"`
		Uint64Field uint64 `env:"UINT64_FIELD"`
	}

	_ = os.Setenv("UINT_FIELD", "4294967295")
	_ = os.Setenv("UINT8_FIELD", "255")
	_ = os.Setenv("UINT16_FIELD", "65535")
	_ = os.Setenv("UINT32_FIELD", "4294967295")
	_ = os.Setenv("UINT64_FIELD", "18446744073709551615")

	cfg := &UintConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.UintField != 4294967295 {
		t.Errorf("expected UintField to be 4294967295, got %d", cfg.UintField)
	}
	if cfg.Uint8Field != 255 {
		t.Errorf("expected Uint8Field to be 255, got %d", cfg.Uint8Field)
	}
	if cfg.Uint16Field != 65535 {
		t.Errorf("expected Uint16Field to be 65535, got %d", cfg.Uint16Field)
	}
	if cfg.Uint32Field != 4294967295 {
		t.Errorf("expected Uint32Field to be 4294967295, got %d", cfg.Uint32Field)
	}
	if cfg.Uint64Field != 18446744073709551615 {
		t.Errorf("expected Uint64Field to be 18446744073709551615, got %d", cfg.Uint64Field)
	}
}

// TestParseEnvInvalidUint tests error handling for invalid unsigned integer values.
func TestParseEnvInvalidUint(t *testing.T) {
	type UintConfig struct {
		UintField uint `env:"UINT_FIELD"`
	}

	_ = os.Setenv("UINT_FIELD", "notauint")

	cfg := &UintConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when UINT_FIELD is not a valid unsigned integer, but got none")
	}
}

// TestParseEnvFloatTypes tests float32 vs float64 distinction.
func TestParseEnvFloatTypes(t *testing.T) {
	type FloatConfig struct {
		Float32Field float32 `env:"FLOAT32_FIELD"`
		Float64Field float64 `env:"FLOAT64_FIELD"`
	}

	_ = os.Setenv("FLOAT32_FIELD", "3.14")
	_ = os.Setenv("FLOAT64_FIELD", "2.718281828")

	cfg := &FloatConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.Float32Field != 3.14 {
		t.Errorf("expected Float32Field to be 3.14, got %f", cfg.Float32Field)
	}
	if cfg.Float64Field != 2.718281828 {
		t.Errorf("expected Float64Field to be 2.718281828, got %f", cfg.Float64Field)
	}
}

// TestParseEnvSliceTypes tests various slice types.
func TestParseEnvSliceTypes(t *testing.T) {
	type SliceConfig struct {
		StringSlice  []string  `env:"STRING_SLICE"`
		Float32Slice []float32 `env:"FLOAT32_SLICE"`
		Float64Slice []float64 `env:"FLOAT64_SLICE"`
		BoolSlice    []bool    `env:"BOOL_SLICE"`
	}

	_ = os.Setenv("STRING_SLICE", "hello,world,test")
	_ = os.Setenv("FLOAT32_SLICE", "1.1,2.2,3.3")
	_ = os.Setenv("FLOAT64_SLICE", "1.11,2.22,3.33")
	_ = os.Setenv("BOOL_SLICE", "true,false,true")

	cfg := &SliceConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expectedStringSlice := []string{"hello", "world", "test"}
	if !reflect.DeepEqual(cfg.StringSlice, expectedStringSlice) {
		t.Errorf("expected StringSlice to be %v, got %v", expectedStringSlice, cfg.StringSlice)
	}

	expectedFloat32Slice := []float32{1.1, 2.2, 3.3}
	if !reflect.DeepEqual(cfg.Float32Slice, expectedFloat32Slice) {
		t.Errorf("expected Float32Slice to be %v, got %v", expectedFloat32Slice, cfg.Float32Slice)
	}

	expectedFloat64Slice := []float64{1.11, 2.22, 3.33}
	if !reflect.DeepEqual(cfg.Float64Slice, expectedFloat64Slice) {
		t.Errorf("expected Float64Slice to be %v, got %v", expectedFloat64Slice, cfg.Float64Slice)
	}

	expectedBoolSlice := []bool{true, false, true}
	if !reflect.DeepEqual(cfg.BoolSlice, expectedBoolSlice) {
		t.Errorf("expected BoolSlice to be %v, got %v", expectedBoolSlice, cfg.BoolSlice)
	}
}

// TestParseEnvInvalidSliceTypes tests error handling for invalid slice element values.
func TestParseEnvInvalidSliceTypes(t *testing.T) {
	tests := []struct {
		name     string
		config   interface{}
		envVar   string
		envValue string
	}{
		{
			name: "InvalidFloatSlice",
			config: &struct {
				FloatSlice []float64 `env:"FLOAT_SLICE"`
			}{},
			envVar:   "FLOAT_SLICE",
			envValue: "1.1,notafloat,3.3",
		},
		{
			name: "InvalidBoolSlice",
			config: &struct {
				BoolSlice []bool `env:"BOOL_SLICE"`
			}{},
			envVar:   "BOOL_SLICE",
			envValue: "true,notabool,false",
		},
		{
			name: "InvalidUintSlice",
			config: &struct {
				UintSlice []uint `env:"UINT_SLICE"`
			}{},
			envVar:   "UINT_SLICE",
			envValue: "1,notauint,3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(tt.envVar, tt.envValue)
			err := ParseEnv(tt.config)
			if err == nil {
				t.Fatalf("expected an error for %s, but got none", tt.name)
			}
		})
	}
}

// TestParseEnvTimeDurationSlice tests time.Duration slice parsing.
func TestParseEnvTimeDurationSlice(t *testing.T) {
	type DurationConfig struct {
		Durations []time.Duration `env:"DURATIONS"`
	}

	_ = os.Setenv("DURATIONS", "1m,2h,30s")

	cfg := &DurationConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expected := []time.Duration{
		1 * time.Minute,
		2 * time.Hour,
		30 * time.Second,
	}

	if !reflect.DeepEqual(cfg.Durations, expected) {
		t.Errorf("expected Durations to be %v, got %v", expected, cfg.Durations)
	}
}

// TestParseEnvInvalidTimeDuration tests error handling for invalid time.Duration values.
func TestParseEnvInvalidTimeDuration(t *testing.T) {
	type DurationConfig struct {
		Duration time.Duration `env:"DURATION"`
	}

	_ = os.Setenv("DURATION", "invalid_duration")

	cfg := &DurationConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when DURATION is not a valid time.Duration, but got none")
	}
}

// TestParseEnvInvalidTimeDurationSlice tests error handling for invalid time.Duration slice values.
func TestParseEnvInvalidTimeDurationSlice(t *testing.T) {
	type DurationConfig struct {
		Durations []time.Duration `env:"DURATIONS"`
	}

	_ = os.Setenv("DURATIONS", "1m,invalid_duration,30s")

	cfg := &DurationConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when DURATIONS contains invalid time.Duration, but got none")
	}
}

// TestParseEnvSetterInterfaceError tests error handling for Setter interface.
func TestParseEnvSetterInterfaceError(t *testing.T) {
	type SetterErrorConfig struct {
		CustomField ErrorCustomType `env:"CUSTOM_FIELD"`
	}

	_ = os.Setenv("CUSTOM_FIELD", "42")

	cfg := &SetterErrorConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when Setter.Scan fails, but got none")
	}
}

// TestParseEnvCustomTypeSlice tests custom type slices with Setter interface.
func TestParseEnvCustomTypeSlice(t *testing.T) {
	type SliceCustomConfig struct {
		CustomSlice []CustomType `env:"CUSTOM_SLICE"`
	}

	_ = os.Setenv("CUSTOM_SLICE", "10,20,30")

	cfg := &SliceCustomConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expected := []CustomType{
		{Val: 10},
		{Val: 20},
		{Val: 30},
	}

	if !reflect.DeepEqual(cfg.CustomSlice, expected) {
		t.Errorf("expected CustomSlice to be %v, got %v", expected, cfg.CustomSlice)
	}
}

// TestParseEnvCustomTypeSliceError tests error handling for custom type slices.
func TestParseEnvCustomTypeSliceError(t *testing.T) {
	type SliceErrorConfig struct {
		CustomSlice []ErrorSliceCustomType `env:"CUSTOM_SLICE"`
	}

	_ = os.Setenv("CUSTOM_SLICE", "10,20,30")

	cfg := &SliceErrorConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when custom type slice Setter.Scan fails, but got none")
	}
}

// TestParseEnvUnsupportedSliceType tests error handling for unsupported slice element types.
func TestParseEnvUnsupportedSliceType(t *testing.T) {
	type UnsupportedSliceConfig struct {
		UnsupportedSlice []map[string]string `env:"UNSUPPORTED_SLICE"`
	}

	_ = os.Setenv("UNSUPPORTED_SLICE", "test")

	cfg := &UnsupportedSliceConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when parsing unsupported slice element type, but got none")
	}
}

// TestParseEnvParserText tests parser="text" tag functionality.
func TestParseEnvParserText(t *testing.T) {
	type TextParserConfig struct {
		TextField TextUnmarshalType `env:"TEXT_FIELD,parser=text"`
	}

	_ = os.Setenv("TEXT_FIELD", "hello")

	cfg := &TextParserConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expected := "text:hello"
	if cfg.TextField.Value != expected {
		t.Errorf("expected TextField.Value to be '%s', got '%s'", expected, cfg.TextField.Value)
	}
}

// TestParseEnvParserJSON tests parser="json" tag functionality.
func TestParseEnvParserJSON(t *testing.T) {
	type JSONParserConfig struct {
		JSONField JSONUnmarshalType `env:"JSON_FIELD,parser=json"`
	}

	_ = os.Setenv("JSON_FIELD", `{"key":"value","number":42}`)

	cfg := &JSONParserConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.JSONField.Data["key"] != "value" {
		t.Errorf("expected JSONField.Data[\"key\"] to be 'value', got '%v'", cfg.JSONField.Data["key"])
	}
	if cfg.JSONField.Data["number"] != float64(42) {
		t.Errorf("expected JSONField.Data[\"number\"] to be 42, got '%v'", cfg.JSONField.Data["number"])
	}
}

// TestParseEnvParserBothText tests parser="text" with type that implements both interfaces.
func TestParseEnvParserBothText(t *testing.T) {
	type BothParserConfig struct {
		BothField BothUnmarshalType `env:"BOTH_FIELD,parser=text"`
	}

	_ = os.Setenv("BOTH_FIELD", "hello")

	cfg := &BothParserConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expected := "text:hello"
	if cfg.BothField.TextValue != expected {
		t.Errorf("expected BothField.TextValue to be '%s', got '%s'", expected, cfg.BothField.TextValue)
	}
	if cfg.BothField.JSONData != nil {
		t.Errorf("expected BothField.JSONData to be nil, got '%v'", cfg.BothField.JSONData)
	}
}

// TestParseEnvParserBothJSON tests parser="json" with type that implements both interfaces.
func TestParseEnvParserBothJSON(t *testing.T) {
	type BothParserConfig struct {
		BothField BothUnmarshalType `env:"BOTH_FIELD,parser=json"`
	}

	_ = os.Setenv("BOTH_FIELD", `{"test":"data"}`)

	cfg := &BothParserConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.BothField.JSONData["test"] != "data" {
		t.Errorf("expected BothField.JSONData[\"test\"] to be 'data', got '%v'", cfg.BothField.JSONData["test"])
	}
	if cfg.BothField.TextValue != "" {
		t.Errorf("expected BothField.TextValue to be empty, got '%s'", cfg.BothField.TextValue)
	}
}

// TestParseEnvFallbackText tests fallback to UnmarshalText without parser tag.
func TestParseEnvFallbackText(t *testing.T) {
	type FallbackConfig struct {
		TextField TextUnmarshalType `env:"TEXT_FIELD"`
	}

	_ = os.Setenv("TEXT_FIELD", "fallback")

	cfg := &FallbackConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	expected := "text:fallback"
	if cfg.TextField.Value != expected {
		t.Errorf("expected TextField.Value to be '%s', got '%s'", expected, cfg.TextField.Value)
	}
}

// TestParseEnvFallbackJSON tests fallback to UnmarshalJSON without parser tag.
func TestParseEnvFallbackJSON(t *testing.T) {
	type FallbackConfig struct {
		JSONField JSONUnmarshalType `env:"JSON_FIELD"`
	}

	_ = os.Setenv("JSON_FIELD", `{"fallback":"test"}`)

	cfg := &FallbackConfig{}
	err := ParseEnv(cfg)
	if err != nil {
		t.Fatalf("ParseEnv returned an error: %v", err)
	}

	if cfg.JSONField.Data["fallback"] != "test" {
		t.Errorf("expected JSONField.Data[\"fallback\"] to be 'test', got '%v'", cfg.JSONField.Data["fallback"])
	}
}

// TestParseEnvParserTextError tests error when parser="text" but type doesn't implement TextUnmarshaler.
func TestParseEnvParserTextError(t *testing.T) {
	type ErrorConfig struct {
		StringField string `env:"STRING_FIELD,parser=text"`
	}

	_ = os.Setenv("STRING_FIELD", "test")

	cfg := &ErrorConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when parser=text but type doesn't implement TextUnmarshaler, but got none")
	}
}

// TestParseEnvParserJSONError tests error when parser="json" but type doesn't implement JSONUnmarshaler.
func TestParseEnvParserJSONError(t *testing.T) {
	type ErrorConfig struct {
		StringField string `env:"STRING_FIELD,parser=json"`
	}

	_ = os.Setenv("STRING_FIELD", "test")

	cfg := &ErrorConfig{}
	err := ParseEnv(cfg)
	if err == nil {
		t.Fatal("expected an error when parser=json but type doesn't implement JSONUnmarshaler, but got none")
	}
}
