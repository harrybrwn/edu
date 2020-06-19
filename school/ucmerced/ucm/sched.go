package ucm

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/harrybrwn/edu/school"
	"github.com/harrybrwn/errs"
)

const selector = "div.pagebodydiv table.datadisplaytable tr"

var terms = map[string]string{
	"spring": "10",
	"summer": "20",
	"fall":   "30",
}

// Schedule is a map of courses by course CRN
type Schedule map[int]*Course

// Get will get a course given the course id
func (s *Schedule) Get(id int) school.Course {
	c, ok := (*s)[id]
	if !ok {
		return nil
	}
	return c
}

// Courses returns a list of courses as a course interface.
func (s *Schedule) Courses() []school.Course {
	courses := make([]school.Course, len(*s))
	for _, c := range *s {
		courses[c.order] = c
	}
	return courses
}

// Ordered will return a slice of courses that preserves
// the original order.
func (s *Schedule) Ordered() []*Course {
	list := make([]*Course, len(*s))
	for _, c := range *s {
		list[c.order] = c
	}
	return list
}

// Course holds data for a specific
// course that has been parsed from the courses table.
type Course struct {
	CRN      int
	Fullcode string
	Number   int

	Title    string
	Units    int
	Activity string
	Days     []time.Weekday
	Time     struct {
		Start, End time.Time
	}
	BuildingRoom string
	StartEnd     string
	Instructor   string

	MaxEnrolled    int
	ActiveEnrolled int

	timeStr string
	seats   string
	order   int
}

// ID returns the course's crn
func (c *Course) ID() int {
	return c.CRN
}

// Name returns the courses title
func (c *Course) Name() string {
	return c.Title
}

// SeatsOpen gets the number of seats available for the course.
func (c *Course) SeatsOpen() int {
	seats, err := strconv.Atoi(c.seats)
	if err != nil {
		// if it is anything but a number
		// then there are no seats available
		return 0
	}
	return seats
}

// Get gets the schedule
func Get(year int, term string, open bool) (Schedule, error) {
	return BySubject(year, term, "", open)
}

// BySubject gets the schedule and only one subject given a subject code.
func BySubject(year int, term, subject string, open bool) (Schedule, error) {
	resp, err := getData(fmt.Sprintf("%d", year), term, strings.ToUpper(subject), open)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errs.New(resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	selection := doc.Find(selector)
	schedule := Schedule{}
	keys := make([]string, 13)
	var (
		keyerr error
		order  = 0
	)

	selection.Each(func(i int, s *goquery.Selection) {
		header := s.Find("th.ddlabel p small")
		if header.Length() != 0 {
			keys = make([]string, 13)
			for i, n := range header.Nodes {
				keys[i] = strings.Replace(n.FirstChild.Data, " ", "", -1)
			}
			if len(keys) != 13 {
				keyerr = errs.New("the wrong number of columns were found in the document")
			}
			return
		}
		courses := s.Find("td.dddefault")

		values := make([]string, 0, 13)
		courses.Each(func(k int, ss *goquery.Selection) {
			values = append(values, strings.Trim(ss.Text(), "\n \t"))
		})
		course, err := newCourse(values)
		if err != nil && len(values) == 12 {
			// if we found an error and there are
			// only 12 values we skip this edge case
			// has to do with a few classes having
			// two different locations
			return
		}
		if err != nil {
			panic(err)
		}
		course.order = order
		schedule[course.CRN] = course
		order++
	})
	return schedule, keyerr
}

func newCourse(data []string) (*Course, error) {
	if len(data) != 13 {
		return nil, fmt.Errorf("cannot create a course with %d values", len(data))
	}
	crn, e1 := strconv.Atoi(data[0])
	units, e2 := strconv.Atoi(data[3])
	maxenrl, e3 := strconv.Atoi(data[10])
	activenrl, e4 := strconv.Atoi(data[11])
	err := errs.Chain(e1, e2, e3, e4)
	if err != nil {
		return nil, err
	}
	timeStr := data[6]
	c := &Course{
		CRN:            crn,
		Fullcode:       data[1],
		Title:          data[2],
		Units:          units,
		Activity:       data[4],
		Days:           listDays(data[5]),
		BuildingRoom:   data[7],
		StartEnd:       data[8],
		Instructor:     data[9],
		MaxEnrolled:    maxenrl,
		ActiveEnrolled: activenrl,
		seats:          data[12],
	}
	c.Time.Start, c.Time.End, err = parseTime(timeStr)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(c.Fullcode, "-")
	if len(parts) >= 2 {
		end := parts[1][len(parts[1])-1]
		if end < '0' || end > '9' {
			parts[1] = trimInt(parts[1])
		}
		c.Number, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func parseTime(ts string) (start time.Time, end time.Time, err error) {
	if ts == "TBD-TBD" {
		return
	}
	split := strings.Split(ts, "-")
	if len(split) < 2 {
		return start, end, errors.New("invalid time format")
	}
	start, e1 := time.Parse("3:04", split[0])
	end, e2 := time.Parse("3:04pm", split[1])
	if err = errs.Pair(e1, e2); err != nil {
		return
	}

	diff := end.Hour() - start.Hour()
	if end.Hour() >= 12 && diff >= 12 {
		start = start.Add(12 * time.Hour)
	}
	return
}

func trimInt(s string) string {
	num := make([]rune, 0, len(s))
	for _, c := range s {
		if c >= '0' && c <= '9' {
			num = append(num, c)
		}
	}
	return string(num)
}

var dayMap = map[rune]time.Weekday{
	'M': time.Monday,
	'T': time.Tuesday,
	'W': time.Wednesday,
	'R': time.Thursday,
	'F': time.Friday,
}

func listDays(daystr string) (days []time.Weekday) {
	days = make([]time.Weekday, len(daystr))
	for i, char := range daystr {
		days[i] = dayMap[char]
	}
	return days
}

var client http.Client

func getData(year, term, subject string, openclasses bool) (*http.Response, error) {
	termcode, ok := terms[term]
	if !ok {
		return nil, fmt.Errorf("could not find term %s", term)
	}
	var open string
	if openclasses {
		open = "Y"
	} else {
		open = "N"
	}
	if subject == "" {
		subject = "ALL"
	}
	params := &url.Values{
		"validterm":   {fmt.Sprintf("%s%s", year, termcode)},
		"openclasses": {open},
		"subjcode":    {strings.ToUpper(subject)},
	}
	req := &http.Request{
		Method: "GET",
		Proto:  "HTTP/1.1",
		URL: &url.URL{
			Scheme:   "https",
			Host:     "mystudentrecord.ucmerced.edu",
			Path:     "/pls/PROD/xhwschedule.P_ViewSchedule",
			RawQuery: params.Encode(),
		},
	}
	return client.Do(req)
}

var (
	_ school.Schedule = (*Schedule)(nil)
	_ school.Course   = (*Course)(nil)
)
