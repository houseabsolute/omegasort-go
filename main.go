package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/eidolon/wordwrap"
	"github.com/houseabsolute/omegasort/internal/sorters"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/text/language"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var version = "0.0.3"

type omegasort struct {
	opts       *opts
	app        *kingpin.Application
	sort       sorters.Approach
	locale     language.Tag
	lineEnding []byte
}

type opts struct {
	sort            string
	locale          string
	caseInsensitive bool
	reverse         bool
	windows         bool
	inPlace         bool
	toStdout        bool
	check           bool
	debug           bool
	file            string
}

var errNotSorted = errors.New("file is not sorted")

func main() {
	o, err := new()
	if err != nil {
		log.Fatal(err)
	}

	if err = o.run(); err != nil {
		if err == errNotSorted {
			_, err = os.Stderr.WriteString(fmt.Sprintf("The %s file is not sorted\n", o.opts.file))
			if err != nil {
				panic(err)
			}
			os.Exit(1)
		}

		_, err = os.Stderr.WriteString(fmt.Sprintf("error when sorting %s: %s\n", o.opts.file, err))
		if err != nil {
			panic(err)
		}
		os.Exit(2)
	}

	os.Exit(0)
}

func new() (*omegasort, error) {
	sd, err := sortDocs()
	if err != nil {
		return nil, err
	}

	app := kingpin.New("omegasort", "The last text file sorting tool you'll ever need.").
		Author("Dave Rolsky <autarch@urth.org>").
		Version(version).
		UsageWriter(os.Stdout).
		UsageTemplate(kingpin.DefaultUsageTemplate + sd)
	app.HelpFlag.Short('h')

	validSorts := []string{}
	for _, as := range sorters.AvailableSorts {
		validSorts = append(validSorts, as.Name)
	}
	// We cannot set .Required for this flag (or any others) because we want
	// the --docs flag to work without any other flags needed.
	sortType := app.Flag(
		"sort",
		"The type of sorting to use. See below for options.",
	).Short('s').HintOptions(validSorts...).Enum(validSorts...)
	locale := app.Flag(
		"locale",
		"The locale to use for sorting. If this is not specified the sorting is in codepoint order.",
	).Short('l').Default("").String()
	caseInsensitive := app.Flag(
		"case-insensitive",
		"Sort case-insensitively. Note that many locales always do this so if you specify"+
			" a locale you may get case-insensitive output regardless of this flag.").
		Short('c').Default("false").Bool()
	reverse := app.Flag(
		"reverse",
		"Sort in reverse order.",
	).Short('r').Default("false").Bool()
	windows := app.Flag(
		"windows",
		"Parse paths as Windows paths for path sort.",
	).Default("false").Bool()
	inPlace := app.Flag(
		"in-place",
		"Modify the file in place instead of making a backup.",
	).Short('i').Default("false").Bool()
	toStdout := app.Flag(
		"stdout",
		"Print the sorted output to stdout instead of making a new file.",
	).Default("false").Bool()
	check := app.Flag(
		"check",
		"Check that the file is sorted instead of sorting it. If it is not sorted the exit status will be 2.",
	).Default("false").Bool()
	debug := app.Flag(
		"debug",
		"Print out debugging info while running.",
	).Default("false").Bool()
	docs := app.Flag(
		"docs",
		"Print out extended sorting documentation.",
	).Default("false").Bool()
	file := app.Arg(
		"file",
		"The file to sort.",
	).ExistingFile()

	appOpts := &opts{}
	o := &omegasort{
		app:  app,
		opts: appOpts,
	}

	_, err = app.Parse(os.Args[1:])
	if err != nil {
		return o, err
	}

	if docs != nil && *docs {
		printExtendedDocs()
		os.Exit(0)
	}

	appOpts.sort = *sortType
	for _, as := range sorters.AvailableSorts {
		if as.Name == appOpts.sort {
			o.sort = as
			break
		}
	}

	appOpts.locale = *locale
	appOpts.caseInsensitive = *caseInsensitive
	appOpts.reverse = *reverse
	appOpts.windows = *windows
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

func sortDocs() (string, error) {
	docs := "Sorting Options:\n"

	width, err := getWidth()
	if err != nil {
		return "", err
	}

	longest := 0
	for _, as := range sorters.AvailableSorts {
		if len(as.Name) > longest {
			longest = len(as.Name)
		}
	}
	width -= longest
	width -= 4 // length of "  * "

	wrapper := wordwrap.Wrapper(width, false)

	for _, as := range sorters.AvailableSorts {
		indented := wrapper(as.Description)
		docs += wordwrap.Indent(indented, fmt.Sprintf("  * %s - ", as.Name), false) + "\n"
	}

	return docs, nil
}

func (o *omegasort) validateArgs() error {
	if o.opts.sort == "" {
		return errors.New("you must set a --sort method")
	}

	if o.opts.file == "" {
		return errors.New("you must pass a file to sort as the final argument")
	}

	if o.opts.locale != "" && !o.sort.SupportsLocale {
		return fmt.Errorf("you cannot set a locale when sorting by %s", o.sort.Name)
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

	if o.opts.windows && !o.sort.SupportsPathType {
		return fmt.Errorf("you cannot pass the --windows flag when sorting by %s", o.sort.Name)
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

//nolint: lll
var extendedSortDocs = `There are a number of different sorting methods available.

## Text

This sorts each line of the file as text without any special parsing. The exact sorting is determined by the --locale, --case-insensitive, and --reverse flags. See below for details on how locales work.

## Numbered Text

This assumes that each line of the file starts with a numeric value, optionally followed by non-numeric text.

Lines should not have any leading space before the number. The number can either be an integer (including 0) or a simple float (no scientific notation).

The lines will be sorted numerically first. If two lines have the same number they will be sorted by text as above.

Lines without numbers always sort after lines with numbers.

This sorting method accepts the --locale, --case-insensitive, and --reverse flags.

## Path Sort

Each line is treated as a path.

The paths are sorted by the following rules:

* Absolute paths come before relative.
* Paths are sorted by depth before sorting by the path content, so /z comes before /a/a.
* If you pass the --windows flag, then paths with drive letters are sorted based on the drive letter first. Paths with drive letters sort before paths without them.

This sorting method accepts the --locale, --case-insensitive, and --reverse flags in addition to the --windows flag.

## Datetime Sort

This sorting method assumes that each line starts with a date or datetime, without any space in it. That means datetimes need to be in a format like "2019-08-27T19:13:16".

Lines should not have any leading space before the datetime.

This sorting method accepts the --locale, --case-insensitive, and --reverse flags.

## IP Sort

This method assumes that each line is an IPv4 or IPv6 address (not a network).

The sorting method is the same as if each line were the corresponding integer for the address.

This sorting method accepts the --reverse flag.

## Network Sort

This method assumes that each line is an IPv4 or IPv6 network in CIDR notation.

If there are two networks with the same base address they are sorted with the larger network first (so 1.1.1.0/24 comes before 1.1.1.0/28).

This sorting method accepts the --reverse flag.

`

func printExtendedDocs() {
	width, err := getWidth()
	if err != nil {
		panic(err)
	}

	wrapper := wordwrap.Wrapper(width, false)

	lines := strings.Split(extendedSortDocs, "\n")

	for _, l := range lines {
		var err error
		_, err = os.Stdout.WriteString(wrapper(l) + "\n")
		if err != nil {
			panic(err)
		}
	}
}

const maxWidth = 90

func getWidth() (int, error) {
	width, _, err := terminal.GetSize(int(os.Stderr.Fd()))
	if width > maxWidth {
		width = maxWidth
	}

	if err == syscall.ENOTTY {
		return 80, nil
	}

	return width, err
}

const firstChunk = 2048

func (o *omegasort) run() error {
	p := sorters.SortParams{
		Locale:          o.locale,
		CaseInsensitive: o.opts.caseInsensitive,
		Reverse:         o.opts.reverse,
	}
	if o.opts.windows {
		p.PathType = sorters.WindowsPaths
	}

	lines, err := o.readLines()
	if err != nil {
		return err
	}

	sorter, errRef := o.sort.MakeSortFunc(&lines, p)
	if o.opts.check {
		ok := sort.SliceIsSorted(lines, sorter)
		if *errRef != nil {
			return *errRef
		}
		if !ok {
			return errNotSorted
		}

		return nil
	}

	sort.Slice(lines, sorter)
	if *errRef != nil {
		return *errRef
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
		err := o.updateFiles(out.Name())
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

	switch {
	case bytes.Contains(buf, crlf):
		o.lineEnding = crlf
	case bytes.Contains(buf, cr):
		o.lineEnding = cr
	case bytes.Contains(buf, nl):
		o.lineEnding = nl
	default:
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

func (o *omegasort) updateFiles(from string) error {
	if !o.opts.inPlace {
		bak := o.opts.file + ".bak"
		err := copy(o.opts.file, bak)
		if err != nil {
			return fmt.Errorf("error copying %s to %s: %w", o.opts.file, bak, err)
		}
	}

	if err := copy(from, o.opts.file); err != nil {
		return fmt.Errorf("error copying %s to %s: %w", from, o.opts.file, err)
	}

	if err := os.Remove(from); err != nil {
		return fmt.Errorf("error deleting %s: %w", from, err)
	}

	return nil
}

func copy(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", from, err)
	}
	// nolint:errcheck
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", to, err)
	}
	// nolint:errcheck
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
