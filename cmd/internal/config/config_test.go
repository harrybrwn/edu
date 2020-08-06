package config

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func cleanup() {
	c = &Config{}
}

func TestPaths(t *testing.T) {
	defer cleanup()
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(os.TempDir(), ".config"))
	os.Setenv("AppData", os.TempDir())
	os.Setenv("HOME", os.TempDir())
	os.Setenv("USERPROFILE", os.TempDir())
	os.Setenv("home", os.TempDir())

	type C struct{}
	t.Run("WithHome", func(t *testing.T) {
		defer cleanup()
		SetConfig(&C{})
		os.Setenv("HOME", os.TempDir())
		os.Setenv("USERPROFILE", os.TempDir())
		os.Setenv("home", os.TempDir())
		if err := AddHomeDir("config_test"); err != nil {
			t.Error(err)
		}
		if c.paths[0] != filepath.Join(os.TempDir(), ".config_test") {
			t.Error("home dir not set as a path")
		}
	})

	t.Run("WithConfig", func(t *testing.T) {
		defer cleanup()
		SetConfig(&C{})
		os.Setenv("XDG_CONFIG_HOME", filepath.Join(os.TempDir(), ".config"))
		os.Setenv("AppData", os.TempDir())
		if err := AddConfigDir("config_test"); err != nil {
			t.Error(err)
		}
		var exp string
		switch runtime.GOOS {
		case "windows":
			exp = os.TempDir()
		case "darwin":
			exp = filepath.Join(os.TempDir(), "/Library/Application Support")
		case "plan9":
			exp = filepath.Join(os.TempDir(), "lib")
		default:
			exp = filepath.Join(os.TempDir(), ".config")
		}
		exp = filepath.Join(exp, "config_test")
		if c.paths[0] != exp {
			t.Errorf("expected %s; got %s", exp, c.paths[0])
		}
		c.paths = []string{}
		AddDefaultDirs("config_test")
		if c.paths[0] != exp {
			t.Error("home dir not set as a path")
		}
	})
	SetConfig(&C{})
	AddPath("$HOME")
	if c.paths[0] != os.TempDir() {
		t.Error("AddPath did set the wrong path")
	}
}

func TestFileTypes(t *testing.T) {
	defer cleanup()
	matchFn := func(n string, i interface{}) {
		name, err := url.QueryUnescape(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name())
		if err != nil {
			t.Error(err)
		}
		if name != n {
			t.Errorf("wrong function name: want %s; got %s", n, name)
		}
	}
	err := SetType("invalid")
	if err == nil {
		t.Error("expected an invalid filetype error")
	}
	if err = SetType("yml"); err != nil {
		t.Error(err)
	}
	matchFn("gopkg.in/yaml.v2.Unmarshal", c.unmarshal)
	matchFn("gopkg.in/yaml.v2.Marshal", c.marshal)
	if err = SetType("yaml"); err != nil {
		t.Error(err)
	}
	matchFn("gopkg.in/yaml.v2.Unmarshal", c.unmarshal)
	matchFn("gopkg.in/yaml.v2.Marshal", c.marshal)
	if err = SetType("json"); err != nil {
		t.Error("err")
	}
	matchFn("encoding/json.Unmarshal", c.unmarshal)
	matchFn("encoding/json.Marshal", c.marshal)
}

func TestReadConfig_Err(t *testing.T) {
	type C struct {
		Val int `config:"val"`
	}
	defer cleanup()
	check := func(e error) {
		if e != nil {
			t.Error(e)
		}
	}
	var err error
	SetConfig(&C{})
	dir := filepath.Join(
		os.TempDir(), fmt.Sprintf("config_test.%s_%d_%d",
			t.Name(), os.Getpid(), time.Now().UnixNano()))
	AddPath(dir)
	if dirused := DirUsed(); dirused != "" {
		t.Error("expected empty dir because config paths do not exist")
	}
	if err = ReadConfigFile(); err != ErrNoConfigDir {
		t.Error("should return the 'no config dir' error")
	}

	check(os.MkdirAll(dir, 0700))
	defer os.RemoveAll(dir)

	if dirused := DirUsed(); dirused != dir {
		t.Errorf("wrong DirUsed: got %s; want %s", dirused, dir)
	}
	SetFilename("config")
	check(SetType("yml"))
	err = ReadConfigFile()
	if err != ErrNoConfigFile {
		t.Error("exected the 'no config file' error")
	}
	fake := filepath.Join(dir, "config")
	check(os.Mkdir(fake, 0700))
	err = ReadConfigFile()
	if err == nil {
		t.Error("expected an error while reading the config")
	}
	check(os.Remove(fake))
	check(ioutil.WriteFile(fake, []byte(`val: 10`), 0600))
	check(ReadConfigFile())
	if FileUsed() != fake {
		t.Error("files should be the same")
	}
}

func TestGet(t *testing.T) {
	defer cleanup()
	type C struct {
		S string `config:"a-string"`
	}
	conf := &C{"this is a test"}
	cfg := New(conf)
	cfg.SetFilename("test.txt")
	ires := cfg.Get("a-string")
	s, ok := ires.(string)
	if !ok {
		t.Error("should have returned a string")
	}
	if s != "this is a test" {
		t.Errorf("expected %s; got %s", conf.S, s)
	}
	if _, err := cfg.GetErr("a-string"); err != nil {
		t.Error(err)
	}
	// testing the panic in Config.get
	c = &Config{}
	defer func() {
		r := recover()
		if r != errElemNotSet {
			t.Error("should have paniced with errElemNotSet")
		}
	}()
	x := Get("a-string")
	if x != nil {
		t.Error("should be nil")
	}
}

func TestGet_Err(t *testing.T) {
	defer cleanup()
	type C struct {
		NotASlice int
	}
	key := "not-here"
	conf := &C{5}
	SetConfig(conf)
	if GetConfig() != conf {
		t.Error("wrong struct pointer")
	}
	if HasKey(key) {
		t.Error("config struct should not have this key")
	}
	if Get(key) != nil {
		t.Error("expected a nil value")
	}
	if _, err := GetErr(key); err == nil {
		t.Error("expected an error")
	}
	if GetInt(key) > 0 {
		t.Error("invalid key should be an invalid value")
	}
	if _, err := GetIntErr(key); err == nil {
		t.Error("expected an error")
	}
	if GetString(key) != "" {
		t.Error("nonexistant key should give an empty string")
	}
	if _, err := GetStringErr(key); err == nil {
		t.Error("expected an error")
	}
	if GetBool(key) {
		t.Error("config struct should not have this key")
	}
	if GetIntSlice(key) != nil {
		t.Error("nonexistant key should give nil int slice")
	}
	if GetInt64Slice(key) != nil {
		t.Error("nonexistant key should give nil int64 slice")
	}
	if GetFloat(key) != 0.0 {
		t.Error("nonexistant key should give a zero value")
	}
	if GetFloat32(key) != 0.0 {
		t.Error("nonexistant key should give a zero value")
	}

	if GetInt("NotASlice") != 5 {
		t.Error("dummy check failed for GetInt")
	}
	if GetIntSlice("NotASlice") != nil {
		t.Error("should return nil for non-slice fields")
	}
	if GetInt64Slice("NotASlice") != nil {
		t.Error("should return nil for non-slice fields")
	}
}

func TestDefaults(t *testing.T) {
	defer cleanup()
	type C struct {
		A  string  `config:"a" env:"TEST_A"`
		B  int     `config:"b" default:"89"`
		TF bool    `config:"truefalse" default:"true"`
		F  float64 `config:"f" env:"PI"`
		F2 float32 `config:"f2" default:"1.3"`
	}
	conf := &C{}
	SetConfig(conf)
	os.Setenv("TEST_A", "testing-value")
	os.Setenv("PI", strconv.FormatFloat(math.Pi, 'f', 15, 64))

	if !HasKey("a") {
		t.Error("key 'a' should exist")
	}
	if GetString("a") != "testing-value" {
		t.Error("environment default gave the wrong value")
	}
	if GetInt("b") != 89 {
		t.Error("`default` tag gave the wrong default value")
	}
	if GetBool("truefalse") == false || GetBool("TF") == false {
		t.Error("wrong default boolean value")
	}
	if v, err := GetBoolErr("truefalse"); err != nil || v == false {
		t.Error("wrong value or error:", err)
	}
	if GetFloat("f") != math.Pi {
		t.Error("got wrong float default")
	}
	if GetFloat32("f2") != 1.3 {
		t.Error("wrong defalt float32 value")
	}

	conf.A = "yeet"
	if GetString("a") == "testing-value" {
		t.Error("default string value should have been overridden")
	}
	conf.F = math.E
	if GetFloat("f") == math.Pi {
		t.Error("default float64 value should have been overridden")
	}
	conf.F2 = 5.9
	if GetFloat32("f2") == 1.3 {
		t.Error("default float32 value should have been overridden")
	}
}

func TestDefaults_Err(t *testing.T) {
	defer cleanup()
	type C struct {
		A  string  `config:"a" env:"TEST_A"`
		B  int     `config:"b" default:"x"`
		TF bool    `config:"truefalse" default:"8"`
		F  float64 `config:"f" env:"PI"`
		F2 float32 `config:"f2" default:"what am i even doing"`
	}
	conf := &C{}
	SetConfig(conf)
	os.Setenv("PI", "not a number")

	if _, err := GetIntErr("b"); err == nil {
		t.Error("expected an error")
	}
	if GetInt("b") != 0 {
		t.Error("should not be anything but 0")
	}
	if GetFloat("f") != 0.0 {
		t.Error("default should not be a valid number")
	}
	if GetFloat("f2") != 0 {
		t.Error("default should not be a valid number")
	}
	if _, err := GetBoolErr("truefalse"); err == nil {
		t.Error("expected an error")
	}
}

func TestGetMap(t *testing.T) {
	defer cleanup()
	type C struct {
		M      map[string]string `config:"map"`
		Notmap int               `config:"not-map"`
	}
	SetConfig(&C{M: map[string]string{"one": "1", "two": "2"}})
	m := GetStringMap("map")
	if m["one"] != "1" {
		t.Error("wrong map result")
	}
	if m["two"] != "2" {
		t.Error("wrong map result")
	}
	m = GetStringMap("not-map")
	if m != nil {
		t.Error("a non-map should be nil")
	}
	m = GetStringMap("not_here")
	if m != nil {
		t.Error("non-existant key should be nil")
	}
}

func TestSlices(t *testing.T) {
	type inner struct {
		Ints []int64 `config:"inner-ints"`
	}
	type C struct {
		Ints  []int `config:"ints"`
		Inner inner `config:"inner"`
	}
	obj := &C{
		Ints:  []int{1, 2, 3, 4, 5},
		Inner: inner{[]int64{1, 2, 3, 4, 5}},
	}
	SetConfig(obj)
	ints := GetIntSlice("ints")
	expi := 5
	if len(ints) != expi {
		t.Errorf("expected length %d, got length: %d", expi, len(ints))
		return
	}
	for i := range ints {
		if ints[i] != obj.Ints[i] {
			t.Errorf("expected %d; got %d", ints[i], obj.Ints[i])
		}
	}
	if !HasKey("Inner.inner-ints") {
		t.Error("key should exist")
	}
	int64s := GetInt64Slice("Inner.inner-ints")
	if len(int64s) != expi {
		t.Errorf("expected length %d, got length: %d", expi, len(int64s))
		return
	}
	for i := range int64s {
		if int64s[i] != obj.Inner.Ints[i] {
			t.Errorf("expected %d; got %d", ints[i], obj.Inner.Ints[i])
		}
	}
}

func TestSet(t *testing.T) {
	defer cleanup()
	type C struct {
		I int        `config:"i"`
		C complex128 `config:"c"`
	}
	conf := &C{I: 5, C: 5.5i}
	SetConfig(conf)
	if GetInt("i") != 5 {
		t.Error("wrong value")
	}
	if Get("C").(complex128) != 5.5i {
		t.Error("has wrong value")
	}
	if err := set(conf, "i", 10); err != nil {
		t.Error(err)
	}
	if conf.I != 10 {
		t.Error("set did not set the value")
	}
	if err := set(conf, "c", 99.99i); err != nil {
		t.Error(err)
	}
	if conf.C != 99.99i {
		t.Error("did not set the correct value")
	}
}
