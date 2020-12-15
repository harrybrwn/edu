package ucm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
)

const (
	testyear = 2021
	testterm = "spring"
)

var (
	mu   sync.Mutex
	once sync.Once
	buf  bytes.Buffer
)

func getTestData(t *testing.T) io.Reader {
	t.Helper()
	b := bytes.Buffer{}
	once.Do(func() {
		resp, err := getData(fmt.Sprintf("%d", testyear), testterm, "", false)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if _, err = buf.ReadFrom(resp.Body); err != nil {
			t.Fatal(err)
		}
	})
	mu.Lock()
	b.ReadFrom(&buf)
	mu.Unlock()
	return &b
}

func testSchedule(t *testing.T) Schedule {
	t.Helper()
	// note: uses data from spring 2021
	r := getTestData(t)
	rows, err := parseRows(r)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := parse(rows, 2021)
	if err != nil {
		t.Fatal(err)
	}
	return sc
}

func Test(t *testing.T) {
}

func TestParser(t *testing.T) {
	r := getTestData(t)
	rows, err := parseRows(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("did not parse any rows")
	}
	sc, err := parse(rows, 2021)
	if err != nil {
		t.Error(err)
	}

	for _, row := range rows {
		if row.crn == 0 {
			continue
		}
		if c, ok := sc[row.crn]; ok {
			if c.Exam == nil {
				continue
				// } else if c.Activity != "LECT" && c.CRN == 10163 {
				// 	fmt.Println(c.CRN, c.Activity)
			}
		} else {
			t.Error("should be a course here")
		}
		if row.kind == kindExam {
			t.Error("should have found this exam")
		}
	}
	sch := testSchedule(t)
	for crn, course := range sch {
		c, ok := sc[crn]
		if !ok {
			t.Errorf("%d should be in both schedules", crn)
		}
		if !cmp(c, course) {
			t.Error("courses should be the same")
		}
	}
}

func TestGet(t *testing.T) {
	sch := testSchedule(t)
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
	sch, err := BySubject(2020, "spring", "cse", false)
	if err != nil {
		t.Error(err)
	}
}

func TestSched_Err(t *testing.T) {
	_, err := Get(testyear, "", true)
	if err == nil {
		t.Error("expected an error for a bad term")
	}
	_, err = Get(1850, "spring", false)
	if err == nil {
		t.Error("expeted an error for a rediculous year")
	}
}

func TestGetFall(t *testing.T) {
	t.Skip("this test doen't actuall test anything... fix it")
	sc, err := Get(testyear, "fall", true)
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

func TestScratch(t *testing.T) {
	t.Skip("don't need this test")
	// https://crossenrollcourses.universityofcalifornia.edu/?home_campus=7&term_year=1-2020&subject_area=9&with_prev&pageSize=10
	url := "http://registrar.ucmerced.edu/go/schedule"
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			fmt.Println(r.URL)
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var b bytes.Buffer
	if _, err = b.ReadFrom(resp.Body); err != nil {
		t.Error(err)
	}
	fmt.Println(b.String())
}

func cmp(a, b *Course) bool {
	return a.CRN == b.CRN &&
		a.Fullcode == b.Fullcode &&
		a.Subject == b.Subject &&
		a.Number == b.Number &&
		a.Section == b.Section &&
		a.Title == b.Title &&
		a.Units == b.Units &&
		a.Activity == b.Activity &&
		a.BuildingRoom == b.BuildingRoom &&
		a.Instructor == b.Instructor &&
		a.Capacity == b.Capacity &&
		a.Enrolled == b.Enrolled &&
		a.seats == b.seats
}
