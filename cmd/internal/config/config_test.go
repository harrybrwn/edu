package config

import (
	"reflect"
	"testing"
)

func TestFind(t *testing.T) {
	type inner struct {
		One   string
		Two   string `config:"2"`
		Three int    `default:"333"`
	}
	type config struct {
		Inner inner
	}
	conf := &config{Inner: inner{One: "one", Two: "two"}}
	c := New(conf)
	_, val := findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "One"})
	if val.String() != "one" {
		t.Error("wrong value")
	}
	_, val = findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "2"})
	if val.String() != "two" {
		t.Error("wrong value")
	}
	_, val = findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "Three"})
	if val.Int() != 333 {
		t.Error("wrong value")
	}
}
