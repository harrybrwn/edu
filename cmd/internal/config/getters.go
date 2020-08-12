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

// IsEmpty returns true if the value stored at some
// key is a zero value or an empty value
func IsEmpty(key string) bool {
	return c.IsEmpty(key)
}

// IsEmpty returns true if the value stored at some
// key is a zero value or an empty value
func (c *Config) IsEmpty(key string) bool {
	val, err := c.get(key)
	if err != nil {
		return true
	}
	if !val.IsValid() {
		return false
	}
	return val.IsZero()
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
	val, err := find(c.elem, keys)
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
	ret, ok := val.Interface().([]int)
	if !ok {
		return nil
	}
	return ret
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
	ret, ok := res.Interface().([]int64)
	if !ok {
		return nil
	}
	return ret
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

func find(val reflect.Value, keyPath []string) (reflect.Value, error) {
	var err error
	typ := val.Type()
	n := typ.NumField()
	for i := 0; i < n; i++ {
		typFld := typ.Field(i)
		// if the first key is the same as the fieldname
		if isCorrectLabel(keyPath[0], typFld) {
			value := val.Field(i)
			if len(keyPath) > 1 {
				return find(value, keyPath[1:])
			}
			if !isZero(value) {
				// if the field has been set then we return it
				return value, nil
			}

			defvalue, err := getDefaultValue(&typFld, &value)
			switch err {
			case errNoDefaultValue:
				return value, nil
			case nil:
				return defvalue, nil
			default: // err != nil
				return defvalue, err
			}
		}
	}
	if err == nil {
		err = ErrFieldNotFound
	}
	return nilval, err
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

func setDefaults(val reflect.Value) (err error) {
	var seterr error
	typ := val.Type()
	n := typ.NumField()
	for i := 0; i < n; i++ {
		fldVal := val.Field(i)  // field's value
		fldType := typ.Field(i) // field's type
		// make recursive calls
		if fldVal.Kind() == reflect.Struct {
			return setDefaults(fldVal)
		}
		// if the field has been set already, then
		// it is a significant value to the user
		// do not override with defaults
		if !isZero(fldVal) {
			continue
		}

		defval, err := getDefaultValue(&fldType, &fldVal)
		switch err {
		case errNoDefaultValue:
			continue
		case nil:
			break
		default:
			return err
		}
		if fldVal.CanSet() {
			fldVal.Set(defval)
		} else {
			if seterr == nil {
				seterr = fmt.Errorf("cannot set value for field '%s'", fldType.Name)
			}
			continue
		}
	}
	return seterr
}

var errNoDefaultValue = errors.New("no default value found")

func getDefaultValue(fld *reflect.StructField, fldval *reflect.Value) (def reflect.Value, err error) {
	val := fld.Tag.Get("default")
	env := fld.Tag.Get("env")
	if env != "" {
		val = os.Getenv(env)
	}
	if val == "" {
		return nilval, errNoDefaultValue
	}

	var (
		ival  int64
		uival uint64
		fval  float64
	)
	switch fld.Type.Kind() {
	case reflect.String:
		return reflect.ValueOf(val), nil
	case reflect.Int:
		ival, err = strconv.ParseInt(val, 10, 64)
		def = reflect.ValueOf(int(ival))
	case reflect.Int8:
		ival, err = strconv.ParseInt(val, 10, 8)
		def = reflect.ValueOf(int8(ival))
	case reflect.Int16:
		ival, err = strconv.ParseInt(val, 10, 16)
		def = reflect.ValueOf(int16(ival))
	case reflect.Int32:
		ival, err = strconv.ParseInt(val, 10, 32)
		def = reflect.ValueOf(int32(ival))
	case reflect.Int64:
		ival, err = strconv.ParseInt(val, 10, 64)
		def = reflect.ValueOf(int64(ival))
	case reflect.Uint:
		uival, err = strconv.ParseUint(val, 10, 64)
		def = reflect.ValueOf(uint(uival))
	case reflect.Uint8:
		uival, err = strconv.ParseUint(val, 10, 8)
		def = reflect.ValueOf(uint8(uival))
	case reflect.Uint16:
		uival, err = strconv.ParseUint(val, 10, 16)
		def = reflect.ValueOf(uint16(uival))
	case reflect.Uint32:
		uival, err = strconv.ParseUint(val, 10, 32)
		def = reflect.ValueOf(uint32(uival))
	case reflect.Uint64:
		uival, err = strconv.ParseUint(val, 10, 64)
		def = reflect.ValueOf(uival)
	case reflect.Float32:
		fval, err = strconv.ParseFloat(val, 32)
		def = reflect.ValueOf(float32(fval))
	case reflect.Float64:
		fval, err = strconv.ParseFloat(val, 64)
		def = reflect.ValueOf(fval)
	case reflect.Bool:
		var bval bool
		bval, err = strconv.ParseBool(val)
		def = reflect.ValueOf(bval)
	case reflect.Slice:
		// TODO: figure out how to detect a []byte
		switch fldval.Interface().(type) {
		case []byte:
			def = reflect.ValueOf([]byte(val))
		default:
			panic("don't know how to parse these yet")
		}
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Func:
	default:
		return nilval, errors.New("unknown default config type")
	}
	if err != nil {
		return nilval, fmt.Errorf("could not parse default value: %v", err)
	}
	return def, err
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

func set(obj interface{}, key string, val interface{}) error {
	objval := reflect.ValueOf(obj).Elem()
	field, err := find(objval, strings.Split(key, "."))
	if err != nil {
		return err
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
