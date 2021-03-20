package print

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
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
	tableMu sync.Mutex

	Table    *tablewriter.Table
	Now      time.Time
	Location *time.Location
	All      bool
}

// PrintCourseAssignments prints all the assignments for one course.
func (p *AssignmentPrinter) PrintCourseAssignments(course *canvas.Course) {
	var dates DueDates
	for as := range course.Assignments(canvas.Opt("order_by", "due_at")) {
		dueAt := as.DueAt.In(p.Location)
		if !p.All && dueAt.Before(p.Now) {
			continue
		}
		dates = append(dates, DueDate{
			Id:   strconv.Itoa(as.ID),
			Name: as.Name,
			Date: dueAt,
		})
	}
	sort.Sort(dates)

	// rendering
	p.tableMu.Lock() // might be using share shared table concurrently
	fmt.Fprintln(p, term.Colorf("  %m", course.Name))
	for _, d := range dates {
		p.Table.Append([]string{d.Id, d.Name, d.Date.Format(time.RFC822), humanizeDuration(d.Date.Sub(p.Now))})
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

func humanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"day", days},
		{"hour", hours},
		{"minute", minutes},
		{"second", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("%d %s", chunk.amount, chunk.singularName))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}

type DueDate struct {
	Id, Name string
	Date     time.Time
}

type DueDates []DueDate

func (dd DueDates) Len() int {
	return len(dd)
}

func (dd DueDates) Swap(i, j int) {
	dd[i], dd[j] = dd[j], dd[i]
}

func (dd DueDates) Less(i, j int) bool {
	return dd[i].Date.Before(dd[j].Date)
}
