package commands

import (
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
)

func init() {
	canvasCmd.AddCommand(
		newFilesCmd(),
		// modulesCmd,
		dueCmd,
	)
}

var canvasCmd = &cobra.Command{
	Use:     "canvas",
	Aliases: []string{"canv", "ca"},
	Short:   "A small collection of helper commands for canvas",
}

func newFilesCmd() *cobra.Command {
	var (
		contentType string
		// search      string
		sortby = []string{"created_at"}
		all    bool
	)
	c := &cobra.Command{
		Use:   "files",
		Short: "List course files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			courses, err := internal.GetCourses(all)
			if err != nil {
				return err
			}

			opts := []canvas.Option{canvas.SortOpt(sortby...)}
			if contentType != "" {
				opts = append(opts, canvas.ContentTypes(contentType))
			}
			count := 0
			for _, course := range courses {
				files := course.Files(opts...)
				for f := range files {
					cmd.Println(f.CreatedAt, f.Size, f.Filename)
					count++
				}
			}
			cmd.Println(count, "files total")
			return nil
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&all, "all", "a", all, "get files from all courses")
	flags.StringVarP(&contentType, "content-type", "c", "", "filter out files by content type (ex. application/pdf)")
	// flags.StringVar(&search, "search", "", "search for files by name")
	flags.StringArrayVarP(&sortby, "sortyby", "s", sortby, "how the files should be sorted")
	return c
}

var (
	modulesCmd = &cobra.Command{
		Use:     "modules",
		Short:   "List canvas modules.",
		Aliases: []string{"mod", "mods"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
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
