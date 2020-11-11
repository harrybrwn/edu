package ucm

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/harrybrwn/edu/school"
	"github.com/harrybrwn/errs"
	"golang.org/x/net/html"
)

const selector = "div.pagebodydiv table.datadisplaytable tr"

var terms = map[string]string{
	"spring": "10",
	"summer": "20",
	"fall":   "30",
}

// ScheduleConfig holds options for getting
// the UC Merced schedule
type ScheduleConfig struct {
	Year    int
	Term    string
	Subject string
	Open    bool
}

// Schedule is a map of courses by course CRN
type Schedule map[int]*Course

// NewSchedule will return a new schedule based on the config.
func NewSchedule(config ScheduleConfig) (Schedule, error) {
	sched, err := getSchedule(config.Year, config.Term, config.Subject, config.Open)
	if err != nil {
		return nil, err
	}
	return sched, nil
}

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

// Len returns the number of courses in the schedule
func (s *Schedule) Len() int {
	return len(*s)
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
	CRN int
	// comprised of the subject code and course number
	Fullcode string

	// Course subject code
	Subject string
	// course number
	Number int
	// Lab or Discussion section. If the object is
	// a lecture then this should be 01
	Section string
	Title   string

	Exam     *Exam
	Units    int
	Activity string
	Days     []time.Weekday
	Time     struct {
		Start, End time.Time
	}
	BuildingRoom string
	StartEnd     string
	Date         struct {
		Start, End time.Time
	}
	Instructor string

	MaxEnrolled    int
	ActiveEnrolled int

	timeStr string
	seats   string
	order   int
}

// Exam is a course exam
type Exam struct {
	Day      time.Weekday
	Building string
	Date     time.Time
	Time     struct {
		Start, End time.Time
	}
}

// ID returns the course's crn
func (c *Course) ID() int {
	return c.CRN
}

// Name returns the courses title
func (c *Course) Name() string {
	return fmt.Sprintf("%s %s", c.Fullcode, c.Title)
}

// CourseNumber returns the number used to semantically
// identify the course
func (c *Course) CourseNumber() int {
	return c.Number
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
	return getSchedule(year, term, "", open)
}

// BySubject gets the schedule and only one subject given a subject code.
func BySubject(year int, term, subject string, open bool) (Schedule, error) {
	return getSchedule(year, term, subject, open)
}

var (
	errNotACourse   = errors.New("not a course")
	errPrevNotFound = errors.New("crn not found in previous html element")
)

// bySubject gets the schedule and only one subject given a subject code.
func getSchedule(year int, term, subject string, open bool) (Schedule, error) {
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
	const selector = "div.pagebodydiv table.datadisplaytable tr"
	var (
		selection        = doc.Find(selector)
		schedule         = Schedule{}
		keys             = make([]string, 13)
		keyerr, parseErr error
		order            = 0
	)

	selection.Each(func(i int, s *goquery.Selection) {
		header := s.Find("th.ddlabel p small")
		if header.Length() != 0 {
			keys = make([]string, 13)
			for j, n := range header.Nodes {
				keys[j] = strings.Replace(n.FirstChild.Data, " ", "", -1)
			}
			if len(keys) != 13 {
				keyerr = errs.New("the wrong number of columns were found in the document")
			}
			return
		}

		var (
			val     string
			ss      *goquery.Selection
			courses = s.Find("td.dddefault small")
			values  = make([]string, 0, 13)
		)

		// Get each row value
		for _, n := range courses.Nodes {
			ss = &goquery.Selection{Nodes: []*html.Node{n}}
			val = strings.Trim(ss.Text(), "\n \t\u00a0")
			values = append(values, val)
		}
		// clean out empty values
		for i := 0; i < len(values); i++ {
			val = values[i]
			if val == "\u00a0" || val == "" {
				continue
			} else {
				values = values[i:]
				break
			}
		}

		// Handle short rows, this happens occationally when
		// a course has two different times
		switch values[0] {
		case "EXAM":
			var (
				lect *Course
				ok   bool
			)
			// Find the crn for the previous row
			// which is the course this exam belongs to
			lectCRN, err := getPrevCRN(s)
			if err == errPrevNotFound {
				return
			} else if err != nil {
				goto HandleErrExam
			}
			lect, ok = schedule[int(lectCRN)]
			if !ok {
				err = errors.New("could not find lecture for this exam")
				goto HandleErrExam
			}
			lect.Exam, err = parseExam(values, year)
			if err != nil {
				goto HandleErrExam
			}
		HandleErrExam:
			if err != nil && parseErr == nil {
				parseErr = err
			} else if err != nil {
				log.Println("schedule: Internal error:", err)
			}
			return
		case "LECT":
			return // TODO figure out what to do with the second lecture time
		case "LAB":
			return // TODO figure out what to do with the second lab time
		}

		course, err := newCourse(values, year)
		if err == errNotACourse {
			panic(err)
		} else if err != nil {
			if parseErr == nil {
				parseErr = err
			}
			return
		}
		course.order = order
		schedule[course.CRN] = course
		order++
	})
	if keyerr != nil {
		return nil, keyerr
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return schedule, nil
}

func newCourse(data []string, year int) (*Course, error) {
	if len(data) != 13 {
		return nil, errNotACourse
	}
	crn, err := strconv.Atoi(data[0])
	if err != nil {
		err = fmt.Errorf("could not parse crn: %w", err)
		return nil, err
	}
	units, err := strconv.Atoi(data[3])
	if err != nil {
		err = fmt.Errorf("could not parse units: %w", err)
		return nil, err
	}
	maxenrl, err := strconv.Atoi(data[10])
	if err != nil {
		err = fmt.Errorf("could not parse max enrollment: %w", err)
		maxenrl = 0
		return nil, err
	}
	activenrl, err := strconv.Atoi(data[11])
	if err != nil {
		err = fmt.Errorf("could not parse active enrollment: %w", err)
		activenrl = 0
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
	date, err := parseDateRange(data[8], year)
	if err != nil {
		return nil, err
	}
	c.Date = *date
	c.Time.Start, c.Time.End, err = parseTime(timeStr)
	if err != nil {
		return nil, err
	}
	// parsing the course number from the course code
	parts := strings.Split(c.Fullcode, "-")
	if len(parts) >= 3 {
		// Fullcode has the form: <Subject>-<Number>-<Section>
		c.Subject = parts[0]
		end := parts[1][len(parts[1])-1]
		if end < '0' || end > '9' {
			parts[1] = trimInt(parts[1])
		}
		c.Number, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("could not parse course number: %w", err)
		}
		c.Section = parts[2]
	}
	return c, nil
}

func parseExam(values []string, year int) (*Exam, error) {
	var (
		err error
	)
	exam := &Exam{
		Day:      listDays(values[1])[0],
		Building: values[3],
	}
	exam.Time.Start, exam.Time.End, err = parseTime(values[2])
	if err != nil {
		return nil, err
	}
	date, err := parseDateRange(values[4], year)
	if err != nil {
		return nil, err
	}
	exam.Date = date.Start
	return exam, nil
}

func getPrevCRN(s *goquery.Selection) (int, error) {
	prev := s.Prev()
	inner := prev.Find("td p a")
	if inner.Text() == "" {
		// return // TODO some lectures have two sections, so the previous
		inner = prev.Prev().Find("td p a")
		if inner.Text() == "" {
			return 0, errPrevNotFound
		}
	}
	crn, err := strconv.ParseInt(inner.Text(), 10, 32)
	if err != nil {
		return 0, err
	}
	return int(crn), nil
}

func (c *Course) setDate(dateRange string, year int) (err error) {
	dates := strings.Split(dateRange, " ")
	if len(dates) != 2 {
		return errors.New("unexpected date format")
	}
	format := "02-Jan"
	dates[0] = strings.ToTitle(dates[0])
	dates[1] = strings.ToTitle(dates[1])

	c.Date.Start, err = time.Parse(format, dates[0])
	if err != nil {
		return err
	}
	c.Date.End, err = time.Parse(format, dates[1])
	if err != nil {
		return err
	}
	c.Date.Start = c.Date.Start.AddDate(year, 0, 0)
	c.Date.End = c.Date.End.AddDate(year, 0, 0)
	return nil
}

type dateRange struct {
	Start, End time.Time
}

func parseDateRange(dateString string, year int) (*dateRange, error) {
	dates := strings.Split(dateString, " ")
	if len(dates) != 2 {
		return nil, errors.New("unexpected date format")
	}
	format := "02-Jan"
	dates[0] = strings.ToTitle(dates[0])
	dates[1] = strings.ToTitle(dates[1])

	var (
		date dateRange
		err  error
	)
	date.Start, err = time.Parse(format, dates[0])
	if err != nil {
		return nil, err
	}
	date.End, err = time.Parse(format, dates[1])
	if err != nil {
		return nil, err
	}
	date.Start = date.Start.AddDate(year, 0, 0)
	date.End = date.End.AddDate(year, 0, 0)
	return &date, nil
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
