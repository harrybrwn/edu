package commands

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

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
		newUploadCmd(),
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
			internal.SetTableHeader(tab, []string{"name", "due"}, true)
			for _, course := range courses {
				for as := range course.Assignments() {
					tab.Append([]string{as.Name, as.DueAt.String()})
				}
			}
			tab.Render()
			return nil
		},
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

func newUploadCmd() *cobra.Command {
	var (
		file       string
		folderPath string
		uploadAs   string
	)
	c := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file to canvas user account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && file == "" {
				file = args[0]
			}
			if file == "" {
				return errors.New("no filename given")
			}
			if folderPath != "" {
				uploadAs = path.Join(folderPath, file)
			}
			if uploadAs == "" {
				_, filename := filepath.Split(file)
				uploadAs = "/" + filename
			}
			return upload(file, uploadAs)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return args, cobra.ShellCompDirectiveDefault
		},
	}
	if err := c.MarkZshCompPositionalArgumentFile(1, "*"); err != nil {
		fmt.Fprintf(os.Stderr, "Completion error: %v", err)
	}
	flags := c.Flags()
	flags.StringVarP(&file, "file", "f", "", "give a filename for the file to upload")
	flags.StringVarP(&uploadAs, "upload-as", "u", "", "rename the file being uploaded")
	flags.StringVarP(&folderPath, "folder", "d", "", "set the folder path to upload the file to")
	return c
}

func upload(filename, uploadname string) (err error) {
	dir, uploadname := filepath.Split(uploadname)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		// set the return value in case of close error
		if e := file.Close(); e != nil {
			err = e
		}
	}()
	var opts []canvas.Option
	if dir != "" {
		opts = append(opts, canvas.Opt("parent_folder_path", dir))
	}
	_, err = canvas.UploadFile(uploadname, file, opts...)
	return err
}
