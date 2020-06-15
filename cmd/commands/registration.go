package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/edu/school/ucmerced/sched"
	"github.com/harrybrwn/errs"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type schedualFlags struct {
	*opts.Global
	term   string
	year   int
	open   bool
	colums []string
}

func (sf *schedualFlags) install(fset *pflag.FlagSet) {
	fset.StringVar(&sf.term, "term", sf.term, "specify the term (spring|summer|fall)")
	fset.IntVar(&sf.year, "year", sf.year, "specify the year for registration")
	fset.BoolVar(&sf.open, "open", sf.open, "only get classes that have seats open")
}

var regHeader = []string{
	"crn",
	"code",
	"seats open",
	"activity",
	"title",
	"time",
	"days",
}

func newRegistrationCmd(globals *opts.Global) *cobra.Command {
	var sflags = schedualFlags{
		term:   viper.GetString("registration.term"),
		year:   viper.GetInt("registration.year"),
		Global: globals,
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				subj string
				num  int
			)
			if len(args) >= 1 {
				subj = args[0]
			}
			if len(args) >= 2 {
				num, err = strconv.Atoi(args[1])
				if err != nil {
					return err
				}
			}
			schedual, err := sched.BySubject(
				sflags.year, sflags.term,
				subj, sflags.open,
			) // still works with an empty subj
			if err != nil {
				return err
			}

			tab := internal.NewTable(cmd.OutOrStdout())
			internal.SetTableHeader(tab, regHeader, sflags.NoColor)
			tab.SetAutoWrapText(false)
			for _, c := range schedual.Ordered() {
				if num != 0 && c.Number != num {
					continue
				}
				tab.Append(courseRow(c, true))
			}
			if tab.NumLines() == 0 {
				return &internal.Error{Msg: "no matches", Code: 1}
			}
			tab.Render()
			return nil
		},
	}
	sflags.install(c.PersistentFlags())
	c.AddCommand(newCheckCRNCmd(&sflags))
	return c
}

func newCheckCRNCmd(sflags *schedualFlags) *cobra.Command {
	var subject string
	cmd := &cobra.Command{
		Use: "check-crns",
		RunE: func(cmd *cobra.Command, args []string) error {
			schedual, err := sched.BySubject(sflags.year, sflags.term, subject, true)
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
			header := []string{"crn", "code", "open", "type", "time", "days"}
			internal.SetTableHeader(tab, header, sflags.NoColor)
			tab.SetAutoWrapText(false)
			for _, crn := range crns {
				course, ok := schedual[crn]
				if !ok {
					continue
				}
				tab.Append(courseRow(course, false))
			}
			if tab.NumLines() == 0 {
				return &internal.Error{Msg: fmt.Sprintf("could not find %v in schedual", crns), Code: 1}
			}
			tab.Render()
			return nil
		},
	}
	cmd.Flags().StringVar(&subject, "subject", "", "check the CRNs for a specific subject")
	return cmd
}

func checkCRNList(crns []int, subject string, sflags *schedualFlags) error {
	schedual, err := sched.BySubject(sflags.year, sflags.term, subject, true)
	if err != nil {
		return err
	}
	openCrns := make([]int, 0)
	for _, crn := range crns {
		_, ok := schedual[crn]
		if !ok {
			continue
		}
		openCrns = append(openCrns, crn)
	}
	if len(openCrns) == 0 {
		return &internal.Error{Msg: fmt.Sprintf("could not find %v in schedual", crns), Code: 1}
	}
	if viper.GetBool("notifications") {
		return beeep.Notify("Found Open Courses", fmt.Sprintf("Open crns: %v", openCrns), "")
	}
	return nil
}

func courseRow(c *sched.Course, title bool) []string {
	var timeStr = "TBD"
	if c.Time.Start.Hour() != 0 && c.Time.End.Hour() != 0 {
		timeStr = fmt.Sprintf("%s-%s",
			c.Time.Start.Format("3:04pm"),
			c.Time.End.Format("3:04pm"))
	}
	days := strjoin(c.Days, ",")
	if title {
		return []string{
			strconv.Itoa(c.CRN),
			c.Fullcode,
			strconv.Itoa(c.SeatsAvailible()),
			c.Activity,
			cleanTitle(c.Title),
			timeStr,
			days,
		}
	}
	return []string{
		strconv.Itoa(c.CRN),
		c.Fullcode,
		strconv.Itoa(c.SeatsAvailible()),
		c.Activity,
		timeStr,
		days,
	}
}

var mustAlsoRegex = regexp.MustCompile(`Must Also.*$`)

func cleanTitle(title string) string {
	title = mustAlsoRegex.ReplaceAllString(title, "")
	title = strings.Replace(title, "Class is fully online", ": Class is fully online", -1)
	if len(title) > 175 {
		title = title[:175]
	}
	return title
}

func stoiArr(arr []string) (ints []int, err error) {
	ints = make([]int, len(arr))
	for i, n := range arr {
		ints[i], err = strconv.Atoi(n)
		if err != nil {
			return
		}
	}
	return
}

func strjoin(list []time.Weekday, sep string) string {
	strs := make([]string, len(list))
	for i, s := range list {
		strs[i] = s.String()[:3]
	}
	return strings.Join(strs, sep)
}

func courseAsDict(c *sched.Course) map[string]interface{} {
	m := make(map[string]interface{})
	mapstructure.Decode(c, &m)
	return m
}
