package sort

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"

	"github.com/houseabsolute/omegasort/internal/posixpath"
	"github.com/houseabsolute/omegasort/internal/winpath"
	"golang.org/x/text/cases"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type pathType int

const (
	UnixPaths pathType = iota
	WindowsPaths
)

type SortParams struct {
	Locale          language.Tag
	CaseInsensitive bool
	Reverse         bool
	PathType        pathType
}

type sortFunc func(lines []string, p SortParams) error

type Type struct {
	Name             string
	Description      string
	SupportsLocale   bool
	SupportsPathType bool
	SortFunc         sortFunc
}

var AvailableSorts = []Type{
	{
		"text",
		"sort the file as text according to the specified locale",
		true,
		false,
		textSort,
	},
	{
		"numbered-text",
		"sort the file assuming that each line starts with a numeric prefix," +
			" then fall back to sorting by text according to the specified locale",
		true,
		false,
		numberedTextSort,
	},
	{
		"datetime-text",
		"sort the file assuming that each line starts with a date or datetime prefix," +
			" then fall back to sorting by text according to the specified locale",
		true,
		false,
		datetimeTextSort,
	},
	{
		"path",
		"sort the file assuming that each line is a path," +
			" sorted so that deeper paths come after shorter",
		true,
		true,
		pathSort,
	},
	{
		"ip",
		"sort the file assuming that each line is an IP address",
		false,
		false,
		ipSort,
	},
	{
		"network",
		"sort the file assuming that each line is a network",
		false,
		false,
		networkSort,
	},
}

func textSort(lines []string, p SortParams) error {
	comparer := stringComparer(p.Locale, p.CaseInsensitive, p.Reverse)
	sort.Slice(lines, func(i, j int) bool { return comparer(lines[i], lines[j]) })
	return nil
}

var numberedTextRE = regexp.MustCompile(`\A([0-9]+(?:\.[0-9]+)?)?(.+)\z`)

func numberedTextSort(lines []string, p SortParams) error {
	comparer := stringComparer(p.Locale, p.CaseInsensitive, p.Reverse)
	var err error
	sort.Slice(
		lines,
		func(i, j int) bool {
			if err != nil {
				return false
			}

			matchI := numberedTextRE.FindStringSubmatch(lines[i])
			matchJ := numberedTextRE.FindStringSubmatch(lines[j])
			var less *bool
			switch {
			case matchI[1] != "" && matchJ[1] != "":
				numI, errI := strconv.ParseFloat(matchI[1], 64)
				numJ, errJ := strconv.ParseFloat(matchJ[1], 64)
				if errI == nil && errJ == nil && numI != numJ {
					less = boolPointer(numI < numJ)
				}
			case matchI[1] != "":
				less = boolPointer(true)
			case matchJ[1] != "":
				less = boolPointer(false)
			}
			if less != nil {
				if p.Reverse {
					return !*less
				}
				return *less
			}

			return comparer(matchI[2], matchJ[2])
		},
	)

	return err
}

func datetimeTextSort(lines []string, p SortParams) error {
	return nil
}

func pathSort(lines []string, p SortParams) error {
	comparer := stringComparer(p.Locale, p.CaseInsensitive, p.Reverse)
	var err error

	sort.Slice(
		lines,
		func(i, j int) bool {
			if err != nil {
				return false
			}

			var less *bool
			// Absolute paths sort before relative
			if isAbs(lines[i], p.PathType) && !isAbs(lines[j], p.PathType) {
				less = boolPointer(true)
			} else if !isAbs(lines[i], p.PathType) && isAbs(lines[j], p.PathType) {
				less = boolPointer(false)
			}
			if less != nil {
				if p.Reverse {
					return !*less
				}
				return *less
			}

			elemI := splitPath(lines[i], p.PathType)
			elemJ := splitPath(lines[j], p.PathType)

			if p.PathType == WindowsPaths {
				iIs := isDriveLetter(elemI[0])
				jIs := isDriveLetter(elemJ[0])
				switch {
				case iIs && !jIs:
					less = boolPointer(true)
				case !iIs && jIs:
					less = boolPointer(false)
				case iIs && jIs && elemI[0] != elemJ[0]:
					less = boolPointer(elemI[0] < elemJ[0])
				}
			}
			if less != nil {
				if p.Reverse {
					return !*less
				}
				return *less
			}

			if len(elemI) != len(elemJ) {
				less = boolPointer(len(elemI) < len(elemJ))
			}
			if less != nil {
				if p.Reverse {
					return !*less
				}
				return *less
			}

			for x := range elemI {
				if elemI[x] != elemJ[x] {
					return comparer(elemI[x], elemJ[x])
				}
			}

			return true
		},
	)

	return err
}

func splitPath(path string, typ pathType) []string {
	if typ == WindowsPaths {
		return winpath.SplitElem(path)
	}

	return posixpath.SplitElem(path)
}

func isAbs(path string, typ pathType) bool {
	if typ == WindowsPaths {
		return winpath.IsAbs(path)
	}

	return posixpath.IsAbs(path)
}

var driveLetterRE = regexp.MustCompile(`^[A-Z]:\\`)

func isDriveLetter(elem string) bool {
	return driveLetterRE.MatchString(elem)
}

func ipSort(lines []string, p SortParams) error {
	var err error
	sort.Slice(
		lines,
		func(i, j int) bool {
			if err != nil {
				return false
			}

			addrI := net.ParseIP(lines[i])
			if addrI == nil {
				err = fmt.Errorf("invalid IP address '%s' at line %d", lines[i], i+1)
				return false
			}

			addrJ := net.ParseIP(lines[j])
			if addrJ == nil {
				err = fmt.Errorf("invalid IP address '%s' at line %d", lines[j], j+1)
				return false
			}

			var less *bool
			if len(addrI) != len(addrJ) {
				less = boolPointer(len(addrI) < len(addrJ))
			}

			if less == nil {
				less = boolPointer(bytes.Compare(addrI, addrJ) < 0)
			}

			if p.Reverse {
				return !*less
			}
			return *less
		},
	)

	return err
}

func networkSort(lines []string, p SortParams) error {
	return nil
}

func boolPointer(val bool) *bool {
	return &val
}

func stringComparer(locale language.Tag, caseInsensitive, reverse bool) func(i, j string) bool {
	if locale == language.Und {
		if caseInsensitive {
			caser := cases.Fold()
			if reverse {
				return func(i, j string) bool {
					return caser.String(i) > caser.String(j)
				}
			}
			return func(i, j string) bool {
				return caser.String(i) < caser.String(j)
			}
		}

		if reverse {
			return func(i, j string) bool { return i > j }
		}
		return func(i, j string) bool { return i < j }
	}

	opts := []collate.Option{}
	if caseInsensitive {
		opts = append(opts, collate.IgnoreCase)
	}

	coll := collate.New(locale, opts...)
	if reverse {
		return func(i, j string) bool { return coll.CompareString(i, j) == 1 }
	}

	return func(i, j string) bool { return coll.CompareString(i, j) == -1 }
}
