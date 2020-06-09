package commands

import (
	"fmt"
	"os"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type fileFinder struct {
	contentType string
	search      string
	all         bool
}

func (ff *fileFinder) flagset() *pflag.FlagSet {
	fset := pflag.NewFlagSet("", pflag.ExitOnError)
	ff.addToFlagSet(fset)
	return fset
}

func (ff *fileFinder) getCourses() ([]*canvas.Course, error) {
	courses, err := internal.GetCourses(ff.all, canvas.OptStudent)
	if err != nil {
		return courses, err
	}
	return courses, nil
}

func (ff *fileFinder) options() (opts []canvas.Option) {
	if ff.contentType != "" {
		opts = append(opts, canvas.ContentTypes(ff.contentType))
	}
	if ff.search != "" {
		opts = append(opts, canvas.Opt("search_term", ff.search))
	}
	return opts
}

func (ff *fileFinder) addToFlagSet(fset *pflag.FlagSet) {
	fset.BoolVarP(&ff.all, "all", "a", ff.all, "query files from all courses")
	fset.StringVarP(&ff.contentType, "content-type", "c", "", "filter out files by content type (ex. application/pdf)")
	fset.StringVar(&ff.search, "search", "", "search for files by name")
}

func init() {
	canvasCmd.AddCommand(
		newFilesCmd(),
		dueCmd,
	)
}

var (
	canvasCmd = &cobra.Command{
		Use:     "canvas",
		Aliases: []string{"canv", "ca"},
		Short:   "A small collection of helper commands for canvas",
	}
	dueCmd = &cobra.Command{
		Use:   "due",
		Short: "List all the due date on canvas.",
		RunE: func(cmd *cobra.Command, args []string) error {
			courses, err := internal.GetCourses(false)
			if err != nil {
				return err
			}
			tab := internal.NewTable(cmd.OutOrStdout())
			internal.SetTableHeader(tab, []string{"name", "due"})
			for _, course := range courses {
				for as := range course.Assignments() {
					tab.Append([]string{as.Name, as.DueAt.String()})
				}
			}
			tab.Render()
			return nil
		},
	}

	testCmd = &cobra.Command{
		Use:    "test",
		Hidden: false,
		RunE:   func(cmd *cobra.Command, args []string) error { return nil },
	}
)

func newFilesCmd() *cobra.Command {
	var (
		sortby = []string{"created_at"}
		ff     fileFinder
	)
	c := &cobra.Command{
		Use:   "files",
		Short: "List course files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			courses, err := ff.getCourses()
			if err != nil {
				return err
			}
			opts := []canvas.Option{canvas.SortOpt(sortby...)}
			opts = append(opts, ff.options()...)
			count := 0
			for _, course := range courses {
				if course.AccessRestrictedByDate {
					fmt.Fprintf(
						os.Stderr, "Access to %d %s has been restricted to a certain date\n",
						course.ID, course.Name)
					continue
				}
				files := course.Files(opts...)
				for f := range files {
					cmd.Println(f.CreatedAt, f.Size, f.Filename)
					count++
				}
			}
			cmd.Println(count, "files total.")
			return nil
		},
	}
	flags := c.Flags()
	flags.StringArrayVarP(&sortby, "sortyby", "s", sortby, "how the files should be sorted")
	ff.addToFlagSet(flags)
	return c
}
