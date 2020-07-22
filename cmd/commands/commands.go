package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/files"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config is the main configuration struct
type Config struct {
	Host          string `default:"canvas.instructure.com" yaml:"host"`
	Editor        string `yaml:"editor"`
	BaseDir       string `default:"$HOME/.edu/files" yaml:"basedir"`
	Token         string `yaml:"token"`
	TwilioNumber  string `yaml:"twilio_number"`
	Notifications bool   `default:"true"`
	Registration  struct {
		Term string `yaml:"term"`
		Year int    `yaml:"year"`
	} `yaml:"registration"`
	Watch struct {
		Duration string `default:"12h" yaml:"duration"`
		CRNs     []int  `yaml:"crns"`
		Term     string `yaml:"term"`
		Year     int    `yaml:"year"`
		Files    bool   `yaml:"files"`
	} `yaml:"watch"`
	Replacements       []files.Replacement          `yaml:"replacements"`
	CourseReplacements map[string]files.Replacement `yaml:"course-replacements"`
}

// All returns all the commands.
func All(globals *opts.Global) []*cobra.Command {
	canvasCmd.AddCommand(
		canvasCommands(globals)...,
	)
	all := []*cobra.Command{
		newCoursesCmd(globals),
		newConfigCmd(),
		canvasCmd,
		newUpdateCmd(),
		newRegistrationCmd(globals),
		newTextCmd(),
	}
	if runtime.GOOS == "linux" {
		all = append(all, genServiceCmd())
	}
	return all
}

func newCoursesCmd(opts *opts.Global) *cobra.Command {
	var all bool
	c := &cobra.Command{
		Use:     "courses",
		Short:   "Show info on courses",
		Aliases: []string{"course", "crs"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err     error
				courses []*canvas.Course
			)
			courses, err = internal.GetCourses(all)
			if err != nil {
				return internal.HandleAuthErr(err)
			}
			tab := internal.NewTable(cmd.OutOrStderr())
			header := []string{"id", "name", "uuid", "code", "ends"}
			internal.SetTableHeader(tab, header, !opts.NoColor)
			for _, c := range courses {
				tab.Append([]string{fmt.Sprintf("%d", c.ID), c.Name, c.UUID, c.CourseCode, c.EndAt.Format("01/02/06")})
			}
			tab.Render()
			return nil
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&all, "all", "a", all, "show all courses (defaults to only active courses)")
	return c
}

func newConfigCmd() *cobra.Command {
	var file, edit bool
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage configuration variables.",
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {
			f := viper.ConfigFileUsed()
			if file {
				fmt.Println(f)
				return nil
			}
			if edit {
				if f == "" {
					return errs.New("no config file found")
				}
				editor := viper.GetString("editor")
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
				c.Println(viper.Get(arg))
			}
		}})
	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "edit the config file")
	cmd.Flags().BoolVarP(&file, "file", "f", false, "print the config file path")
	return cmd
}
