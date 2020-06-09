package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/school/ucmerced/sched"
	"github.com/harrybrwn/errs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type schedualFlags struct {
	term string
	year int
	open bool
}

func newRegistrationCmd() *cobra.Command {
	var sflags = schedualFlags{
		term: viper.GetString("registration.term"),
		year: viper.GetInt("registration.year"),
	}
	c := &cobra.Command{
		Use:   "registration",
		Short: "Get registration data",
		Long: `Use the 'registration' command to get information on class
registration information.`,
		Aliases: []string{"reg", "register"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if sflags.year == 0 {
				return errs.New("no year given")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var subj, num string
			if len(args) >= 1 {
				subj = args[0]
			}
			if len(args) >= 2 {
				num = args[1]
			}
			schedual, err := sched.BySubject(
				sflags.year,
				sflags.term,
				subj,
				sflags.open,
			) // still works with an empty subj
			if err != nil {
				return err
			}

			tab := internal.NewTable(cmd.OutOrStdout())
			header := []string{"crn", "code", "title", "activity", "time", "seats open"}
			internal.SetTableHeader(tab, header)
			tab.SetAutoWrapText(false)
			for _, c := range schedual.Ordered() {
				if num != "" && c.Number != num {
					continue
				}
				tab.Append(courseRow(c))
			}
			tab.Render()
			return nil
		},
	}
	flags := c.PersistentFlags()
	flags.StringVar(&sflags.term, "term", sflags.term, "specify the term (spring|summer|fall)")
	flags.IntVar(&sflags.year, "year", sflags.year, "specify the year for registration")
	flags.BoolVar(&sflags.open, "open", sflags.open, "only get classes that have seats open")
	c.AddCommand(newCheckCRNCmd(&sflags))
	return c
}

func newCheckCRNCmd(sflags *schedualFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use: "check-crns",
		RunE: func(cmd *cobra.Command, args []string) error {
			schedual, err := sched.Get(sflags.year, sflags.term, true)
			if err != nil {
				return err
			}
			crns := viper.GetIntSlice("crns")
			crnargs, err := stoiArr(args)
			if err != nil {
				return err
			}
			crns = append(crns, crnargs...)

			tab := internal.NewTable(cmd.OutOrStdout())
			header := []string{"crn", "code", "title", "activity", "time", "seats open"}
			internal.SetTableHeader(tab, header)
			tab.SetAutoWrapText(false)
			for _, crn := range crns {
				course, ok := schedual[crn]
				if !ok {
					continue
				}
				tab.Append(courseRow(course))
			}
			if tab.NumLines() == 0 {
				return &internal.Error{Msg: fmt.Sprintf("could not find %v in schedual", crns), Code: 1}
			}
			tab.Render()
			return nil
		},
	}
	return cmd
}

func courseRow(c *sched.Course) []string {
	return []string{
		strconv.FormatInt(int64(c.CRN), 10),
		c.Fullcode,
		cleanTitle(c.Title),
		c.Activity,
		c.Time,
		strconv.FormatInt(int64(c.SeatsAvailible()), 10),
	}
}

// im so sorry for this but the data source is very messy.
func cleanTitle(title string) string {
	title = strings.Replace(title, "Class is fully online", ": Class is fully online", -1)
	ss := []string{
		"Must Also Register for a Corresponding Discussion",
		"Must Also Register for a Corresponding Lab",
		"Must Also Register For a Corresponding Lab",
		"Must Also Register for Corresponding Lab",
		"Must Also Register for a Lab",
		"Taught as a blend of online learning withface-to-face instruction",
		"/Discussion Combination*Lab and Discussion Section Numbers Must Match",
		"Students Who Require a Lab Must Also Register for a Lab",
		"Students Who Require a Lab",
		"If Required, Also Register for a Corresponding Lab",
		"Lab and Discussion Section Numbers Do Not Have to Match",
		"Lab and Discussion Section Numbers Do Not Have To Match",
		"*",
	}
	for _, s := range ss {
		title = strings.Replace(title, s, "", -1)
	}
	if len(title) > 175 {
		title = title[:175]
	}
	return title
}

func stoiArr(arr []string) ([]int, error) {
	ints := make([]int, len(arr))
	for i, n := range arr {
		num, err := strconv.ParseInt(n, 10, 32)
		if err != nil {
			return nil, err
		}
		ints[i] = int(num)
	}
	return ints, nil
}
