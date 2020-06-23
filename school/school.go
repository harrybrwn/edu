package school

import "strings"

//go:generate stringer -type School -linecomment school.go

// School is a type that identifies a school
type School int

const (
	// UCBerkeley is the school type for UC Berkeley.
	UCBerkeley School = iota
	// UCMerced is the school type of UC Merced
	UCMerced
)

// Course is an interface that defines a generic
// course object for a school schedule.
type Course interface {
	SeatsOpen() int
	Name() string
	ID() int
}

// Schedule is an interface that represents a
// school schedule.
type Schedule interface {
	Courses() []Course
	Get(id int) Course
}

// FromName returns a school code based on the name of
// the school. Returns -1 if the name is unknown.
func FromName(schoolname string) School {
	switch strings.ToLower(schoolname) {
	case "ucmerced", "merced":
		return UCMerced
	case "ucberkeley", "berkeley":
		return UCBerkeley
	default:
		return -1
	}
}
