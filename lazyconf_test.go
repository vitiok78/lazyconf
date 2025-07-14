package lazyconf

import (
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

	// TODO support time.Time
	// Time          time.Time       `env:"TIME_FIELD"`

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

	// TODO support time.Time
	// _ = os.Setenv("TIME_FIELD", time.Now().String())

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
