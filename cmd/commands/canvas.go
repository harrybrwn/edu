package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/edu/pkg/term"
	"github.com/harrybrwn/go-canvas"
	"github.com/jaytaylor/html2text"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type fileFinder struct {
	contentType string
	search      string
	all         bool
}

func (ff *fileFinder) flagset() *pflag.FlagSet {
	flagset := pflag.NewFlagSet("", pflag.ExitOnError)
	ff.addToFlagSet(flagset)
	return flagset
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

func (ff *fileFinder) addToFlagSet(flagset *pflag.FlagSet) {
	flagset.BoolVarP(&ff.all, "all", "a", ff.all, "query files from all courses")
	flagset.StringVarP(&ff.contentType, "content-type", "c", "", "filter out files by content type (ex. application/pdf)")
	flagset.StringVar(&ff.search, "search", "", "search for files by name")
}

// CanvasCommands gets all the canvas commands
func canvasCommands(flags *opts.Global) []*cobra.Command {
	return []*cobra.Command{
		newFilesCmd(),
		newDueCmd(flags),
		newUploadCmd(),
		assignmentsCmd(),
	}
}

var (
	canvasCmd = &cobra.Command{
		Use:     "canvas",
		Aliases: []string{"canv", "ca"},
		Short:   "A small collection of helper commands for canvas",
	}
)

type dueDate struct {
	id, name string
	date     time.Time
}

type dueDates []dueDate

func (dd dueDates) Len() int {
	return len(dd)
}

func (dd dueDates) Swap(i, j int) {
	dd[i], dd[j] = dd[j], dd[i]
}

func (dd dueDates) Less(i, j int) bool {
	return dd[i].date.Before(dd[j].date)
}

func newDueCmd(flags *opts.Global) *cobra.Command {
	var nolinks, all bool
	dueCmd := &cobra.Command{
		Use:   "due",
		Short: "List all the due date on canvas.",
		RunE: func(cmd *cobra.Command, args []string) error {
			courses, err := internal.GetCourses(false)
			if err != nil {
				return internal.HandleAuthErr(err)
			}
			if len(args) > 0 {
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				for _, course := range courses {
					as, err := course.Assignment(id)
					if err != nil {
						continue
					}
					text, err := html2text.FromString(
						as.Description,
						html2text.Options{
							PrettyTables: true,
							OmitLinks:    nolinks,
						},
					)
					if err != nil {
						return err
					}
					fmt.Println(term.Colorf("%b %r", as.Name, as.DueAt.Local().String()))
					fmt.Println(text)
					return nil
				}
				return nil
			}

			// if the user has not given a crn, then we print out all the assignments
			var wg sync.WaitGroup
			wg.Add(len(courses))
			tab := internal.NewTable(cmd.OutOrStdout())
			internal.SetTableHeader(tab, []string{"id", "name", "due"}, !flags.NoColor)

			printer := &assignmentPrinter{
				w:   cmd.OutOrStdout(),
				tab: tab,
				wg:  &wg,
				all: all,
				now: time.Now(),
			}
			for _, course := range courses {
				go printer.printCourse(course)
			}
			wg.Wait()
			return nil
		},
	}
	dueCmd.Flags().BoolVar(&nolinks, "no-links", false, "hide links from assignment description")
	dueCmd.Flags().BoolVarP(&all, "all", "a", false, "show all the assignments")
	return dueCmd
}

type assignmentPrinter struct {
	w       io.Writer
	tab     *tablewriter.Table
	tableMu sync.Mutex
	wg      *sync.WaitGroup
	all     bool
	now     time.Time
}

func (p *assignmentPrinter) printCourse(course *canvas.Course) {
	var dates dueDates
	for as := range course.Assignments() {
		dueAt := as.DueAt.Local()
		if !p.all && dueAt.Before(p.now) {
			continue
		}
		dates = append(dates, dueDate{
			id:   strconv.Itoa(as.ID),
			name: as.Name,
			date: dueAt,
		})
	}
	sort.Sort(dates)

	// rendering
	p.tableMu.Lock()
	fmt.Fprintln(p.w, term.Colorf("  %m", course.Name))
	for _, d := range dates {
		p.tab.Append([]string{d.id, d.name, d.date.Format(time.RFC822)})
	}
	p.tab.Render()
	p.tab.ClearRows()
	fmt.Fprintf(p.w, "\n")

	// clean up
	p.tableMu.Unlock()
	p.wg.Done()
}

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
				return internal.HandleAuthErr(err)
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

func assignmentsCmd() *cobra.Command {
	var nolinks bool
	c := &cobra.Command{
		Use:     "assignments",
		Hidden:  true,
		Aliases: []string{"as"},
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			courses, err := internal.GetCourses(true)
			if err != nil {
				return internal.HandleAuthErr(err)
			}
			for _, course := range courses {
				as, err := course.Assignment(id)
				if err != nil {
					continue
				}
				text, err := html2text.FromString(
					as.Description,
					html2text.Options{
						PrettyTables: true,
						OmitLinks:    nolinks,
					},
				)
				if err != nil {
					return err
				}
				fmt.Println(term.Colorf("%b %r", as.Name, as.DueAt.Local().String()))
				fmt.Println(text)
				return nil
			}
			return fmt.Errorf("did not find assignment %d", id)
		},
	}
	c.Flags().BoolVar(&nolinks, "no-links", nolinks, "hide all links in the assignment description")
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
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()
	var opts []canvas.Option
	if dir != "" {
		opts = append(opts, canvas.Opt("parent_folder_path", dir))
	}
	_, err = canvas.UploadFile(uploadname, file, opts...)
	return internal.HandleAuthErr(err)
}

type iter interface {
	Next() html.TokenType
	Token() html.Token
}

func parseHTML(raw string) error {
	root, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return err
	}
	// html.Render(os.Stdout, root)
	for n := root.FirstChild; n != nil; n = n.NextSibling {
		traverse(n, 1)
	}
	return nil
}

func traverse(node *html.Node, depth int) {
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		// for i := 0; i < depth; i++ {
		// 	print("  ")
		// }
		// fmt.Print(n.Type, " ")

		switch n.Type {
		case html.TextNode:
			fmt.Print(n.Data)
		case html.ElementNode:
			switch n.DataAtom {
			case atom.P:
				printPTag(n, depth)
			case atom.A:
				printATag(n)
			case atom.Br:
				print("\n")
			}
		}
		// println()

		traverse(n, depth+1)
	}
}

func printATag(node *html.Node) {
	println()
	print(node.FirstChild.Data)
}

func printPTag(node *html.Node, depth int) {
	// fmt.Printf("'%s'", node.Data)
	for n := node.FirstChild; n != nil; n = n.NextSibling {
	}
}
