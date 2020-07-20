package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/files"
	"github.com/harrybrwn/errs"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		basedir: os.ExpandEnv(viper.GetString("basedir")),
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
	coursereps, replacements, err := getReplacements()
	if err != nil {
		return err
	}

	var fn = dl.Download
	if uc.testPatters {
		fn = dl.CheckReplacements
	}

	for _, course := range courses {
		if course.AccessRestrictedByDate {
			continue
		}
		reps, ok := coursereps[course.CourseCode]
		if !ok {
			reps = replacements
		} else {
			reps = append(replacements, reps...)
		}
		fn(course, reps)
	}
	dl.Wait()
	fmt.Println("done.")
	return nil
}

func getReplacements() (map[string][]files.Replacement, []files.Replacement, error) {
	filepats := viperTryGetKeys([]string{
		"replacements",
		"file-patterns",
		"filepatterns",
		"replacement-patterns",
	})
	reps := make([]files.Replacement, 0)
	courseReps := make(map[string][]files.Replacement)
	coursePats := viperTryGetKeys([]string{"course-replacements", "course_replacements"})

	err := errs.Pair(
		mapstructure.Decode(filepats, &reps),
		mapstructure.Decode(coursePats, &courseReps),
	)
	return upperMapKeys(courseReps), reps, err
}

func viperTryGetKeys(keys []string) interface{} {
	var result interface{}
	for _, key := range keys {
		result = viper.Get(key)
		if result != nil {
			break
		}
	}
	return result
}

func upperMapKeys(m map[string][]files.Replacement) map[string][]files.Replacement {
	cp := make(map[string][]files.Replacement)
	for key, val := range m {
		cp[strings.ToUpper(key)] = val
	}
	return cp
}
