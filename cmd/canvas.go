package cmd

import (
	"fmt"

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

func newFilesCmd() *cobra.Command {
	var (
		contentType string
		sortby      = []string{"created_at"}
		all         bool
	)
	c := &cobra.Command{
		Use:   "files",
		Short: "List course files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			courses, err := canvas.ActiveCourses()
			if err != nil {
				return err
			}
			if all {
				completed, err := canvas.CompletedCourses()
				if err != nil {
					return err
				}
				courses = append(courses, completed...)
			}

			opts := []canvas.Option{canvas.SortOpt(sortby...)}
			if contentType != "" {
				opts = append(opts, canvas.ContentType(contentType))
			}
			count := 0
			for _, course := range courses {
				course.SetErrorHandler(errorHandler)
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
			courses, err := getCourses(false)
			if err != nil {
				return err
			}
			for _, course := range courses {
				for as := range course.Assignments() {
					fmt.Println(as.Name, as.DueAt)
				}
			}
			return nil
		},
	}
)

func getCourses(all bool) ([]*canvas.Course, error) {
	courses, err := canvas.ActiveCourses()
	if err != nil {
		return nil, err
	}
	if all {
		completed, err := canvas.CompletedCourses()
		if err != nil {
			return courses, err
		}
		courses = append(courses, completed...)
	}
	return courses, nil
}
