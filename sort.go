package main

import (
	"regexp"
	"sort"
	"strconv"

	"golang.org/x/text/cases"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type sortType struct {
	name           string
	description    string
	supportsLocale bool
	sortFunc       func(lines []string, locale language.Tag, caseInsensitive, reverse bool) error
}

var availableSorts = []sortType{
	{
		"text",
		"sort the file as text according to the specified locale",
		true,
		textSort,
	},
	{
		"numbered-text",
		"sort the file assuming that each line starts with a numeric prefix, then fall back to sorting by text according to the specified locale",
		true,
		numberedTextSort,
	},
	{
		"datetime-text",
		"sort the file assuming that each line starts with a date or datetime prefix, then fall back to sorting by text according to the specified locale",
		true,
		datetimeTextSort,
	},
	{
		"path",
		"sort the file assuming that each line is a path, sorted so that deeper paths come after shorter",
		true,
		pathSort,
	},
	{
		"ip",
		"sort the file assuming that each line is an IP address",
		false,
		ipSort,
	},
}

func textSort(lines []string, locale language.Tag, caseInsensitive, reverse bool) error {
	comparer := stringComparer(locale, caseInsensitive, reverse)
	sort.Slice(lines, func(i, j int) bool { return comparer(lines[i], lines[j]) })
	return nil
}

var numberedTextRE = regexp.MustCompile(`\A([0-9]+)?(.+)\z`)

func numberedTextSort(lines []string, locale language.Tag, caseInsensitive, reverse bool) error {
	comparer := stringComparer(locale, caseInsensitive, reverse)
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
			if matchI[1] != "" && matchJ[1] != "" {
				numI, err := strconv.ParseInt(matchI[1], 10, 64)
				numJ, err := strconv.ParseInt(matchJ[1], 10, 64)
				if err == nil && numI != numJ {
					less = boolPointer(numI < numJ)
				}
			} else if matchI[1] != "" {
				less = boolPointer(true)
			} else if matchJ[1] != "" {
				less = boolPointer(false)
			}
			if less != nil {
				if reverse {
					return !*less
				}
				return *less
			}

			return comparer(matchI[2], matchJ[2])
		},
	)

	return err
}

func boolPointer(val bool) *bool {
	return &val
}

func datetimeTextSort(lines []string, locale language.Tag, caseInsensitive, reverse bool) error {
	return nil
}

func pathSort(lines []string, locale language.Tag, caseInsensitive, reverse bool) error {
	return nil
}

func ipSort(lines []string, _ language.Tag, caseInsensitive, reverse bool) error {
	return nil
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
		} else {
			if reverse {
				return func(i, j string) bool { return i > j }
			}
			return func(i, j string) bool { return i < j }
		}
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
