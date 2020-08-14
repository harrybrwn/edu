package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/harrybrwn/config"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/files"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/edu/pkg/term"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
)

// Conf is the global config
var Conf = &Config{}

// Config is the main configuration struct
type Config struct {
	Host         string `yaml:"host" default:"canvas.instructure.com"`
	Editor       string `yaml:"editor" env:"EDITOR"`
	BaseDir      string `yaml:"basedir" default:"$HOME/.edu/files"`
	Token        string `yaml:"token" env:"CANVAS_TOKEN"`
	TwilioNumber string `yaml:"twilio_number"`
	Twilio       struct {
		SID    string `yaml:"sid" env:"TWILIO_SID"`
		Token  string `yaml:"token" env:"TWILIO_TOKEN"`
		Number string `yaml:"number"`
	} `yaml:"twilio"`
	Notifications bool `yaml:"notifications" default:"true"`
	Registration  struct {
		Term string `yaml:"term"`
		Year int    `yaml:"year"`
	} `yaml:"registration"`
	Watch struct {
		Duration string `yaml:"duration" default:"12h"`
		CRNs     []int  `yaml:"crns"`
		Term     string `yaml:"term"`
		Year     int    `yaml:"year"`
		Files    bool   `yaml:"files"`
		Subject  string `yaml:"subject"`
	} `yaml:"watch"`
	Replacements       []files.Replacement            `yaml:"replacements"`
	CourseReplacements map[string][]files.Replacement `yaml:"course-replacements"`
}

// All returns all the commands.
func All(globals *opts.Global) []*cobra.Command {
	all := []*cobra.Command{
		newConfigCmd(),

		newCourseCmd(globals),
		newUserCmd(),

		newDueCmd(globals),
		newFilesCmd(),
		newUploadCmd(),

		newUpdateCmd(),
		newRegistrationCmd(globals),
		newTextCmd(),
	}
	if runtime.GOOS == "linux" {
		all = append(all, genServiceCmd())
	}
	return all
}

func printCourses(out io.Writer, opts *opts.Global, all bool) error {
	var (
		err     error
		courses []*canvas.Course
	)
	courses, err = internal.GetCourses(all)
	if err != nil {
		return internal.HandleAuthErr(err)
	}
	tab := internal.NewTable(out)
	header := []string{"id", "name", "uuid", "code", "ends"}
	internal.SetTableHeader(tab, header, !opts.NoColor)
	for _, c := range courses {
		tab.Append([]string{
			strconv.Itoa(c.ID), c.Name, c.UUID, c.CourseCode, c.EndAt.Format("01/02/06"),
		})
	}
	tab.Render()
	return nil
}

func newCourseCmd(globals *opts.Global) *cobra.Command {
	var all, users, ass bool
	c := &cobra.Command{
		Use:     "course [id|name|uuid|code]",
		Short:   "Get more detailed information for a canvas course or canvas courses.",
		Long:    ``,
		Aliases: []string{"crs"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// if there are not arguments just print out all the courses
			if len(args) < 1 {
				if users || ass {
					return fmt.Errorf("no course\nUsage: %s", cmd.UseLine())
				}
				return printCourses(cmd.OutOrStdout(), globals, all)
			}

			// if the first argument is a number then we assume it
			// is a course ID. Otherwise loop through all the courses
			// to see if any of the courses match.
			var course *canvas.Course
			if id, err := strconv.Atoi(args[0]); err == nil {
				course, err = canvas.GetCourse(id, canvas.IncludeOpt("calendar"))
				if err != nil {
					return err
				}
			} else {
				courses, err := internal.GetCourses(all)
				if err != nil {
					return err
				}
				courseID := args[0] // course identifier
				for _, course = range courses {
					if course.CourseCode == courseID {
						goto FoundCourse
					}
					if course.Name == courseID {
						goto FoundCourse
					}
					if course.UUID == courseID {
						goto FoundCourse
					}
				}
				return errors.New("could not find course")
			FoundCourse:
			}

			cmd.Printf("%d %s\n", course.ID, term.Colorf("%m", course.Name))
			tab := internal.NewTable(cmd.OutOrStdout())

			if users {
				internal.SetTableHeader(tab, []string{"id", "name", "type", "email", "profile"}, !globals.NoColor)
				userlist, err := course.Users(
					canvas.Opt("include_inactive", true),
					canvas.IncludeOpt("enrollments", "email", "observed_users", "custom_links", "avatar_url"),
				)
				if err != nil {
					return err
				}
				var e canvas.Enrollment
				for _, u := range userlist {
					e = u.Enrollments[0]
					tab.Append([]string{
						strconv.Itoa(u.ID), u.Name, strings.Replace(e.Type, "Enrollment", "", 1), u.Email, e.HTMLURL,
					})
				}
				tab.Render()
			} else if ass {
				internal.SetTableHeader(tab, []string{"id", "name", "due date"}, !globals.NoColor)
				var dates dueDates
				for as := range course.Assignments() {
					dueAt := as.DueAt.Local()
					dates = append(dates, dueDate{
						id:   strconv.Itoa(as.ID),
						name: as.Name,
						date: dueAt,
					})
				}
				sort.Sort(dates)
				for _, d := range dates {
					tab.Append([]string{d.id, d.name, d.date.Format(time.RFC822)})
				}
				tab.Render()
			} else {
				cmd.Printf("Download calendar: %s\n", course.Calendar.ICSDownload)
				cmd.Printf("Start: %v\nEnd: %v\n", course.CreatedAt, course.EndAt)
			}
			return nil
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&users, "users", "u", users, "Print out a list of users")
	flags.BoolVarP(&ass, "assignments", "a", ass, "Print out a list of assignments")
	flags.BoolVar(&all, "all", all, "Look at all the courses in your canvas history")
	return c
}

func newUserCmd() *cobra.Command {
	c := &cobra.Command{
		Use:    "user",
		Short:  "",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var user *canvas.User
			opts := []canvas.Option{
				canvas.IncludeOpt("enrollments"),
			}

			if len(args) == 0 {
				user, err = canvas.CurrentUser(opts...)
			} else if id, err := strconv.Atoi(args[0]); err == nil {
				user, err = canvas.GetUser(id, opts...)
			} else {
				user, err = canvas.CurrentUser(opts...)
			}
			if err != nil {
				return err
			}

			fmt.Println(user.Name, user.Enrollments)
			return nil
		},
	}
	return c
}

func newConfigCmd() *cobra.Command {
	var file, edit bool
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration variables.",
		Long: `The config command helps manage program configuration variables.

Configuration for this program has a static component in the for
of a config file. The yaml formatted config file used will be
the file named 'config.yml' that first appears in one of the
following directores in the order that they are listed here by os:

  $EDU_CONFIG
  $XDG_CONFIG_HOME/edu
  $HOME/.edu

The environment variables listed above may changed depending on
the operating system, for example, on windows the config directory
($XDG_CONFIG_HOME on linux) will be found using the environment
variable %AppData%. For more documentation of this see the go docs
https://pkg.go.dev/os?tab=doc#UserConfigDir.
`,
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {
			f := config.FileUsed()
			if file {
				fmt.Println(f)
				return nil
			}
			if edit {
				if f == "" {
					return errs.New("no config file found")
				}
				editor := config.GetString("editor")
				if editor == "" {
					return errs.New("no editor set (use $EDITOR or set it in the config)")
				}
				ex := exec.Command(editor, f)
				ex.Stdout, ex.Stderr, ex.Stdin = os.Stdout, os.Stderr, os.Stdin
				return ex.Run()
			}
			return cmd.Usage()
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use: "get", Short: "Get a config variable",
		Run: func(c *cobra.Command, args []string) {
			for _, arg := range args {
				c.Printf("%+v\n", config.Get(arg))
			}
		}})
	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "edit the config file")
	cmd.Flags().BoolVarP(&file, "file", "f", false, "print the config file path")
	return cmd
}
