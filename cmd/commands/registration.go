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

func newRegistrationCmd() *cobra.Command {
	var (
		term string = viper.GetString("registration.term")
		year int    = viper.GetInt("registration.year")
	)
	c := &cobra.Command{
		Use:   "registration",
		Short: "Get registration data",
		Long: `Use the 'registration' command to get information on class
registration information.`,
		Aliases: []string{"reg", "register"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var subj, num string
			if len(args) >= 1 {
				subj = args[0]
			}
			if len(args) >= 2 {
				num = args[1]
			}
			if year == 0 {
				return errs.New("no year given")
			}

			schedual, err := sched.BySubject(year, term, subj) // still works with an empty subj
			if err != nil {
				return err
			}

			tab := internal.NewTable(cmd.OutOrStdout())
			header := []string{"crn", "code", "title", "activity", "time", "seats open"}
			internal.SetTableHeader(tab, header)
			tab.SetAutoWrapText(false)
			for _, c := range schedual {
				if num != "" && c.Number != num {
					continue
				}
				tab.Append(courseRow(c))
			}
			tab.Render()
			return nil
		},
	}
	flags := c.Flags()
	flags.StringVar(&term, "term", term, "specify the term (spring|summer|fall)")
	flags.IntVar(&year, "year", year, "specify the year for registration")
	c.AddCommand(newCheckCRNCmd())
	return c
}

func newCheckCRNCmd() *cobra.Command {
	var (
		term string = viper.GetString("registration.term")
		year int    = viper.GetInt("registration.year")
	)
	cmd := &cobra.Command{
		Use: "check-crns",
		RunE: func(cmd *cobra.Command, args []string) error {
			schedual, err := sched.Get(year, term) // still works with an empty subj
			if err != nil {
				return err
			}
			tab := internal.NewTable(cmd.OutOrStdout())
			header := []string{"crn", "code", "title", "activity", "time", "seats open"}
			internal.SetTableHeader(tab, header)
			tab.SetAutoWrapText(false)
			for _, crn := range args {
				crn, err := strconv.ParseInt(crn, 10, 32)
				if err != nil {
					return err
				}
				course, ok := schedual[int(crn)]
				if !ok {
					continue
				}
				tab.Append(courseRow(course))
			}
			if tab.NumLines() == 0 {
				return &internal.Error{Msg: fmt.Sprintf("could not find %s in schedual", strings.Join(args, ", ")), Code: 1}
			}
			tab.Render()
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&term, "term", term, "specify the term (spring|summer|fall)")
	flags.IntVar(&year, "year", year, "specify the year for registration")
	return cmd
}

func courseRow(c *sched.Course) []string {
	return []string{
		strconv.FormatInt(int64(c.CRN), 10),
		c.Fullcode,
		strings.Replace(c.Title, "Must Also Register for a Corresponding Lab", "", -1),
		c.Activity,
		c.Time,
		strconv.FormatInt(int64(c.SeatsAvailible()), 10),
	}
}
