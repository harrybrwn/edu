package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/harrybrwn/go-canvas"
	table "github.com/olekukonko/tablewriter"
)

// Homedir will get the correct home directory
func Homedir() string {
	home := os.Getenv("HOME")
	if home == "" && runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return home
}

// Error is an error
type Error struct {
	Msg  string
	Code int
}

func (e *Error) Error() string {
	return e.Msg
}

// Mkdir creates a directory
func Mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0775)
	}
	return nil
}

// GetCourses gets all the courses
func GetCourses(all bool, opts ...canvas.Option) ([]*canvas.Course, error) {
	if !all {
		opts = append(opts, canvas.ActiveCourses)
	}
	courses, err := canvas.Courses(opts...)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// CoursesChan returns a channel of courses
func CoursesChan(all bool, opts ...canvas.Option) <-chan *canvas.Course {
	if !all {
		opts = append(opts, canvas.ActiveCourses)
	}
	return canvas.CoursesChan(opts...)
}

// FindCourse will find a course given a generice identifier
func FindCourse(identifier interface{}, opts ...canvas.Option) (*canvas.Course, error) {
	switch id := identifier.(type) {
	case int:
		return canvas.GetCourse(id, opts...)
	case string:
		i, err := strconv.Atoi(id)
		if err == nil {
			c, err := canvas.GetCourse(i, opts...)
			if err == nil {
				return c, nil
			}
		} else {
			i = -1
		}
		courses, err := GetCourses(false, opts...)
		if err != nil {
			return nil, err
		}
		for _, c := range courses {
			if c.Name == id {
				return c, nil
			} else if c.CourseCode == id {
				return c, nil
			} else if c.UUID == id {
				return c, nil
			} else if c.ID == i {
				return c, nil
			} else if c.SisCourseID == i {
				return c, nil
			}
		}
	}
	return nil, errors.New("could not find course")
}

var errAssignmentNotFound = errors.New("could not find assignment")

// FindAssignment will find an assignment the matches a generic identifier.
func FindAssignment(identifier string, all bool, opts ...canvas.Option) (*canvas.Assignment, error) {
	var (
		wg      sync.WaitGroup
		done    = make(chan struct{}, 1) // just make sure it never blocks
		ch      = make(chan *canvas.Assignment)
		idLower = strings.ToLower(identifier)
	)
	// stop all the goroutines when the function stops
	defer close(done)

	courses, err := GetCourses(all)
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(identifier)
	if err != nil {
		id = -1
	}

	wg.Add(len(courses))
	go func() {
		// close all the channels when the courses
		// loop is finished
		wg.Wait()
		close(ch)
	}()
	go func() {
		for _, c := range courses {
			go func(c *canvas.Course) {
				defer wg.Done()
				if c.AccessRestrictedByDate {
					return
				}
				assignments := c.Assignments(
					canvas.Opt("search_term", identifier),
					canvas.Opt("order_by", "name"),
					canvas.Opt("all_dates", all),
				)
				for {
					select {
					case <-done:
						return
					case as := <-assignments:
						if as == nil {
							break
						}
						if as.Name == identifier {
							ch <- as
						} else if strings.ToLower(as.Name) == idLower {
							ch <- as
						} else if as.QuizID == id { // be carfull here, if there is no quiz it will be 0
							ch <- as
						} else if strings.HasPrefix(as.Name, identifier) {
							ch <- as
						}
					}
				}

				// for as := range c.Assignments(
				// 	canvas.Opt("search_term", identifier),
				// 	canvas.Opt("order_by", "name"),
				// 	canvas.Opt("all_dates", all),
				// ) {
				// 	select {
				// 	case <-done:
				// 		return
				// 	default:
				// 	}
				// 	if as.Name == identifier {
				// 		ch <- as
				// 	} else if strings.ToLower(as.Name) == idLower {
				// 		ch <- as
				// 	} else if as.QuizID == id { // be carfull here, if there is no quiz it will be 0
				// 		ch <- as
				// 	} else if strings.HasPrefix(as.Name, identifier) {
				// 		ch <- as
				// 	}
				// }
			}(c)

			// remember we set `id` to -1 if identifier was not an int
			if id > 0 {
				ass, err := c.Assignment(id)
				if err == nil {
					ch <- ass
				}
			}
		}
	}()

	select {
	case a := <-ch:
		if a == nil {
			// if its nil then the channel is closed
			return nil, errAssignmentNotFound
		}
		return a, nil
	}
}

// GetAssignment will search courses for an assignment with the assignment's ID
func GetAssignment(id int, all bool, opts ...canvas.Option) (*canvas.Assignment, error) {
	ch := make(chan *canvas.Assignment)
	if !all {
		opts = append(opts, canvas.ActiveCourses, canvas.Opt("order_by", "date"))
	}
	for c := range canvas.CoursesChan(opts...) {
		go func(c *canvas.Course) {
			as, err := c.Assignment(id)
			if err == nil {
				ch <- as
			}
		}(c)
	}

	select {
	case as := <-ch:
		if as == nil {
			return nil, errAssignmentNotFound
		}
		return as, nil
	}
}

// NewTable creates a table with some default parameters
func NewTable(r io.Writer) *table.Table {
	t := table.NewWriter(r)
	t.SetBorder(false)
	t.SetColumnSeparator("")
	t.SetAlignment(table.ALIGN_LEFT)
	t.SetAutoFormatHeaders(false)
	t.SetHeaderLine(false)
	t.SetHeaderAlignment(table.ALIGN_LEFT)
	return t
}

// SetTableHeader sets the table header and automatically manages header color.
func SetTableHeader(t *table.Table, header []string, color bool) {
	t.SetHeader(header)
	if color {
		headercolors := make([]table.Colors, len(header))
		for i := range header {
			headercolors[i] = table.Colors{table.FgCyanColor}
		}
		t.SetHeaderColor(headercolors...)
	}
}

// HandleAuthErr will handle a canvas auth error
// and give a more relevant error message.
func HandleAuthErr(err error) error {
	autherr, ok := err.(*canvas.AuthError)
	// i'm so sorry for string comparison error handling i know its bad
	if ok && autherr.Errors[0].Message == "Invalid access token." {
		return fmt.Errorf("%w (set 'token' in config file or '$CANVAS_TOKEN' env variable)", autherr)
	}
	return err
}
