package btime

import (
	"fmt"
	"testing"
)

var testingCatalog *Catalog

func testCatalog() Catalog {
	var err error
	if testingCatalog == nil {
		testingCatalog, err = New()
		if err != nil {
			panic(err)
		}
	}
	return *testingCatalog
}

func Test(t *testing.T) {
	schedule := testCatalog()
	all := schedule.AllItems()
	if len(all) < 1 {
		t.Error("should not be empty")
	}
	res, err := schedule.DefaultFilter()
	if err != nil {
		t.Error(err)
	}
	r := res[0]
	course, err := r.Course()
	if err != nil {
		t.Error(err)
	}
	if course == nil {
		t.Error("got nil course")
	}
}

func TestSchedual(t *testing.T) {
	cat := testCatalog()
	for _, u := range cat.Semester {
		fmt.Println(u)
	}
}
