package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	table "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

// Execute will execute the root comand on the cli
func Execute() (err error) {
	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && ok {
		path := os.ExpandEnv("$HOME/.config/edu")
		if err = mkdir(path); err != nil {
			return fmt.Errorf("couldn't create config dir: %w", err)
		}
		viper.SetConfigFile(filepath.Join(path, "config.yml"))
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	root.AddCommand(
		canvasCmd,
		newUpdateCmd(),
		newConfigCmd(),
		newCoursesCmd(),
		completionCmd,
		newRegistrationCmd(),
	)

	err = root.Execute()
	if err == nil {
		return nil
	}
	errorMessage(err)
	os.Exit(1)
	return nil
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	viper.AddConfigPath("$XDG_CONFIG_HOME/edu")
	viper.AddConfigPath("$HOME/.config/edu")
	viper.AddConfigPath("$HOME/.edu")

	viper.AddConfigPath("$XDG_CONFIG_HOME/canvas")
	viper.AddConfigPath("$HOME/.config/canvas")
	viper.AddConfigPath("$HOME/.canvas")

	viper.SetEnvPrefix("edu")
	viper.BindEnv("host")
	viper.BindEnv("canvas_token", "CANVAS_TOKEN")
	viper.SetDefault("editor", os.Getenv("EDITOR"))
}

var (
	root = &cobra.Command{
		Use:           "edu",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			host := viper.GetString("host")
			if host != "" {
				canvas.DefaultHost = host
			}
			token := viper.GetString("token")
			if token != "" {
				canvas.SetToken(os.ExpandEnv(token))
			} else {
				viper.Set("token", viper.GetString("canvas_token"))
				canvas.SetToken(os.ExpandEnv(viper.GetString("token")))
			}
			canvas.ConcurrentErrorHandler = errorHandler
		},
	}

	canvasCmd = &cobra.Command{
		Use:     "canvas",
		Aliases: []string{"canv", "ca"},
		Short:   "A small collection of helper commands for canvas",
	}

	completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Print a completion script to stdout.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			root := cmd.Root()
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				return errors.New("no shell type given")
			}
			switch args[0] {
			case "zsh":
				return root.GenZshCompletion(out)
			case "ps", "powershell":
				return root.GenPowerShellCompletion(out)
			case "bash":
				return root.GenBashCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, false)
			}
			return errs.New("unknown shell type")
		},
		ValidArgs: []string{"zsh", "bash", "ps", "powershell", "fish"},
		Aliases:   []string{"comp"},
	}
)

func newCoursesCmd() *cobra.Command {
	var (
		all     bool
		pending bool
	)
	c := &cobra.Command{
		Use:     "courses",
		Short:   "Show info on courses",
		Aliases: []string{"course", "crs"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err     error
				courses []*canvas.Course
			)
			if pending {
				courses, err = canvas.Courses(canvas.Opt("enrollment_state", "invited_or_pending"))
			} else {
				courses, err = getCourses(all)
			}
			if err != nil {
				return err
			}
			var namelen = 1
			for _, course := range courses {
				if len(course.Name) > namelen {
					namelen = len(course.Name)
				}
			}
			tab := newTable(cmd.OutOrStderr())
			header := []string{"id", "name", "uuid", "code", "ends"}
			setTableHeader(tab, header)
			for _, c := range courses {
				tab.Append([]string{fmt.Sprintf("%d", c.ID), c.Name, c.UUID, c.CourseCode, c.EndAt.Format("01/02/06")})
			}
			tab.Render()
			return nil
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&all, "all", "a", all, "show all courses (defaults to only active courses)")
	flags.BoolVar(&pending, "pending", pending, "show all invited or pending courses")
	return c
}

func newConfigCmd() *cobra.Command {
	var file, edit bool
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage configuration",
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

func newTable(r io.Writer) *table.Table {
	t := table.NewWriter(r)
	t.SetBorder(false)
	t.SetColumnSeparator("")
	t.SetAlignment(table.ALIGN_LEFT)
	t.SetAutoFormatHeaders(false)
	t.SetHeaderLine(false)
	t.SetHeaderAlignment(table.ALIGN_LEFT)
	return t
}

func mkdir(d string) error {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err = os.MkdirAll(d, 0775); err != nil {
			return err
		}
	}
	return nil
}

func errorMessage(err error) {
	switch err.(type) {
	case *canvas.AuthError:
		fmt.Fprintf(os.Stderr, "Authentication Error: %v\n", err)
	case *canvas.Error:
		fmt.Fprintf(os.Stderr, "Canvas Error: %v\n", err)
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func errorHandler(e error) error {
	if e != nil {
		fmt.Println("Error: " + e.Error())
		os.Exit(1)
	}
	return nil
}

func stop(msg string) {
	errmsg(msg)
	os.Exit(1)
}

func errmsg(msg interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", msg)
}
