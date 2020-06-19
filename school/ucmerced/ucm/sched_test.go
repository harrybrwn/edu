package ucm

import (
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
			if course.SeatsOpen() != 0 {
				t.Error("should be zero")
			}
		} else if course.SeatsOpen() == 0 {
			t.Error("should not be zero")
		}
	}
	sch, err = BySubject(2020, "spring", "cse", false)
	if err != nil {
		t.Error(err)
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
	sc, err := Get(2020, "fall", true)
	if err != nil {
		t.Error(err)
	}
	for _, c := range sc {
		if c.Days[0] == 194 {
			continue
		}
		// fmt.Println(c.Days)
	}
}

func TestParseTime(t *testing.T) {
	testcases := []struct {
		str        string
		start, end int
	}{
		{str: "11:30-2:15pm", start: 11, end: 14},
		{str: "10:30-1:20pm", start: 10, end: 13},
		{str: "5:30-7:20pm", start: 17, end: 19},
		{str: "9:30-11:20am", start: 9, end: 11},
		{str: "9:00-12:00am", start: 9, end: 0},
		{str: "TBD-TBD", start: 0, end: 0},
	}
	for _, tc := range testcases {
		start, end, err := parseTime(tc.str)
		if err != nil {
			t.Error(err)
		}
		if start.Hour() != tc.start {
			t.Errorf("wrong starting hour %d; want %d", start.Hour(), tc.start)
		}
		if end.Hour() != tc.end {
			t.Errorf("wrong ending hour %d; want %d", end.Hour(), tc.end)
		}
	}
}
