package print

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/harrybrwn/edu/pkg/term"
	"github.com/harrybrwn/go-canvas"
	"github.com/olekukonko/tablewriter"
)

// AssignmentPrinter prints assignments in a concurrency
// safe way.
type AssignmentPrinter struct {
	io.Writer
	*sync.WaitGroup
	Table   *tablewriter.Table
	Now     time.Time
	tableMu sync.Mutex
}

// PrintCourseAssignments prints all the assignments for one course.
func (p *AssignmentPrinter) PrintCourseAssignments(course *canvas.Course, all bool) {
	var dates dueDates
	for as := range course.Assignments(canvas.Opt("order_by", "due_at")) {
		dueAt := as.DueAt.Local()
		if all && dueAt.Before(p.Now) {
			continue
		}
		dates = append(dates, dueDate{
			id:   strconv.Itoa(as.ID),
			name: as.Name,
			date: dueAt,
		})
	}
	sort.Sort(dates)

	// rendering
	p.tableMu.Lock()
	fmt.Fprintln(p, term.Colorf("  %m", course.Name))
	for _, d := range dates {
		p.Table.Append([]string{d.id, d.name, d.date.Format(time.RFC822)})
	}
	if p.Table.NumLines() > 0 {
		p.Table.Render()
	}
	p.Table.ClearRows()
	fmt.Fprintf(p, "\n")

	// clean up
	p.tableMu.Unlock()
	p.Done()
}

// TODO: this may become useful in it's own package
type dueDate struct {
	id, name string
	date     time.Time
}

type dueDates []dueDate

func (dd dueDates) Len() int {
	return len(dd)
}

func (dd dueDates) Swap(i, j int) {
	dd[i], dd[j] = dd[j], dd[i]
}

func (dd dueDates) Less(i, j int) bool {
	return dd[i].date.Before(dd[j].date)
}
