package internal

import (
	"fmt"
	"io"
	"os"

	"github.com/harrybrwn/go-canvas"
	table "github.com/olekukonko/tablewriter"
)

func Stop(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", msg)
	os.Exit(1)
}

func Mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0775)
	}
	return nil
}

func GetCourses(all bool) ([]*canvas.Course, error) {
	courses, err := canvas.Courses(canvas.ActiveCourses)
	if err != nil {
		return nil, err
	}
	if all {
		completed, err := canvas.Courses(canvas.CompletedCourses)
		if err != nil {
			return courses, err
		}
		courses = append(courses, completed...)
		pending, err := canvas.Courses(canvas.InvitedOrPendingCourses)
		if err != nil {
			return courses, err
		}
		courses = append(courses, pending...)
	}
	return courses, nil
}

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

func SetTableHeader(t *table.Table, header []string) {
	headercolors := make([]table.Colors, len(header))
	for i := range header {
		headercolors[i] = table.Colors{table.FgCyanColor}
	}
	t.SetHeader(header)
	t.SetHeaderColor(headercolors...)
}
