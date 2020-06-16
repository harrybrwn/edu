package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// All returns all the commands.
func All(globals *opts.Global) []*cobra.Command {
	return []*cobra.Command{
		newCoursesCmd(globals),
		newConfigCmd(),
		canvasCmd,
		newUpdateCmd(),
		newRegistrationCmd(globals),
		newTextCmd(),
		genServiceCmd(),
	}
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
				return err
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
