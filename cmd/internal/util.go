package internal

import (
	"io"
	"os"

	"github.com/harrybrwn/go-canvas"
	table "github.com/olekukonko/tablewriter"
)

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
	o := append(opts, canvas.ActiveCourses)
	courses, err := canvas.Courses(o...)
	if err != nil {
		return nil, err
	}
	if all {
		o = append(opts, canvas.CompletedCourses)
		completed, err := canvas.Courses(o...)
		if err != nil {
			return courses, err
		}
		courses = append(courses, completed...)
		o = append(opts, canvas.InvitedOrPendingCourses)
		pending, err := canvas.Courses(o...)
		if err != nil {
			return courses, err
		}
		courses = append(courses, pending...)
	}
	return courses, nil
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
func SetTableHeader(t *table.Table, header []string) {
	headercolors := make([]table.Colors, len(header))
	for i := range header {
		headercolors[i] = table.Colors{table.FgCyanColor}
	}
	t.SetHeader(header)
	t.SetHeaderColor(headercolors...)
}
