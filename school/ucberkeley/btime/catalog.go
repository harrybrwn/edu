// Package btime is a library for the berkeleytime api. See https://www.berkeleytime.com/landing.
package btime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/harrybrwn/edu/school"
)

const (
	catalogURL = "https://www.berkeleytime.com/api/catalog/catalog_json/filters/"
	filterURL  = ""
)

// New creates a new catalog
func New() (*Catalog, error) {
	c := &Catalog{}
	resp, err := http.Get(catalogURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Catalog is a json struct for a catalog
type Catalog struct {
	Level       Items `json:"level"`
	Haas        Items `json:"haas"`
	University  Items `json:"university"`
	Engineering Items `json:"engineering"`
	Department  Items `json:"department"`
	Ls          Items `json:"ls"`
	Semester    Items `json:"semester"`
	Units       Items `json:"units"`

	// TODO: find out what type these should be
	Time       []interface{} `json:"time"`
	Length     []interface{} `json:"length"`
	Chemistry  []interface{} `json:"chemistry"`
	Enrollment []interface{} `json:"enrollment"`

	DefaultPlaylists string `json:"default_playlists"`
	DefaultCourse    string `json:"default_course"`

	courses map[int]*Result
}

// AllItems returns a slice of all of the items in the catalog.
func (c *Catalog) AllItems() []Item {
	length := len(c.Level) +
		len(c.Haas) +
		len(c.University) +
		len(c.Engineering) +
		len(c.Department) +
		len(c.Ls) +
		len(c.Semester) +
		len(c.Units)
	items := make([]Item, 0, length) // avoid reallocations
	items = append(items, c.Level...)
	items = append(items, c.Haas...)
	items = append(items, c.University...)
	items = append(items, c.Engineering...)
	items = append(items, c.Department...)
	items = append(items, c.Ls...)
	items = append(items, c.Semester...)
	items = append(items, c.Units...)
	return items
}

// Get will get a course given an id.
func (c *Catalog) Get(id int) school.Course {
	if c.courses == nil {
		return nil
	}
	course, ok := c.courses[id]
	if !ok {
		return nil
	}
	return course
}

// Items is a slice of Item structs
type Items []Item

// Search the list of items given a search term.
func (its Items) Search(term string) *Item {
	term = strings.ToLower(term)
	for _, itm := range its {
		if strings.Contains(strings.ToLower(itm.Name), term) {
			return &itm
		}
	}
	return nil
}

// Item is an item on the catalog
type Item struct {
	Name     string `json:"name"`
	Semester string `json:"semester"`
	Category string `json:"category"`
	Year     string `json:"year"`
	ID       int    `json:"id"`
}

// Results is a slice of result structs
type Results []Result

// Result is the result from the filter endpoint.
type Result struct {
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	ResultID           int     `json:"id"`
	Units              string  `json:"units"`
	OpenSeats          int     `json:"open_seats"`
	Abbreviation       string  `json:"abbreviation"`
	EnrolledPercentage float64 `json:"enrolled_percentage"`
	CourseNumber       string  `json:"course_number"`
	FavoriteCount      int     `json:"favorite_count"`
	Waitlisted         int     `json:"waitlisted"`
	Enrolled           int     `json:"enrolled"`

	GradeAverage  float64 `json:"grade_average"`
	LetterAverage string  `json:"letter_average"`
}

// ID returns the id
func (r *Result) ID() int {
	return r.ResultID
}

// Name returns the courses Title
func (r *Result) Name() string {
	return r.Title
}

// Course will get the course associated with the filter result.
func (r *Result) Course() (*Course, error) {
	req := &http.Request{
		Method: "GET",
		Proto:  "HTTP/1.1",
		URL: &url.URL{
			Scheme:   "https",
			Host:     "www.berkeleytime.com",
			Path:     "/api/catalog/catalog_json/course_box/",
			RawQuery: fmt.Sprintf("course_id=%d", r.ResultID),
		},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	course := &Course{}
	if err = json.NewDecoder(resp.Body).Decode(course); err != nil {
		return nil, err
	}
	return course, nil
}

// SeatsOpen returns the OpenSeats field and is here for
// interface implimentation
func (r *Result) SeatsOpen() int {
	return r.OpenSeats
}

// DefaultFilter makes a filter request to the catalog's default
// filter parameters
func (c *Catalog) DefaultFilter() (Results, error) {
	opts := strings.Split(c.DefaultPlaylists, ",")
	return sendFilter(opts)
}

// Courses returns a slice of Results in a generic
// Course interface format. Returns nil on error.
func (c *Catalog) Courses() []school.Course {
	results, err := c.DefaultFilter()
	if err != nil {
		return nil
	}
	courses := make([]school.Course, len(results))
	for i, r := range results {
		c.courses[r.ResultID] = &r
		courses[i] = &r
	}
	return courses
}

// Filter will return the results from a filter request given
// filter option IDs.
func Filter(opts ...interface{}) (Results, error) {
	filter := make([]string, len(opts))
	for i, o := range opts {
		filter[i] = fmt.Sprintf("%v", o)
	}
	return sendFilter(filter)
}

func sendFilter(filter []string) (Results, error) {
	req := &http.Request{
		Method: "GET",
		Proto:  "HTTP/1.1",
		URL: &url.URL{
			Scheme: "https",
			Host:   "www.berkeleytime.com",
			Path:   "/api/catalog/filter/",
			RawQuery: (&url.Values{
				"filters": filter,
			}).Encode(),
		},
		Header: make(http.Header),
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res := make([]Result, 0)
	return res, json.NewDecoder(resp.Body).Decode(&res)
}

var (
	_ school.Schedule = (*Catalog)(nil)
	_ school.Course   = (*Result)(nil)
)
