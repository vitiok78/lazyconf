package lazyconf

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const setterMethodName = "Scan"

type Setter interface {
	Scan(value interface{}) error
}

func ParseEnv(cfg any) error {
	op := "xconf.ParseEnv"

	val := reflect.ValueOf(cfg)
	v := val.Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("env")

		// If the field is a struct, recursively parse it
		if field.Type.Kind() == reflect.Struct {
			if err := ParseEnv(v.Field(i).Addr().Interface()); err != nil {
				return err
			}
		}

		// If the field is not tagged, skip it
		if tag == "" {
			continue
		}

		// Parse the tag
		parts := strings.Split(tag, ",")
		envKey := parts[0]
		required := false
		defaultVal := ""
		setterName := ""

		// Parse the tag options
		for _, opt := range parts[1:] {
			if opt == "required" {
				required = true
			} else if strings.HasPrefix(opt, "default=") {
				defaultVal = strings.TrimPrefix(opt, "default=")
			} else if strings.HasPrefix(opt, "setter=") {
				setterName = strings.TrimPrefix(opt, "setter=")
			}
		}

		// Get the value from the environment
		var envVal string
		if envKey == "_" {
			envVal = ""
		} else {
			envVal = os.Getenv(envKey)
		}

		if envVal == "" {
			if required && defaultVal == "" {
				return fmt.Errorf("%s: required environment variable %s not set", op, envKey)
			}
			if defaultVal != "" {
				envVal = defaultVal
			}
		}

		// Set the value by provided setter method if it's name is mentioned in the tag option "setter"
		if setterName != "" {
			setter := val.MethodByName(setterName)
			if !setter.IsValid() {
				return fmt.Errorf("%s: setter method '%s' for field '%s' not found", op, setterName, field.Name)
			}

			errs := setter.Call([]reflect.Value{reflect.ValueOf(envVal)})
			if len(errs) > 0 && !errs[0].IsNil() {
				return fmt.Errorf("%s: setter method '%s' for field '%s' failed: %v", op, setterName, field.Name, errs[0].Interface())
			}
			continue
		}

		// Check if the field is exported
		if !v.Field(i).CanSet() {
			return fmt.Errorf("%s: field %s is not exported", op, field.Name)
		}

		// Check if the field implements the Setter interface
		if v.Field(i).CanAddr() {
			set := v.Field(i).Addr().MethodByName(setterMethodName)
			if set.IsValid() {
				errs := set.Call([]reflect.Value{reflect.ValueOf(envVal)})
				if len(errs) > 0 && !errs[0].IsNil() {
					return fmt.Errorf("%s: failed to set value for field %s: %v", op, field.Name, errs[0].Interface())
				}
				continue
				//
				//if setter, ok := v.Field(i).Addr().Interface().(Setter); ok {
				//	if err := setter.Scan(envVal); err != nil {
				//		return fmt.Errorf("%s: failed to set value for field %s: %v", op, field.Name, err)
				//	}
				//	continue
				//}
			}
		}

		// Set the value based on the field type
		if envVal != "" {
			switch field.Type.Kind() {
			case reflect.String:
				v.Field(i).SetString(envVal)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
				vl, err := strconv.ParseInt(envVal, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: invalid int value for %s: %v", op, envKey, err)
				}
				v.Field(i).SetInt(vl)
			case reflect.Int64:
				if checkTimeDuration(field.Type) {
					dur, err := time.ParseDuration(envVal)
					if err != nil {
						return fmt.Errorf("%s: invalid time duration value for field \"%s\", env var \"%s\": %s, error: %v", op, field.Name, envKey, envVal, err)
					}
					v.Field(i).Set(reflect.ValueOf(dur))
					break
				}
				vl, err := strconv.ParseInt(envVal, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: invalid %s value for %s: %v", op, field.Type.Kind(), envKey, err)
				}
				v.Field(i).SetInt(vl)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				vl, err := strconv.ParseUint(envVal, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: invalid unsigned integer value for %s: %v", op, envKey, err)
				}
				v.Field(i).SetUint(vl)
			case reflect.Float32, reflect.Float64:
				vl, err := strconv.ParseFloat(envVal, 64)
				if err != nil {
					return fmt.Errorf("%s: invalid float value for %s: %v", op, envKey, err)
				}
				v.Field(i).SetFloat(vl)
			case reflect.Bool:
				val, err := strconv.ParseBool(envVal)
				if err != nil {
					return fmt.Errorf("%s: invalid boolean value for %s: %v", op, envKey, err)
				}
				v.Field(i).SetBool(val)
			case reflect.Slice:
				// If the field is a slice, split the value by comma and set the elements
				vals := strings.Split(envVal, ",")
				ln := len(vals)
				refSlice := reflect.MakeSlice(field.Type, 0, ln)

				// If Slice elements implement Setter interface then set the value
				if checkSliceElementsSetter(field.Type) {
					for _, vl := range vals {
						elem := reflect.New(field.Type.Elem()).Interface().(Setter)
						if err := elem.Scan(vl); err != nil {
							return fmt.Errorf("%s: failed to set value for field %s: %v", op, field.Name, err)
						}
						refSlice = reflect.Append(refSlice, reflect.ValueOf(elem).Elem())
					}
				} else {
					// If Slice elements are of basic types then set the value
					switch field.Type.Elem().Kind() {
					case reflect.String:
						refSlice = reflect.ValueOf(vals)
					case reflect.Int:
						for _, vl := range vals {
							intVal, err := strconv.ParseInt(vl, 10, 32)
							if err != nil {
								return fmt.Errorf("%s: invalid integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(int(intVal)))
						}
					case reflect.Int8:
						for _, vl := range vals {
							intVal, err := strconv.ParseInt(vl, 10, 8)
							if err != nil {
								return fmt.Errorf("%s: invalid integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(int8(intVal)))
						}
					case reflect.Int16:
						for _, vl := range vals {
							intVal, err := strconv.ParseInt(vl, 10, 16)
							if err != nil {
								return fmt.Errorf("%s: invalid integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(int16(intVal)))
						}
					case reflect.Int32:
						for _, vl := range vals {
							intVal, err := strconv.ParseInt(vl, 10, 32)
							if err != nil {
								return fmt.Errorf("%s: invalid integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(int32(intVal)))
						}
					case reflect.Int64:
						if checkTimeDuration(field.Type.Elem()) {
							for _, vl := range vals {
								dur, err := time.ParseDuration(vl)
								if err != nil {
									return fmt.Errorf("%s: invalid time duration value for %s: %v", op, envKey, err)
								}
								refSlice = reflect.Append(refSlice, reflect.ValueOf(dur))
							}
						} else {
							for _, vl := range vals {
								intVal, err := strconv.ParseInt(vl, 10, 64)
								if err != nil {
									return fmt.Errorf("%s: invalid integer value for %s: %v", op, envKey, err)
								}
								refSlice = reflect.Append(refSlice, reflect.ValueOf(intVal))
							}
						}
					case reflect.Uint:
						for _, vl := range vals {
							uintVal, err := strconv.ParseUint(vl, 10, 32)
							if err != nil {
								return fmt.Errorf("%s: invalid unsigned integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(uint(uintVal)))
						}
					case reflect.Uint8:
						for _, vl := range vals {
							uintVal, err := strconv.ParseUint(vl, 10, 8)
							if err != nil {
								return fmt.Errorf("%s: invalid unsigned integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(uint8(uintVal)))
						}
					case reflect.Uint16:
						for _, vl := range vals {
							uintVal, err := strconv.ParseUint(vl, 10, 16)
							if err != nil {
								return fmt.Errorf("%s: invalid unsigned integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(uint16(uintVal)))
						}
					case reflect.Uint32:
						for _, vl := range vals {
							uintVal, err := strconv.ParseUint(vl, 10, 32)
							if err != nil {
								return fmt.Errorf("%s: invalid unsigned integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(uint32(uintVal)))
						}
					case reflect.Uint64:
						for _, vl := range vals {
							uintVal, err := strconv.ParseUint(vl, 10, 64)
							if err != nil {
								return fmt.Errorf("%s: invalid unsigned integer value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(uintVal))
						}
					case reflect.Float32:
						for _, vl := range vals {
							floatVal, err := strconv.ParseFloat(vl, 32)
							if err != nil {
								return fmt.Errorf("%s: invalid float value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(float32(floatVal)))
						}
					case reflect.Float64:
						for _, vl := range vals {
							floatVal, err := strconv.ParseFloat(vl, 64)
							if err != nil {
								return fmt.Errorf("%s: invalid float value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(floatVal))
						}
					case reflect.Bool:
						for _, vl := range vals {
							boolVal, err := strconv.ParseBool(vl)
							if err != nil {
								return fmt.Errorf("%s: invalid boolean value for %s: %v", op, envKey, err)
							}
							refSlice = reflect.Append(refSlice, reflect.ValueOf(boolVal))
						}
					case reflect.Struct:
						if checkTime(field.Type.Elem()) {
							for _, vl := range vals {
								timeVal, err := time.Parse(time.RFC3339, vl)
								if err != nil {
									return fmt.Errorf("%s: invalid time value for %s: %v", op, envKey, err)
								}
								refSlice = reflect.Append(refSlice, reflect.ValueOf(timeVal))
							}
						} else {
							return fmt.Errorf("%s: unsupported struct slice type for field %s", op, field.Name)
						}
					default:
						return fmt.Errorf("%s: unsupported slice type for field %s", op, field.Name)
					}
				}
				v.Field(i).Set(refSlice)
			case reflect.Complex64, reflect.Complex128:
				val, err := strconv.ParseComplex(envVal, 128)
				if err != nil {
					return fmt.Errorf("%s: invalid complex value for %s: %v", op, envKey, err)
				}
				v.Field(i).SetComplex(val)
			case reflect.Struct:
				if checkTime(field.Type) {
					timeVal, err := time.Parse(time.RFC3339, envVal)
					if err != nil {
						return fmt.Errorf("%s: invalid time value for field \"%s\", env var \"%s\": %s, error: %v", op, field.Name, envKey, envVal, err)
					}
					v.Field(i).Set(reflect.ValueOf(timeVal))
				} else {
					return fmt.Errorf("%s: unsupported struct type for field %s", op, field.Name)
				}
			default:
				return fmt.Errorf("%s: unsupported type for field %s", op, field.Name)
			}
		}
	}
	return nil
}

func checkSliceElementsSetter(sliceType reflect.Type) bool {
	if sliceType.Kind() != reflect.Slice {
		return false
	}

	// Get the element type of the slice
	elemType := sliceType.Elem()

	// Get the Setter interface type
	setterType := reflect.TypeOf((*Setter)(nil)).Elem()

	// Check if the element type implements Setter
	return reflect.PointerTo(elemType).Implements(setterType)
}

func checkTimeDuration(fieldType reflect.Type) bool {
	return fieldType == reflect.TypeOf(time.Duration(0))
}

func checkTime(fieldType reflect.Type) bool {
	return fieldType == reflect.TypeOf(time.Time{})
}
