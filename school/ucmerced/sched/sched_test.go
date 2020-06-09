package sched

import (
	"fmt"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	sch, err := Get(2020, "spring", false)
	if err != nil {
		t.Error(err)
	}
	for crn, course := range sch {
		if crn == 0 {
			t.Error("should not have a crn of zero")
		}
		if course.CRN == 0 {
			t.Error("should not have a crn of zero")
		}
		if crn != course.CRN {
			t.Error("key does not match value")
		}
		if course.seats == "Closed" {
			if course.SeatsAvailible() != 0 {
				t.Error("should be zero")
			}
		} else if course.SeatsAvailible() == 0 {
			t.Error("should not be zero")
		}
	}
	sch, err = BySubject(2020, "spring", "cse", false)
	if err != nil {
		t.Error(err)
	}
	for _, course := range sch {
		if !strings.HasPrefix(course.Number, "CSE") {
			t.Error("should be a cse course")
		}
	}
}

func TestSched_Err(t *testing.T) {
	_, err := Get(2020, "", true)
	if err == nil {
		t.Error("expected an error for a bad term")
	}
	_, err = Get(1850, "spring", false)
	if err == nil {
		t.Error("expeted an error for a rediculous year")
	}
}

func Test(t *testing.T) {
	s, err := BySubject(2020, "summer", "cse", false)
	if err != nil {
		t.Error(err)
	}
	for _, c := range s {
		fmt.Println(c.Title)
	}
}
