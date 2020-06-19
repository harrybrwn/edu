package commands

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/edu/pkg/info"
	"github.com/harrybrwn/edu/school/ucmerced/sched"
	"github.com/harrybrwn/errs"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type scheduleFlags struct {
	*opts.Global
	term    string
	year    int
	open    bool
	columns []string
}

func (sf *scheduleFlags) install(fset *pflag.FlagSet) {
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
	var sflags = scheduleFlags{
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
			schedule, err := sched.BySubject(
				sflags.year, sflags.term,
				subj, sflags.open,
			) // still works with an empty subj
			if err != nil {
				return err
			}

			tab := internal.NewTable(cmd.OutOrStdout())
			internal.SetTableHeader(tab, regHeader, !sflags.NoColor)
			tab.SetAutoWrapText(false)
			for _, c := range schedule.Ordered() {
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
	c.AddCommand(newCheckCRNCmd(&sflags), newWatchCmd(&sflags))
	return c
}

func newCheckCRNCmd(sflags *scheduleFlags) *cobra.Command {
	var subject string
	cmd := &cobra.Command{
		Use: "check-crns",
		RunE: func(cmd *cobra.Command, args []string) error {
			schedule, err := sched.BySubject(sflags.year, sflags.term, subject, true)
			if err != nil {
				return err
			}
			crns := viper.GetIntSlice("crns")
			crnargs, err := stroiArr(args)
			if err != nil {
				return err
			}
			crns = append(crns, crnargs...)

			tab := internal.NewTable(cmd.OutOrStdout())
			header := []string{"crn", "code", "open", "type", "time", "days"}
			internal.SetTableHeader(tab, header, !sflags.NoColor)
			tab.SetAutoWrapText(false)
			for _, crn := range crns {
				course, ok := schedule[crn]
				if !ok {
					continue
				}
				tab.Append(courseRow(course, false))
			}
			if tab.NumLines() == 0 {
				return &internal.Error{Msg: fmt.Sprintf("could not find %v in schedule", crns), Code: 1}
			}
			tab.Render()
			return nil
		},
	}
	cmd.Flags().StringVar(&subject, "subject", "", "check the CRNs for a specific subject")
	return cmd
}

type watcher interface {
	Watch() error
}

type watcherFunc func() error

func (wf watcherFunc) Watch() error {
	return wf()
}

type crnWatcher struct {
	crns    []int
	subject string
	flags   *scheduleFlags
	verbose bool
}

func (cw *crnWatcher) Watch() error {
	err := checkCRNList(cw.crns, cw.subject, cw.flags)
	if err != nil {
		if cw.verbose {
			fmt.Println(err)
		}
		return err
	}
	return nil
}

func newWatchCmd(sflags *scheduleFlags) *cobra.Command {
	var (
		subject string
		verbose bool
	)
	c := &cobra.Command{
		Use:   "watch",
		Short: "Watch for availability changes in a list of CRNs",
		Long: "Watch for availability changes in a list of CRNs." +
			"",
		RunE: func(cmd *cobra.Command, args []string) error {
			crns, err := stroiArr(args)
			if err != nil {
				return err
			}
			crns = append(crns, viper.GetIntSlice("watch.crns")...)
			if len(crns) < 1 {
				return errors.New("no crns to check")
			}

			var duration time.Duration
			duration, err = time.ParseDuration(viper.GetString("watch.duration"))
			if err != nil {
				return err
			}

			var watches = []watcher{
				&crnWatcher{
					crns:    crns,
					subject: subject,
					flags:   sflags,
					verbose: verbose,
				},
			}
			if viper.GetBool("watch.files") {
				watches = append(watches, watcherFunc(func() error {
					return nil
				}))
			}
			if !viper.GetBool("no_runtime_info") {
				go info.Intrp()
			}
			for {
				for _, wt := range watches {
					go runwatch(wt)
				}
				time.Sleep(duration)
			}
		},
	}
	c.Flags().BoolVarP(&verbose, "verbose", "v", verbose, "print out any errors")
	c.Flags().StringVar(&subject, "subject", "", "check the CRNs for a specific subject")
	return c
}

func runwatch(wt watcher) {
	if err := wt.Watch(); err != nil {
		log.Printf("Watch Error: %s", err.Error())
	}
}

func checkCRNList(crns []int, subject string, sflags *scheduleFlags) error {
	schedule, err := sched.BySubject(sflags.year, sflags.term, subject, true)
	if err != nil {
		return err
	}
	openCrns := make([]int, 0)
	for _, crn := range crns {
		_, ok := schedule[crn]
		if !ok {
			continue
		}
		openCrns = append(openCrns, crn)
	}
	if len(openCrns) == 0 {
		return &internal.Error{Msg: fmt.Sprintf("could not find %v in schedule", crns), Code: 1}
	}
	msg := "Open crns:\n"
	for _, crn := range openCrns {
		msg += fmt.Sprintf("%d\n", crn)
	}

	if viper.GetBool("notifications") {
		return beeep.Notify("Found Open Courses", msg, "")
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
			strconv.Itoa(c.SeatsOpen()),
			c.Activity,
			cleanTitle(c.Title),
			timeStr,
			days,
		}
	}
	return []string{
		strconv.Itoa(c.CRN),
		c.Fullcode,
		strconv.Itoa(c.SeatsOpen()),
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

func stroiArr(arr []string) (ints []int, err error) {
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
