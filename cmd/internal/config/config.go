package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

var defaultConfig *Config

func init() {
	defaultConfig = &Config{}
}

// New creates a new config object from a configuration
// struct.
func New(conf interface{}) *Config {
	return &Config{
		config: conf,
	}
}

// Config holds configuration metadata
type Config struct {
	file      string
	paths     []string
	config    interface{}
	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error
}

// SetStruct will set the config struct
func SetStruct(conf interface{}) {
	defaultConfig.SetStruct(conf)
}

// SetStruct will set the config struct
func (c *Config) SetStruct(conf interface{}) {
	c.config = conf
}

// SetType will set the file type of config being used.
func SetType(ext string) error {
	return defaultConfig.SetType(ext)
}

// SetType will set the file type of config being used.
func (c *Config) SetType(ext string) error {
	switch ext {
	case "yaml", "yml":
		c.marshal = yaml.Marshal
		c.unmarshal = yaml.Unmarshal
	case "json":
		c.marshal = json.Marshal
		c.unmarshal = json.Unmarshal
	default:
		return fmt.Errorf("unknown config type %s", ext)
	}
	return nil
}

// ReadConfigFile will read in the config file
func ReadConfigFile() error {
	return defaultConfig.ReadConfigFile()
}

// ReadConfigFile will read in the config file
func (c *Config) ReadConfigFile() error {
	filename := c.FileUsed()
	if filename == "" {
		return errors.New("no config file to read")
	}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return c.unmarshal(raw, c.config)
}

// FileUsed will return the file used for
// configuration.
func FileUsed() string {
	return defaultConfig.FileUsed()
}

// FileUsed will return the file used for
// configuration.
func (c *Config) FileUsed() string {
	for _, path := range c.paths {
		configFile := filepath.Join(path, c.file)
		if _, err := os.Stat(configFile); !os.IsNotExist(err) {
			return configFile
		}
	}
	return ""
}

// SetFilename sets the config filename.
func SetFilename(name string) {
	defaultConfig.SetFilename(name)
}

// SetFilename sets the config filename.
func (c *Config) SetFilename(name string) {
	c.file = name
}

// AddPath will add a path the the list of possible
// configuration folders
func AddPath(path string) {
	defaultConfig.AddPath(path)
}

// AddPath will add a path the the list of possible
// configuration folders
func (c *Config) AddPath(path string) {
	c.paths = append(c.paths, os.ExpandEnv(path))
}

// Get will get a variable by key
func Get(key string) interface{} {
	return defaultConfig.Get(key)
}

// Get will get a variable by key
func (c *Config) Get(key string) interface{} {
	keys := strings.Split(key, ".")
	_, val := findKey(reflect.ValueOf(c.config).Elem(), keys)
	return val.Interface()
}

// GetString will get the config value by name and
// return it as a string
func GetString(key string) string {
	return defaultConfig.GetString(key)
}

// GetString will get the config value by name and
// return it as a string
func (c *Config) GetString(key string) string {
	keys := strings.Split(key, ".")
	_, val := findKey(reflect.ValueOf(c.config).Elem(), keys)
	return val.String()
}

func findKey(val reflect.Value, keyPath []string) (*reflect.StructField, reflect.Value) {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		typFld := typ.Field(i)
		if isCorrectLabel(keyPath[0], typFld) {
			if deflt := typFld.Tag.Get("default"); deflt != "" {
				return &typFld, typedDefaultValue(&typFld, deflt)
			}
			if len(keyPath) > 1 {
				structField, value := findKey(val.Field(i), keyPath[1:])
				return structField, value
			}
			return &typFld, val.Field(i)
		}
	}
	return nil, reflect.ValueOf(nil)
}

func isCorrectLabel(key string, field reflect.StructField) bool {
	return key == field.Name ||
		key == field.Tag.Get("config") ||
		key == field.Tag.Get("yaml") ||
		key == field.Tag.Get("json")
}

func typedDefaultValue(fld *reflect.StructField, val string) reflect.Value {
	switch fld.Type.Kind() {
	case reflect.String:
		return reflect.ValueOf(val)
	case reflect.Int:
		ival, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		return reflect.ValueOf(ival)
	case reflect.Bool:
		bval, err := strconv.ParseBool(val)
		if err != nil {
			panic(err)
		}
		return reflect.ValueOf(bval)
	}
	return reflect.ValueOf(nil)
}
