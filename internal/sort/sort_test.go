package sort

import (
	"testing"

	"github.com/houseabsolute/detest"
	"golang.org/x/text/language"
)

type testCase struct {
	name   string
	input  []string
	expect []string
	params SortParams
}

var textSortTests = []testCase{
	{
		"locale-free ASCII text",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"And", "above", "all", "bears", "go", "home"},
		SortParams{
			language.Und,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"locale-free ASCII text, case-insensitive",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"above", "all", "And", "bears", "go", "home"},
		SortParams{
			language.Und,
			true,
			false,
			UnixPaths,
		},
	},
	{
		"locale-free ASCII text, reversed",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"home", "go", "bears", "all", "above", "And"},
		SortParams{
			language.Und,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"locale-free ASCII text, case-insensitive, reversed",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"home", "go", "bears", "And", "all", "above"},
		SortParams{
			language.Und,
			true,
			true,
			UnixPaths,
		},
	},
	{
		"en-US text",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"above", "all", "And", "bears", "go", "home"},
		SortParams{
			language.English,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"en-US text, reversed",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"home", "go", "bears", "And", "all", "above"},
		SortParams{
			language.English,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"German text",
		[]string{"zoo", "foo", "öoo"},
		[]string{"foo", "öoo", "zoo"},
		SortParams{
			language.German,
			false,
			false,

			UnixPaths,
		},
	},
	{
		"German text, reversed",
		[]string{"zoo", "foo", "öoo"},
		[]string{"zoo", "öoo", "foo"},
		SortParams{
			language.German,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"Swedish text",
		[]string{"zoo", "foo", "öoo"},
		[]string{"foo", "zoo", "öoo"},
		SortParams{
			language.Swedish,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"Swedish text, reversed",
		[]string{"zoo", "foo", "öoo"},
		[]string{"öoo", "zoo", "foo"},
		SortParams{
			language.Swedish,
			false,
			true,
			UnixPaths,
		},
	},
}

func Test_textSort(t *testing.T) {
	for _, test := range textSortTests {
		t.Run(test.name, func(t *testing.T) {
			//nolint:scopelint
			testOneCase(t, test, textSort)
		})
	}
}

var numberedTextSortTests = []testCase{
	{
		"numbered locale-free ASCII text",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"0. bears", "1. all", "2. home", "5. And", "15 - above", "120001 go"},
		SortParams{
			language.Und,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"numbered locale-free ASCII text, case-insensitive",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"0. bears", "1. all", "2. home", "5. And", "15 - above", "120001 go"},
		SortParams{
			language.Und,
			true,
			false,
			UnixPaths,
		},
	},
	{
		"numbered locale-free ASCII text, reversed",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"120001 go", "15 - above", "5. And", "2. home", "1. all", "0. bears"},
		SortParams{
			language.Und,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"numbered locale-free ASCII text, case-insensitive, reversed",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"120001 go", "15 - above", "5. And", "2. home", "1. all", "0. bears"},
		SortParams{
			language.Und,
			true,
			true,
			UnixPaths,
		},
	},
	{
		"German text",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"1. foo", "2. öoo", "2. zoo", "3. zoo"},
		SortParams{
			language.German,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"German text, reversed",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"3. zoo", "2. zoo", "2. öoo", "1. foo"},
		SortParams{
			language.German,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"Swedish text",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"1. foo", "2. zoo", "2. öoo", "3. zoo"},
		SortParams{
			language.Swedish,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"Swedish text, reversed",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"3. zoo", "2. öoo", "2. zoo", "1. foo"},
		SortParams{
			language.Swedish,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"mixed numbered and unnumbered",
		[]string{"10. x", "aloe", "27. bar", "love", "1. hello"},
		[]string{"1. hello", "10. x", "27. bar", "aloe", "love"},
		SortParams{
			language.Und,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"mixed numbered and unnumbered, reversed",
		[]string{"10. x", "aloe", "27. bar", "love", "1. hello"},
		[]string{"love", "aloe", "27. bar", "10. x", "1. hello"},
		SortParams{
			language.Und,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"numbers are decimals",
		[]string{"10.1 - x", "27.2314 - bar", "1.00 - hello"},
		[]string{"1.00 - hello", "10.1 - x", "27.2314 - bar"},
		SortParams{
			language.Und,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"numbers are decimals, reversed",
		[]string{"10.1 - x", "27.2314 - bar", "1.00 - hello"},
		[]string{"27.2314 - bar", "10.1 - x", "1.00 - hello"},
		SortParams{
			language.Und,
			false,
			true,
			UnixPaths,
		},
	},
}

func Test_numberedTextSort(t *testing.T) {
	for _, test := range append(textSortTests, numberedTextSortTests...) {
		t.Run(test.name, func(t *testing.T) {
			//nolint:scopelint
			testOneCase(t, test, numberedTextSort)
		})
	}
}

var pathSortTests = []testCase{
	{
		"path locale-free ASCII text",
		[]string{"/foo", "/bar", "baz/quux", "a/q", "C:\\", "/X", "/A"},
		[]string{"/A", "/X", "/bar", "/foo", "C:\\", "a/q", "baz/quux"},
		SortParams{
			language.Und,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"path locale-free ASCII text, case-insensitive",
		[]string{"/foo", "/bar", "baz/quux", "a/q", "C:\\", "/X", "/A"},
		[]string{"/A", "/bar", "/foo", "/X", "C:\\", "a/q", "baz/quux"},
		SortParams{
			language.Und,
			true,
			false,
			UnixPaths,
		},
	},
	{
		"path locale-free ASCII text, reversed",
		[]string{"/foo", "/bar", "baz/quux", "a/q", "C:\\", "/X", "/A"},
		[]string{"baz/quux", "a/q", "C:\\", "/foo", "/bar", "/X", "/A"},
		SortParams{
			language.Und,
			false,
			true,
			UnixPaths,
		},
	},
	{
		"path locale-free ASCII text, case-insensitive, reversed",
		[]string{"/foo", "/bar", "baz/quux", "a/q", "C:\\", "/X", "/A"},
		[]string{"baz/quux", "a/q", "C:\\", "/X", "/foo", "/bar", "/A"},
		SortParams{
			language.Und,
			true,
			true,
			UnixPaths,
		},
	},
	{
		"path depth sorts before path content",
		[]string{"/zzz", "/bbb", "/xxx/a", "/aaaaaa/q/r"},
		[]string{"/bbb", "/zzz", "/xxx/a", "/aaaaaa/q/r"},
		SortParams{
			language.Und,
			false,
			false,
			UnixPaths,
		},
	},
	{
		"path depth sorts before path content, reversed",
		[]string{"/zzz", "/bbb", "/xxx/a", "/aaaaaa/q/r"},
		[]string{"/aaaaaa/q/r", "/xxx/a", "/zzz", "/bbb"},
		SortParams{
			language.Und,
			false,
			true,
			UnixPaths,
		},
	},
}

func Test_pathSort(t *testing.T) {
	for _, test := range pathSortTests {
		t.Run(test.name, func(t *testing.T) {
			//nolint:scopelint
			testOneCase(t, test, pathSort)
		})
	}
}

func testOneCase(t *testing.T, test testCase, sorter sortFunc) {
	d := detest.New(t)
	// If the test fails and we haven't cloned then we cannot print
	// out debugging info with the original and the (improperly
	// sorted) list.
	clone := make([]string, len(test.input))
	copy(clone, test.input)
	err := sorter(clone, test.params)
	d.Is(err, nil, "no error from calling sorting func")
	d.Is(
		clone,
		d.Slice(func(st *detest.SliceTester) {
			st.End()
			for i, e := range test.expect {
				st.Idx(i, d.Equal(e))
			}
		}),
		"check sorted output",
	)
}
