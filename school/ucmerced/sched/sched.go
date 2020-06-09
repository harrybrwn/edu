package sched

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/harrybrwn/errs"
)

const selector = "div.pagebodydiv table.datadisplaytable tr"

var terms = map[string]string{
	"spring": "10",
	"summer": "20",
	"fall":   "30",
}

// Schedual is a map of courses by course CRN
type Schedual map[int]*Course

// Ordered will return a slice of courses that preserves
// the original order.
func (s *Schedual) Ordered() []*Course {
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
	Number   string

	Title        string
	Units        int
	Activity     string
	Days         string
	Time         string
	BuildingRoom string
	StartEnd     string
	Instructor   string

	MaxEnrolled    int
	ActiveEnrolled int

	seats string
	order int
}

// SeatsAvailible gets the number of seats availible for the course.
func (c *Course) SeatsAvailible() int {
	seats, err := strconv.Atoi(c.seats)
	if err != nil {
		return 0
	}
	return seats
}

// Get gets the schedual
func Get(year int, term string, open bool) (Schedual, error) {
	return BySubject(year, term, "", open)
}

// BySubject gets the schedual and only one subject given a subject code.
func BySubject(year int, term, subject string, open bool) (Schedual, error) {
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
	schedual := Schedual{}
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
		course.order = order
		schedual[course.CRN] = course
		order++
	})
	return schedual, keyerr
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
	c := &Course{
		CRN:            crn,
		Fullcode:       data[1],
		Title:          data[2],
		Units:          units,
		Activity:       data[4],
		Days:           data[5],
		Time:           data[6],
		BuildingRoom:   data[7],
		StartEnd:       data[8],
		Instructor:     data[9],
		MaxEnrolled:    maxenrl,
		ActiveEnrolled: activenrl,
		seats:          data[12],
	}
	parts := strings.Split(c.Fullcode, "-")
	if len(parts) >= 2 {
		c.Number = parts[1]
	}
	return c, nil
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
	params := &url.Values{
		"validterm":   {fmt.Sprintf("%s%s", year, termcode)},
		"openclasses": {open},
		// "subjcode":    {"ALL"},
		"subjcode": {subject},
	}
	if openclasses {
		params.Set("openclasses", "Y")
	}
	if subject != "" {
		params.Set("subjcode", strings.ToUpper(subject))
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
