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
		Nums  []int
	}
	conf := &config{
		Inner: inner{One: "one", Two: "two"},
		Nums:  []int{1, 2, 3, 4, 5},
	}
	c := New(conf)
	_, _, val := findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "One"})
	if val.String() != "one" {
		t.Error("wrong value")
	}
	_, _, val = findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "2"})
	if val.String() != "two" {
		t.Error("wrong value")
	}
	_, _, val = findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "Three"})
	if val.Int() != 333 {
		t.Errorf("got %d, want %d\n", val.Int(), 333)
	}
	conf.Inner.Three = 123
	_, _, val = findKey(reflect.ValueOf(c.config).Elem(), []string{"Inner", "Three"})
	if val.Int() != 123 {
		t.Errorf("got %d, want %d", val.Int(), 123)
	}
	nums := c.GetIntSlice("Nums")
	if len(nums) != 5 {
		t.Error("wrong length")
	}
	for i := range nums {
		if nums[i] != i+1 {
			t.Error("wrong int value")
		}
	}
	empty := c.GetString("empty")
	if empty != "" {
		t.Errorf("a non-existant key should give an empty string: got %s", empty)
	}
}
