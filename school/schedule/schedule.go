package schedule

import (
	"errors"

	"github.com/harrybrwn/edu/school"
	"github.com/harrybrwn/edu/school/ucberkeley/btime"
	"github.com/harrybrwn/edu/school/ucmerced/ucm"
)

// Config is a set of config variables for
// finding a school schedule.
type Config struct {
	Term string
	Year int
	// FilterClosed, if true, will filter out any courses
	// that do not have seats open
	FilterClosed bool
	CourseName   string
}

// New will get a schedule based on the school type given.
func New(sc school.School, config *Config) (school.Schedule, error) {
	switch sc {
	case school.UCBerkeley:
		catalog, err := btime.New()
		if err != nil {
			return nil, err
		}
		return catalog, nil
	case school.UCMerced:
		sched, err := ucm.BySubject(
			config.Year,
			config.Term,
			config.CourseName,
			config.FilterClosed,
		)
		return &sched, err
	default:
		return nil, errors.New("unknown school")
	}
}
