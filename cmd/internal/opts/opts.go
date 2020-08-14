package opts

import "github.com/spf13/pflag"

// Global is a set of global flags
type Global struct {
	NoColor bool
}

// AddToFlagSet will add the flags to a flag set
func (g *Global) AddToFlagSet(set *pflag.FlagSet) {
	set.BoolVar(&g.NoColor, "nocolor", false, "turn off colors")
}

// ScheduleFlags is a collection of flag variables
type ScheduleFlags struct {
	*Global
	Term string
	Year int
	Open bool

	// columns []string // TODO: figure this out
}

// Install will install the flags to a flag set
func (sf *ScheduleFlags) Install(fset *pflag.FlagSet) {
	fset.StringVar(&sf.Term, "term", sf.Term, "specify the term (spring|summer|fall)")
	fset.IntVar(&sf.Year, "year", sf.Year, "specify the year for registration")
	fset.BoolVar(&sf.Open, "open", sf.Open, "only get classes that have seats open")
}
