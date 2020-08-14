package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/harrybrwn/config"
	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/files"
	"github.com/spf13/cobra"
)

type updateCmd struct {
	all, verbose bool
	basedir      string
	testPatters  bool
	sortBy       []string
}

func newUpdateCmd() *cobra.Command {
	uc := &updateCmd{
		all:     false,
		verbose: false,
		sortBy:  []string{"created_at"},
		basedir: os.ExpandEnv(config.GetString("basedir")),
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Download all your files from canvas",
		RunE:  uc.run,
	}
	flags := cmd.Flags()
	flags.BoolVarP(&uc.all, "all", "a", uc.all, "download files from all courses, defaults to only active courses")
	flags.BoolVarP(&uc.verbose, "verbose", "v", uc.verbose, "run update in verbose mode (prints out files)")
	flags.BoolVar(&uc.testPatters, "test-patterns", uc.testPatters, "test the replacement patterns from the config file")
	flags.StringVar(&uc.basedir, "base-dir", uc.basedir, "base directory for file downloads")
	flags.StringArrayVarP(&uc.sortBy, "sort-by", "s", uc.sortBy, "select the file sorting methods")
	return cmd
}

func (uc *updateCmd) run(cmd *cobra.Command, args []string) (err error) {
	courses, err := internal.GetCourses(uc.all)
	if err != nil {
		return internal.HandleAuthErr(err)
	}
	dl := files.NewDownloader(uc.basedir)
	if uc.verbose {
		dl.Stdout = os.Stdout
	}

	var fn = dl.Download
	if uc.testPatters {
		fn = dl.CheckReplacements
	}
	courseReps := upperMapKeys(Conf.CourseReplacements)

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
		fn(course, reps)
	}
	dl.Wait()
	fmt.Println("done.")
	return nil
}

func upperMapKeys(m map[string][]files.Replacement) map[string][]files.Replacement {
	cp := make(map[string][]files.Replacement)
	for key, val := range m {
		cp[strings.ToUpper(key)] = val
	}
	return cp
}
