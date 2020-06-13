package opts

import "github.com/spf13/pflag"

type Options interface {
	AddToFlagSet(*pflag.FlagSet)
}

type Global struct {
	NoColor bool
}

func (g *Global) AddToFlagSet(set *pflag.FlagSet) {
	set.BoolVar(&g.NoColor, "nocolor", false, "turn of colors")
}
