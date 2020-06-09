package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type updateCmd struct {
	wg     *sync.WaitGroup
	cmd    *cobra.Command
	errors chan error

	out, err io.Writer

	all, verbose bool
	basedir      string
	testPatters  bool
	sortBy       []string
}

func newUpdateCmd() *cobra.Command {
	uc := &updateCmd{
		wg:     &sync.WaitGroup{},
		out:    ioutil.Discard,
		err:    os.Stderr,
		errors: make(chan error),
		cmd: &cobra.Command{
			Use:   "update",
			Short: "Download all the files from canvas",
		},
		all:     false,
		verbose: false,
		sortBy:  []string{"created_at"},
		basedir: os.ExpandEnv(viper.GetString("basedir")),
	}
	flags := uc.cmd.Flags()
	flags.BoolVarP(&uc.all, "all", "a", uc.all, "download files from all courses, defaults to only active courses")
	flags.BoolVarP(&uc.verbose, "verbose", "v", uc.verbose, "run update in verbose mode (prints out files)")
	flags.BoolVar(&uc.testPatters, "test-patterns", uc.testPatters, "test the replacement patterns from the config file")
	flags.StringVar(&uc.basedir, "base-dir", uc.basedir, "base directory for file downloads")
	flags.StringArrayVarP(&uc.sortBy, "sort-by", "s", uc.sortBy, "select the file sorting methods")
	uc.cmd.RunE = uc.run
	return uc.cmd
}

func (uc *updateCmd) run(cmd *cobra.Command, args []string) (err error) {
	courses, err := internal.GetCourses(uc.all)
	if err != nil {
		return err
	}
	if uc.verbose {
		uc.out = os.Stdout
	}
	coursereps, replacements, err := getReplacements()
	if err != nil {
		return err
	}

	var fn = uc.downloadCourseFiles
	if uc.testPatters {
		fn = uc.checkPatterns
	}

	uc.wg.Add(len(courses))
	for _, course := range courses {
		reps, ok := coursereps[course.CourseCode]
		if !ok {
			reps = replacements
		} else {
			reps = append(replacements, reps...)
		}
		go fn(course, reps)
	}
	uc.wg.Wait()
	fmt.Println("done.")
	return nil
}

type downloadCourseFunc func(fullpath string, f *canvas.File, wg *sync.WaitGroup) error

func (uc *updateCmd) downloadCourseFiles(c *canvas.Course, replacements []replacement) error {
	defer uc.wg.Done()
	return uc.courseFiles(c, func(fullpath string, file *canvas.File, wg *sync.WaitGroup) error {
		defer wg.Done()
		fullpath, err := doReplacements(replacements, fullpath, false)
		if err != nil {
			return err
		}
		dir := filepath.Dir(fullpath)
		if err := internal.Mkdir(dir); err != nil {
			return err
		}
		wg.Add(1)
		go download(file, fullpath, uc.out, uc.err, wg)
		return nil
	})
}

func (uc *updateCmd) checkPatterns(c *canvas.Course, patterns []replacement) error {
	defer uc.wg.Done()
	return uc.courseFiles(c, func(fullpath string, _ *canvas.File, wg *sync.WaitGroup) error {
		defer wg.Done()
		result, err := doReplacements(patterns, fullpath, false)
		if err != nil {
			return err
		}
		relFullpath := relpath(uc.basedir, fullpath)
		relResult := relpath(uc.basedir, result)

		spaces := 100 - len(relFullpath)
		if spaces < 0 {
			spaces = 0
		}
		fmt.Printf("%s %s=> %s\n", relFullpath, strings.Repeat(" ", spaces), relResult)
		return nil
	})
}

func download(
	file io.WriterTo,
	filename string,
	stdout, stderr io.Writer,
	wg *sync.WaitGroup,
) (err error) {
	defer wg.Done()
	osfile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if os.IsExist(err) {
		fmt.Fprintf(stdout, "file exists %s\n", filename)
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "file error")
	}
	defer func() {
		if e := osfile.Close(); e != nil && err == nil {
			err = e
		}
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
		fmt.Fprintf(stdout, "Downloaded %s\n", filename)
	}()
	fmt.Fprintf(stdout, "Fetching %s\n", filename)
	_, err = file.WriteTo(osfile) // download the contents to the file
	return err
}

func downloadURL(file, url string, stdout, stderr io.Writer, wg *sync.WaitGroup) (err error) {
	defer func() {
		if err != nil {
			fmt.Fprintf(stderr, "Error: %s", err.Error())
		}
		wg.Done()
	}()
	_, err = os.Stat(file)
	if !os.IsNotExist(err) {
		fmt.Fprintf(stdout, "file exists %s\n", file)
		return nil
	}
	fmt.Fprintf(stdout, "getting %s\n", file)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, raw, 0644)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Downloaded %s\n", file)
	return nil
}

func (uc *updateCmd) courseFiles(
	c *canvas.Course,
	fn downloadCourseFunc,
) error {
	var dirmap = make(map[int]string)
	for file := range c.Files(canvas.SortOpt(uc.sortBy...)) {
		folder, ok := dirmap[file.FolderID]
		if !ok {
		GetParent:
			parent, err := file.ParentFolder()
			if canvas.IsRateLimit(err) {
				// retry after half a second
				time.Sleep(time.Second / 2)
				goto GetParent
			}
			if err != nil {
				return err
			}
			folder = parent.FullName
			dirmap[file.FolderID] = folder
		}
		rel, err := filepath.Rel("course files", folder)
		if err != nil {
			return err
		}
		path := filepath.Join(uc.basedir, c.Name, rel)
		fp := filepath.Join(path, file.Filename)
		uc.wg.Add(1)
		go fn(fp, file, uc.wg)
	}
	return nil
}

type replacement struct {
	Pattern     string
	Replacement string
	Lower       bool
}

func doReplacements(patterns []replacement, fullpath string, lower bool) (string, error) {
	var result = fullpath
	for _, pattern := range patterns {
		pat, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			return result, err
		}
		if pattern.Lower {
			result = pat.ReplaceAllStringFunc(result, func(s string) string {
				return strings.ToLower(pat.ReplaceAllString(s, pattern.Replacement))
			})
		} else {
			result = pat.ReplaceAllString(result, pattern.Replacement)
		}
	}
	return result, nil
}

func getReplacements() (map[string][]replacement, []replacement, error) {
	filepats := viperTryGetKeys([]string{
		"replacements",
		"file-patterns",
		"filepatterns",
		"replacement-patterns",
	})
	reps := make([]replacement, 0)
	courseReps := make(map[string][]replacement)
	coursePats := viper.Get("course-replacements")

	err := errs.Pair(
		mapstructure.Decode(filepats, &reps),
		mapstructure.Decode(coursePats, &courseReps),
	)
	return upperMapKeys(courseReps), reps, err
}

func relpath(base, p string) string {
	rel, err := filepath.Rel(base, p)
	if err != nil {
		panic(err)
	}
	return rel
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

func upperMapKeys(m map[string][]replacement) map[string][]replacement {
	cp := make(map[string][]replacement)
	for key, val := range m {
		cp[strings.ToUpper(key)] = val
	}
	return cp
}
