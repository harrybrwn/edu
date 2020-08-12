package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

var (
	// ErrNoConfigFile is returned when the config file cannot be found.
	ErrNoConfigFile = errors.New("no config file")
	// ErrNoConfigDir is returned when all of the possible config paths
	// do not exist.
	ErrNoConfigDir = errors.New("no config directory")
	// ErrFieldNotFound is returned when the field of a struct
	// was not found with reflection
	ErrFieldNotFound = errors.New("could not find struct field")
	// ErrWrongType is returned when the wrong type is used
	ErrWrongType = errors.New("wrong type")

	c      *Config
	nilval = reflect.ValueOf(nil)
)

func init() { c = &Config{} }

// New creates a new config object from a configuration
// struct.
func New(conf interface{}) *Config {
	cfg := &Config{}
	cfg.SetConfig(conf)
	return cfg
}

// Config holds configuration metadata
type Config struct {
	file      string
	paths     []string
	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error

	// Actual config data
	config interface{}
	elem   reflect.Value
}

// SetConfig will set the config struct
func SetConfig(conf interface{}) error { return c.SetConfig(conf) }

// SetConfig will set the config struct
func (c *Config) SetConfig(conf interface{}) error {
	c.config = conf
	c.elem = reflect.ValueOf(conf).Elem()
	return nil
}

// GetConfig will return the the config struct that has been
// set by the user but as an interface type.
func GetConfig() interface{} { return c.GetConfig() }

// GetConfig will return the the config struct that has been
// set by the user but as an interface type.
func (c *Config) GetConfig() interface{} {
	return c.config
}

// AddDefaultDirs sets the config and home directories as possible
// config dir options.
//
// If the config dir is found (see os.UserConfigDir) then
// <config dir>/<name> is added to the list of possible config paths.
// If the home dir is found (see os.UserHomeDir) then
// <home dir>/<name> is added to the list of possible config paths.
func AddDefaultDirs(name string) { c.UseDefaultDirs(name) }

// UseDefaultDirs sets the config and home directories as possible
// config dir options.
//
// If the config dir is found (see os.UserConfigDir) then
// "<config dir>/<name>" is added to the list of possible config paths.
// If the home dir is found (see os.UserHomeDir) then
// "<home dir>/.<name>" is added to the list of possible config paths.
//
// These paths are added the list (its different for windows)
//	$XDG_CONFIG_DIR/<name>
//	$HOME/.<name>
func (c *Config) UseDefaultDirs(name string) {
	configDir, err := os.UserConfigDir()
	if err == nil {
		c.paths = append(c.paths, filepath.Join(configDir, name))
	}
	home, err := os.UserHomeDir()
	if err == nil {
		c.paths = append(c.paths, filepath.Join(home, "."+name))
	}
}

// AddConfigDir will add a config dir using the user config dir
// (see os.UserConfigDir) and join it with the name given.
//	$XDG_CONFIG_DIR/<name>
func AddConfigDir(name string) error { return c.AddConfigDir(name) }

// AddConfigDir will add a config dir using the user config dir
// (see os.UserConfigDir) and join it with the name given.
//	$XDG_CONFIG_DIR/<name>
func (c *Config) AddConfigDir(name string) error {
	dir, err := os.UserConfigDir()
	if err == nil {
		c.paths = append(c.paths, filepath.Join(dir, name))
	}
	return err
}

// AddHomeDir will add a config dir using the user home dir
// (see os.UserHomeDir) and join it with the name given and a "."
//	$HOME/.<name>
func AddHomeDir(name string) error { return c.AddHomeDir(name) }

// AddHomeDir will add a config dir using the user home dir
// (see os.UserHomeDir) and join it with the name given and a "."
//	$HOME/.<name>
func (c *Config) AddHomeDir(name string) error {
	dir, err := os.UserHomeDir()
	if err == nil {
		c.paths = append(c.paths, filepath.Join(dir, "."+name))
	}
	return err
}

// SetType will set the file type of config being used.
func SetType(ext string) error { return c.SetType(ext) }

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
func ReadConfigFile() error { return c.ReadConfigFile() }

// ReadConfigFile will read in the config file
func (c *Config) ReadConfigFile() error {
	filename := c.FileUsed()
	if filename == "" {
		return ErrNoConfigDir
	}
	stat, err := os.Stat(filename)
	if os.IsNotExist(err) || stat.IsDir() {
		return ErrNoConfigFile
	} else if err != nil {
		return err
	}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return c.unmarshal(raw, c.config)
}

// FileUsed will return the file used for
// configuration.
func FileUsed() string { return c.FileUsed() }

// FileUsed will return the file used for
// configuration.
func (c *Config) FileUsed() string {
	var err error
	for _, path := range c.paths {
		if _, err = os.Stat(path); !os.IsNotExist(err) {
			return filepath.Join(path, c.file)
		}
	}
	return ""
}

// DirUsed returns the path of the first existing
// config directory.
func DirUsed() string { return c.DirUsed() }

// DirUsed returns the path of the first existing
// config directory. If none of the paths exist, then
// The first non-empty path will be returned.
func (c *Config) DirUsed() string {
	var (
		err  error
		path string
	)
	for _, path = range c.paths {
		if _, err = os.Stat(path); !os.IsNotExist(err) {
			return path
		}
	}
	for _, path = range c.paths {
		if path != "" {
			return path
		}
	}
	return ""
}

// SetFilename sets the config filename.
func SetFilename(name string) { c.SetFilename(name) }

// SetFilename sets the config filename.
func (c *Config) SetFilename(name string) {
	c.file = name
}

// AddPath will add a path the the list of possible
// configuration folders
func AddPath(path string) { c.AddPath(path) }

// AddPath will add a path the the list of possible
// configuration folders
func (c *Config) AddPath(path string) {
	c.paths = append(c.paths, os.ExpandEnv(path))
}
