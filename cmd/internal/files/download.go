package files

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/harrybrwn/go-canvas"
	"github.com/pkg/errors"
)

// Download will download a canvas file and write it to
// a file named by filename if it does not already exist.
func Download(
	file io.WriterTo,
	filename string,
	stdout, stderr io.Writer,
) (err error) {
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
		// io.Discard is usually set when verbose output is turned off
		// we want to output the "Downloaded" statement anyways
		if stdout == ioutil.Discard {
			fmt.Printf("Downloaded %s\n", filename)
		} else {
			fmt.Fprintf(stdout, "Downloaded %s\n", filename)
		}
		log.Printf("Downloaded %s\n", filename)
	}()
	fmt.Fprintf(stdout, "Fetching %s\n", filename)
	_, err = file.WriteTo(osfile) // download the contents to the file
	return err
}

// NewDownloader creates a new CourseDownloader
func NewDownloader(basedir string) *CourseDownloader {
	return &CourseDownloader{
		Stdout:  ioutil.Discard,
		Stderr:  os.Stderr,
		wg:      new(sync.WaitGroup),
		basedir: basedir,
	}
}

// DownloaderFromWG will create a course downloader form an existing waitgroup.
func DownloaderFromWG(basedir string, wg *sync.WaitGroup) *CourseDownloader {
	return &CourseDownloader{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		wg:      wg,
		basedir: basedir,
	}
}

// CourseDownloader will download files from a
// canvas course.
type CourseDownloader struct {
	Stdout, Stderr io.Writer
	wg             *sync.WaitGroup
	basedir        string
}

// Wait calls wait on the internal waitgroup
func (cd *CourseDownloader) Wait() {
	cd.wg.Wait()
}

// Download will download all the files for a course and perform the
// replacement patterns.
func (cd *CourseDownloader) Download(course *canvas.Course, replacements []Replacement) error {
	pairs := cd.filesGenerator(course)
	for pair := range pairs {
		if pair.err != nil {
			return pair.err
		}
		cd.wg.Add(1)
		go cd.downloadFile(pair.file, pair.path, replacements)
	}
	return nil
}

// CheckReplacements will print the result of replacement patterns
// on the files in a course.
func (cd *CourseDownloader) CheckReplacements(
	course *canvas.Course,
	reps []Replacement,
) (err error) {
	pairs := cd.filesGenerator(course)
	for pair := range pairs {
		if pair.err != nil {
			return pair.err
		}
		result, err := DoReplacements(reps, pair.path)
		if err != nil {
			return err
		}
		relFullpath, err := filepath.Rel(cd.basedir, pair.path)
		if err != nil {
			return err
		}
		relResult, err := filepath.Rel(cd.basedir, result)
		if err != nil {
			return err
		}
		spaces := 100 - len(relFullpath)
		if spaces < 0 {
			spaces = 0
		}
		fmt.Printf("%s %s=> %s\n", relFullpath, strings.Repeat(" ", spaces), relResult)
	}
	return nil
}

type filePathPair struct {
	path string
	file *canvas.File
	err  error
}

func (cd *CourseDownloader) filesGenerator(course *canvas.Course) <-chan *filePathPair {
	var (
		ch     = make(chan *filePathPair)
		dirmap = make(map[int]string)
		rel    string
		err    error
	)
	go func() {
		defer close(ch)
		course.SetErrorHandler(func(e error) error {
			if e != nil {
				fmt.Fprintf(os.Stderr, "Warning: not authorized to get files for \"%s\"\n", course.Name)
			}
			return e
		})
		// perm, err := course.Permissions()
		// if err != nil {
		// 	panic(err)
		// }
		for file := range course.Files() {
			pair := &filePathPair{file: file}
			folder, ok := dirmap[file.FolderID]
			if !ok {
				parent, err := file.ParentFolder()
				if err != nil {
					pair.err = err
					goto SendFile
				}
				folder = parent.FullName
				dirmap[file.FolderID] = folder
			}
			rel, err = filepath.Rel("course files", folder)
			if err != nil {
				pair.err = err
				goto SendFile
			}
			pair.path = filepath.Join(cd.basedir, course.Name, rel, file.Filename)
		SendFile:
			ch <- pair
		}
	}()
	return ch
}

func (cd *CourseDownloader) downloadFile(file *canvas.File, path string, reps []Replacement) error {
	defer cd.wg.Done()
	fullpath, err := DoReplacements(reps, path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(fullpath)
	if err := mkdir(dir); err != nil {
		return err
	}
	return Download(file, fullpath, cd.Stdout, cd.Stderr)
}

func relpath(base, p string) string {
	rel, err := filepath.Rel(base, p)
	if err != nil {
		panic(err)
	}
	return rel
}

func mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0775)
	}
	return nil
}
