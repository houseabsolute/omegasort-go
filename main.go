package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/eidolon/wordwrap"
	"golang.org/x/text/language"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var version = "0.0.1"

type omegasort struct {
	opts       *opts
	app        *kingpin.Application
	sort       sortType
	locale     language.Tag
	lineEnding []byte
}

type opts struct {
	sort            string
	locale          string
	caseInsensitive bool
	reverse         bool
	inPlace         bool
	toStdout        bool
	check           bool
	debug           bool
	file            string
}

func main() {
	o, err := new()
	if err != nil {
		o.app.FatalUsage("%s\n", err)
	}

	if err = o.run(); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error when sorting %s: %s\n", o.opts.file, err))
		os.Exit(2)
	}

	os.Exit(0)
}

func new() (*omegasort, error) {
	app := kingpin.New("omegasort", "The last text file sorting tool you'll ever need.").
		Author("Dave Rolsky <autarch@urth.org>").
		Version(version).
		UsageWriter(os.Stdout).
		UsageTemplate(kingpin.DefaultUsageTemplate + sortDocs())
	app.HelpFlag.Short('h')

	validSorts := []string{}
	for _, as := range availableSorts {
		validSorts = append(validSorts, as.name)
	}
	sortType := app.Flag("sort", "The type of sorting to use. See below for options.").Short('s').Required().
		HintOptions(validSorts...).Enum(validSorts...)
	locale := app.Flag("locale", "The locale to use for sorting. If this is not specified the sorting is in codepoint order.").Short('l').Default("").String()
	caseInsensitive := app.Flag("case-insensitive", "Sort case-insensitively. Note that many Unicode locales always do this so if you specify a locale you may get case-insensitive output regardless of this flag.").Short('c').Default("false").Bool()
	reverse := app.Flag("reverse", "Sort in reverse order.").Short('r').Default("false").Bool()
	inPlace := app.Flag("in-place", "Modify the file in place instead of making a backup.").Short('i').Default("false").Bool()
	toStdout := app.Flag("stdout", "Print the sorted output to stdout instead of making a new file.").Default("false").Bool()
	check := app.Flag("check", "Check that the file is sorted instead of sorting it. If it is not sorted the exit status will be 2.").Default("false").Bool()
	debug := app.Flag("debug", "Print out debugging info.").Default("false").Bool()
	file := app.Arg("file", "The file to sort.").Required().ExistingFile()

	appOpts := &opts{}
	o := &omegasort{
		app:  app,
		opts: appOpts,
	}

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return o, err
	}

	appOpts.sort = *sortType
	for _, as := range availableSorts {
		if as.name == appOpts.sort {
			o.sort = as
			break
		}
	}

	appOpts.locale = *locale
	appOpts.caseInsensitive = *caseInsensitive
	appOpts.reverse = *reverse
	appOpts.inPlace = *inPlace
	appOpts.toStdout = *toStdout
	appOpts.check = *check
	appOpts.debug = *debug
	appOpts.file = *file

	if appOpts.debug {
		fmt.Printf("opts = %+v\n", appOpts)
	}

	err = o.validateArgs()
	return o, err
}

func sortDocs() string {
	docs := "Sorting Options:\n"

	width := guessWidth(os.Stderr)
	longest := 0
	for _, as := range availableSorts {
		if len(as.name) > longest {
			longest = len(as.name)
		}
	}
	width -= longest
	width -= 4 // length of "  * "

	wrapper := wordwrap.Wrapper(width, false)

	for _, as := range availableSorts {
		indented := wrapper(as.description)
		docs += wordwrap.Indent(indented, fmt.Sprintf("  * %s - ", as.name), false) + "\n"
	}

	return docs
}

func (o *omegasort) validateArgs() error {
	if o.opts.locale != "" && !o.sort.supportsLocale {
		return fmt.Errorf("you cannot set a locale when sorting by %s", o.sort.name)
	}

	if o.opts.toStdout && o.opts.inPlace {
		return errors.New("you cannot set both --stdout and --in-place")
	}

	if o.opts.toStdout && o.opts.check {
		return errors.New("you cannot set both --stdout and --check")
	}

	if o.opts.inPlace && o.opts.check {
		return errors.New("you cannot set both --in-place and --check")
	}

	if o.opts.locale != "" {
		tag, err := language.Parse(o.opts.locale)
		if err != nil {
			return fmt.Errorf("could not find a locale matching %s: %s", o.opts.locale, err)
		}
		o.locale = tag
	}

	return nil
}

const firstChunk = 2048

func (o *omegasort) run() error {
	lines, err := o.readLines()
	err = o.sort.sortFunc(lines, o.locale, o.opts.caseInsensitive, o.opts.reverse)
	if err != nil {
		return err
	}

	out, err := o.outputFile()
	if err != nil {
		return err
	}

	for _, l := range lines {
		_, err := out.WriteString(l)
		if err != nil {
			return err
		}
		_, err = out.Write(o.lineEnding)
		if err != nil {
			return err
		}
	}

	if !o.opts.toStdout {
		err := o.moveFiles(out.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *omegasort) readLines() ([]string, error) {
	err := o.determineLineEnding()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(o.opts.file)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(o.splitFunc())

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

var crlf = []byte{'\r', 'n'}
var cr = []byte{'\r'}
var nl = []byte{'\n'}

func (o *omegasort) determineLineEnding() error {
	file, err := os.Open(o.opts.file)
	if err != nil {
		return err
	}

	buf := make([]byte, firstChunk)
	_, err = io.ReadAtLeast(file, buf, firstChunk)
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("could not read any data from %s", o.opts.file)
		}
		// If we got ErrUnexpectedEOF that just means the file is smaller than
		// firstChunk, which is fine.
		if err != io.ErrUnexpectedEOF {
			return fmt.Errorf("error trying to read data from %s", o.opts.file)
		}
	}

	if bytes.Contains(buf, crlf) {
		o.lineEnding = crlf
	} else if bytes.Contains(buf, cr) {
		o.lineEnding = cr
	} else if bytes.Contains(buf, nl) {
		o.lineEnding = nl
	} else {
		return fmt.Errorf("could not determine line ending from reading first %d bytes of %s", firstChunk, o.opts.file)
	}

	return nil
}

func (o *omegasort) splitFunc() bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, o.lineEnding); i >= 0 {
			return i + len(o.lineEnding), data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}
}

func (o *omegasort) outputFile() (*os.File, error) {
	if o.opts.toStdout {
		return os.Stdout, nil
	}

	return ioutil.TempFile("", "omegasort")
}

func (o *omegasort) moveFiles(from string) error {
	if !o.opts.inPlace {
		err := os.Rename(o.opts.file, o.opts.file+".bak")
		if err != nil {
			return nil
		}
	}

	return os.Rename(from, o.opts.file)
}
