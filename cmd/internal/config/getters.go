package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var errElemNotSet = errors.New("Config.elem not set, use config.New() or config.SetStruct()")

// HasKey tests if the config struct has a key given
func HasKey(key string) bool { return c.HasKey(key) }

// HasKey tests if the config struct has a key given
func (c *Config) HasKey(key string) bool {
	return hasKey(c.elem, strings.Split(key, "."))
}

// Get will get a variable by key
func Get(key string) interface{} { return c.Get(key) }

// Get will get a variable by key
func (c *Config) Get(key string) interface{} {
	val, err := c.get(key)
	if err != nil {
		return nil
	}
	return val.Interface()
}

// GetErr will get the value stored at some key and return an error
// if something went wrong.
func GetErr(key string) (interface{}, error) { return c.GetErr(key) }

// GetErr will get the value stored at some key and return an error
// if something went wrong.
func (c *Config) GetErr(key string) (interface{}, error) {
	val, err := c.get(key)
	if err != nil {
		return nil, err
	}
	return val.Interface(), nil
}

func (c *Config) get(key string) (reflect.Value, error) {
	if c.elem.Kind() == reflect.Invalid {
		panic(errElemNotSet)
	}
	keys := strings.Split(key, ".")
	_, _, val, err := find(c.elem, keys)
	return val, err
}

// GetString will get the config value by name and
// return it as a string
func GetString(key string) string { return c.GetString(key) }

// GetString will get the config value by name and
// return it as a string. This function will also expand
// any environment variables in the value returned.
func (c *Config) GetString(key string) string {
	s, _ := c.GetStringErr(key)
	return s
}

// GetStringErr is the same as get string but it returns an error
// when something went wrong, mainly if the key does not exist
func GetStringErr(key string) (string, error) {
	return c.GetStringErr(key)
}

// GetStringErr is the same as get string but it returns an error
// when something went wrong, mainly if the key does not exist
func (c *Config) GetStringErr(key string) (string, error) {
	val, err := c.get(key)
	if err != nil {
		return "", err
	}
	return os.ExpandEnv(val.String()), nil
}

// GetInt will get the int value of a key
func GetInt(key string) int { return c.GetInt(key) }

// GetInt will get the int value of a key
func (c *Config) GetInt(key string) int {
	i, _ := c.GetIntErr(key)
	return i
}

// GetIntErr will return an get an int but also return an error
// if something went wrong, main just missing keys and conversion errors
func GetIntErr(key string) (int, error) { return c.GetIntErr(key) }

// GetIntErr will return an get an int but also return an error
// if something went wrong, main just missing keys and conversion errors
func (c *Config) GetIntErr(key string) (int, error) {
	val, err := c.get(key)
	if err != nil {
		return 0, err
	}
	return int(val.Int()), nil
}

// GetFloat will get a float64 value
func GetFloat(key string) float64 { return c.GetFloat(key) }

// GetFloat will get a float64 value
func (c *Config) GetFloat(key string) float64 {
	val, err := c.get(key)
	if err != nil {
		return 0.0
	}
	return val.Float()
}

// GetFloat32 will get a float32 value
func GetFloat32(key string) float32 { return c.GetFloat32(key) }

// GetFloat32 will get a float32 value
func (c *Config) GetFloat32(key string) float32 {
	val, err := c.get(key)
	if err != nil {
		return 0.0
	}
	return float32(val.Float())
}

// GetBool will get the boolean value at the given key
func GetBool(key string) bool { return c.GetBool(key) }

// GetBool will get the boolean value at the given key
func (c *Config) GetBool(key string) bool {
	val, err := c.get(key)
	if err != nil {
		return false
	}
	return val.Bool()
}

// GetBoolErr will get a boolean value but return an error
// is something went wrong.
func GetBoolErr(key string) (bool, error) {
	return c.GetBoolErr(key)
}

// GetBoolErr will get a boolean value but return an error
// is something went wrong.
func (c *Config) GetBoolErr(key string) (bool, error) {
	val, err := c.get(key)
	if err != nil {
		return false, err
	}
	return val.Bool(), nil
}

// GetIntSlice will get a slice of ints from a key
func GetIntSlice(key string) []int { return c.GetIntSlice(key) }

// GetIntSlice will get a slice of ints from a key
//
// Warning: will panic if the key does not reference
// a []int
func (c *Config) GetIntSlice(key string) []int {
	val, err := c.get(key)
	if err != nil {
		return nil
	}
	if val.Kind() != reflect.Slice {
		return nil
	}
	return val.Interface().([]int)
}

// GetInt64Slice will return a slice of int64.
//
// Warning: will panic if the key given does not
// reference a []int64
func GetInt64Slice(key string) []int64 { return c.GetInt64Slice(key) }

// GetInt64Slice will return a slice of int64.
//
// Warning: will panic if the key given does not
// reference a []int64
func (c *Config) GetInt64Slice(key string) []int64 {
	res, err := c.get(key)
	if err != nil {
		return nil
	}
	if res.Kind() != reflect.Slice {
		return nil
	}
	return res.Interface().([]int64)
}

// GetStringMap will get a map of string keys to string values
func GetStringMap(key string) map[string]string {
	return c.GetStringMap(key)
}

// GetStringMap will get a map of string keys to string values
func (c *Config) GetStringMap(key string) map[string]string {
	res, err := c.get(key)
	if err != nil {
		return nil
	}
	if res.Kind() != reflect.Map {
		return nil
	}
	m := make(map[string]string)
	iter := res.MapRange()
	for iter.Next() {
		m[iter.Key().String()] = iter.Value().String()
	}
	return m
}

func find(val reflect.Value, keyPath []string) (bool, *reflect.StructField, reflect.Value, error) {
	typ := val.Type()
	n := typ.NumField()
	for i := 0; i < n; i++ {
		typFld := typ.Field(i)
		if isCorrectLabel(keyPath[0], typFld) {
			value := val.Field(i)
			if len(keyPath) > 1 {
				return find(value, keyPath[1:])
			}
			if isZero(value) {
				// priority goes to env variables
				env := typFld.Tag.Get("env")
				deflt := typFld.Tag.Get("default")
				var err error
				if env != "" {
					value, err = typedDefaultValue(&typFld, os.Getenv(env))
				} else if deflt != "" {
					value, err = typedDefaultValue(&typFld, deflt)
				}
				if err != nil {
					return false, &typFld, nilval, err
				}
			}
			return true, &typFld, value, nil
		}
	}
	return false, nil, nilval, ErrFieldNotFound
}

func hasKey(val reflect.Value, keyPath []string) bool {
	typ := val.Type()
	n := typ.NumField()
	for i := 0; i < n; i++ {
		typFld := typ.Field(i)
		if isCorrectLabel(keyPath[0], typFld) {
			if len(keyPath) == 1 {
				return true
			}
			return hasKey(val.Field(i), keyPath[1:])
		}
	}
	return false
}

func isCorrectLabel(key string, field reflect.StructField) bool {
	return key == field.Name ||
		key == field.Tag.Get("config") ||
		key == field.Tag.Get("yaml") ||
		key == field.Tag.Get("json")
}

func isZero(val reflect.Value) bool {
	return reflect.DeepEqual(
		val.Interface(),
		reflect.Zero(val.Type()).Interface(),
	)
}

func typedDefaultValue(fld *reflect.StructField, val string) (reflect.Value, error) {
	switch fld.Type.Kind() {
	case reflect.String:
		return reflect.ValueOf(val), nil
	case reflect.Int:
		ival, err := strconv.Atoi(val)
		if err != nil {
			return nilval, fmt.Errorf("could not parse default value: %v", err)
		}
		return reflect.ValueOf(ival), nil
	case reflect.Uint:
		uival, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nilval, fmt.Errorf("could not parse default value: %v", err)
		}
		return reflect.ValueOf(uival), nil
	case reflect.Float32:
		fval, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return nilval, fmt.Errorf("could not parse default value: %v", err)
		}
		return reflect.ValueOf(float32(fval)), nil // ParseFloat always returs float64
	case reflect.Float64:
		fval, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nilval, fmt.Errorf("could not parse default value: %v", err)
		}
		return reflect.ValueOf(fval), nil
	case reflect.Bool:
		bval, err := strconv.ParseBool(val)
		if err != nil {
			return nilval, fmt.Errorf("could not parse default value: %v", err)
		}
		return reflect.ValueOf(bval), nil
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Slice:
	case reflect.Func:
	}
	return nilval, nil
}

func set(obj interface{}, key string, val interface{}) error {
	objval := reflect.ValueOf(obj).Elem()
	ok, _, field, err := find(objval, strings.Split(key, "."))
	if err != nil {
		return err
	}
	if !ok {
		return ErrFieldNotFound
	}
	if !field.CanSet() {
		return errors.New("cannot set value")
	}

	var exptype reflect.Kind
	switch v := val.(type) {
	case string:
		exptype = reflect.String
		field.SetString(v)
	case []byte:
		exptype = reflect.Slice
		field.SetBytes(v)
	case bool:
		exptype = reflect.Bool
		field.SetBool(v)
	case complex64:
		exptype = reflect.Complex64
		field.SetComplex(complex128(v))
	case complex128:
		exptype = reflect.Complex128
		field.SetComplex(v)
	case int:
		exptype = reflect.Int
		field.SetInt(int64(v))
	case int8:
		exptype = reflect.Int8
		field.SetInt(int64(v))
	case int32:
		exptype = reflect.Int32
		field.SetInt(int64(v))
	case int64:
		exptype = reflect.Int64
		field.SetInt(int64(v))
	case float32:
		exptype = reflect.Float32
		field.SetFloat(float64(v))
	case float64:
		exptype = reflect.Float64
		field.SetFloat(v)
	default:
		field.Set(reflect.ValueOf(val))
		return nil
	}
	if field.Kind() != exptype {
		return ErrWrongType
	}
	return nil
}
