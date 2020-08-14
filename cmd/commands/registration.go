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
	"github.com/harrybrwn/edu/cmd/internal/config"
	"github.com/harrybrwn/edu/cmd/internal/files"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/edu/cmd/internal/watch"
	"github.com/harrybrwn/edu/pkg/term"
	"github.com/harrybrwn/edu/school/ucmerced/ucm"
	"github.com/harrybrwn/errs"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		term:   config.GetString("registration.term"),
		year:   config.GetInt("registration.year"),
		Global: globals,
	}

	c := &cobra.Command{
		Use:   "registration",
		Short: "Get registration data",
		Long: `Use the 'registration' command to get information on class
registration information.`,
		Aliases: []string{"reg", "register"},
		Example: "" +
			"$ edu registration cse 100 --term=fall\n" +
			"\t$ edu reg --open --year=2021 --term=summer WRI 10",
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
			schedule, err := ucm.BySubject(
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
				tab.Append(courseRow(c, true, sflags))
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
			schedule, err := ucm.BySubject(sflags.year, sflags.term, subject, true)
			if err != nil {
				return err
			}
			crns := config.GetIntSlice("crns")
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
				tab.Append(courseRow(course, false, *sflags))
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

type crnWatcher struct {
	crns    []int
	names   []string
	subject string
	flags   scheduleFlags
	verbose bool
}

func (cw *crnWatcher) Watch() error {
	var (
		subject = cw.subject
		crns    = cw.crns
	)
	if config.GetInt("watch.year") != 0 {
		cw.flags.year = config.GetInt("watch.year")
	}
	if config.GetString("watch.term") != "" {
		cw.flags.term = config.GetString("watch.term")
	}
	if config.GetString("watch.subject") != "" {
		subject = config.GetString("watch.subject")
	}
	configCrns := config.GetIntSlice("watch.crns")
	if len(configCrns) > 0 {
		crns = append(crns, configCrns...)
	}
	if len(crns) < 1 {
		return errors.New("no crns to check (see 'edu config' watch settings)")
	}
	err := checkCRNList(crns, subject, &cw.flags)
	if err != nil {
		if cw.verbose {
			fmt.Println(err)
		}
		return err
	}
	return nil
}

func watchFiles() error {
	basedir := config.GetString("basedir")
	if basedir == "" {
		return errors.New("cannot download files to an empty base directory")
	}
	courses, err := internal.GetCourses(false)
	if err != nil {
		return internal.HandleAuthErr(err)
	}
	courseReps := upperMapKeys(Conf.CourseReplacements)
	dl := files.NewDownloader(basedir)
	for _, course := range courses {
		if course.AccessRestrictedByDate {
			continue
		}
		reps, ok := courseReps[course.CourseCode]
		if !ok {
			reps = Conf.Replacements
		} else {
			reps = append(Conf.Replacements, reps...)
		}
		dl.Download(course, reps)
	}
	dl.Wait()
	return nil
}

func newWatchCmd(sflags *scheduleFlags) *cobra.Command {
	// TODO: add a setting in the config file for text msg updates with watch
	/*
		watch:
			notify: true # use twilio to notify
		twilio:
			token: <api token>
			sid: <api SID>
			number: <twilio number>
	*/
	var (
		subject string
		verbose bool
		term    = config.GetString("watch.term")
		year    = config.GetInt("watch.year")
	)
	if term != "" {
		sflags.term = term
	}
	if year != 0 {
		sflags.year = year
	}

	c := &cobra.Command{
		Use:   "watch",
		Short: "Watch for availability changes in a list of CRNs",
		Long: "Watch for availability changes in a list of CRNs." +
			"",
		RunE: func(cmd *cobra.Command, args []string) error {
			basecrns, err := stroiArr(args)
			if err != nil {
				return err
			}
			var duration time.Duration
			duration, err = time.ParseDuration(config.GetString("watch.duration"))
			if err != nil {
				return err
			}

			crnWatch := &crnWatcher{
				crns:    basecrns,
				subject: subject,
				flags:   *sflags,
				verbose: verbose,
			}

			var watches = []watch.Watcher{crnWatch}
			if config.GetBool("watch.files") {
				watches = append(watches, watch.WatcherFunc(watchFiles))
			}
			for {
				for _, wt := range watches {
					go func(wt watch.Watcher) {
						if err := wt.Watch(); err != nil {
							log.Printf("Watch Error: %s\n", err.Error())
						}
					}(wt)
				}
				time.Sleep(duration)

				// refresh variables from config
				if err = config.ReadConfigFile(); err != nil {
					log.Printf("could not refresh config during 'watch': %v", err)
				}
				if config.GetString("watch.duration") != "" {
					newdur, err := time.ParseDuration(config.GetString("watch.duration"))
					if err != nil {
						log.Printf("could not refresh duration: %v", err)
					} else if newdur != 0 {
						duration = newdur
					}
				}
			}
		},
	}
	c.Flags().BoolVarP(&verbose, "verbose", "v", verbose, "print out any errors")
	c.Flags().StringVar(&subject, "subject", "", "check the CRNs for a specific subject")
	return c
}

func checkCRNList(crns []int, subject string, sflags *scheduleFlags) error {
	schedule, err := ucm.BySubject(sflags.year, sflags.term, subject, true)
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
	if config.GetBool("notifications") {
		return beeep.Notify("Found Open Courses", msg, "")
	}
	return nil
}

func courseRow(c *ucm.Course, title bool, flags scheduleFlags) []string {
	var timeStr = "TBD"
	if c.Time.Start.Hour() != 0 && c.Time.End.Hour() != 0 {
		timeStr = fmt.Sprintf("%s-%s",
			c.Time.Start.Format("3:04pm"),
			c.Time.End.Format("3:04pm"))
	}
	days := strjoin(c.Days, ",")
	seats := c.SeatsOpen()
	var open = strconv.Itoa(seats)
	if !flags.NoColor {
		if seats <= 0 {
			open = term.Red(open)
		} else {
			open = term.Green(open)
		}
	}

	if title {
		return []string{
			strconv.Itoa(c.CRN),
			c.Fullcode,
			open,
			c.Activity,
			cleanTitle(c.Title),
			timeStr,
			days,
		}
	}
	return []string{
		strconv.Itoa(c.CRN),
		c.Fullcode,
		open,
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

func courseAsDict(c *ucm.Course) map[string]interface{} {
	m := make(map[string]interface{})
	mapstructure.Decode(c, &m)
	return m
}
