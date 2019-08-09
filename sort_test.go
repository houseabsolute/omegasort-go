package main

import (
	"testing"

	"github.com/houseabsolute/detest"
	"golang.org/x/text/language"
)

type testCase struct {
	name            string
	input           []string
	expect          []string
	locale          language.Tag
	caseInsensitive bool
	reverse         bool
}

var textSortTests = []testCase{
	{
		"locale-free ASCII text",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"And", "above", "all", "bears", "go", "home"},
		language.Und,
		false,
		false,
	},
	{
		"locale-free ASCII text, case-insensitive",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"above", "all", "And", "bears", "go", "home"},
		language.Und,
		true,
		false,
	},
	{
		"locale-free ASCII text, reversed",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"home", "go", "bears", "all", "above", "And"},
		language.Und,
		false,
		true,
	},
	{
		"locale-free ASCII text, case-insensitive, reversed",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"home", "go", "bears", "And", "all", "above"},
		language.Und,
		true,
		true,
	},
	{
		"en-US text",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"above", "all", "And", "bears", "go", "home"},
		language.English,
		false,
		false,
	},
	{
		"en-US text, reversed",
		[]string{"go", "bears", "above", "And", "all", "home"},
		[]string{"home", "go", "bears", "And", "all", "above"},
		language.English,
		false,
		true,
	},
	{
		"German text",
		[]string{"zoo", "foo", "öoo"},
		[]string{"foo", "öoo", "zoo"},
		language.German,
		false,
		false,
	},
	{
		"German text, reversed",
		[]string{"zoo", "foo", "öoo"},
		[]string{"zoo", "öoo", "foo"},
		language.German,
		false,
		true,
	},
	{
		"Swedish text",
		[]string{"zoo", "foo", "öoo"},
		[]string{"foo", "zoo", "öoo"},
		language.Swedish,
		false,
		false,
	},
	{
		"Swedish text, reversed",
		[]string{"zoo", "foo", "öoo"},
		[]string{"öoo", "zoo", "foo"},
		language.Swedish,
		false,
		true,
	},
}

func Test_textSort(t *testing.T) {
	for _, test := range textSortTests {
		t.Run(test.name, func(t *testing.T) {
			d := detest.New(t)
			// If the test fails and we haven't cloned then we cannot print
			// out debugging info with the original and the (improperly
			// sorted) list.
			clone := cloneSlice(test.input)
			textSort(clone, test.locale, test.caseInsensitive, test.reverse)
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
		})
	}
}

var numberedTextSortTests = []testCase{
	{
		"numbered locale-free ASCII text",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"0. bears", "1. all", "2. home", "5. And", "15 - above", "120001 go"},
		language.Und,
		false,
		false,
	},
	{
		"numbered locale-free ASCII text, case-insensitive",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"0. bears", "1. all", "2. home", "5. And", "15 - above", "120001 go"},
		language.Und,
		true,
		false,
	},
	{
		"numbered locale-free ASCII text, reversed",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"120001 go", "15 - above", "5. And", "2. home", "1. all", "0. bears"},
		language.Und,
		false,
		true,
	},
	{
		"numbered locale-free ASCII text, case-insensitive, reversed",
		[]string{"120001 go", "0. bears", "15 - above", "5. And", "1. all", "2. home"},
		[]string{"120001 go", "15 - above", "5. And", "2. home", "1. all", "0. bears"},
		language.Und,
		true,
		true,
	},
	{
		"German text",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"1. foo", "2. öoo", "2. zoo", "3. zoo"},
		language.German,
		false,
		false,
	},
	{
		"German text, reversed",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"3. zoo", "2. zoo", "2. öoo", "1. foo"},
		language.German,
		false,
		true,
	},
	{
		"Swedish text",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"1. foo", "2. zoo", "2. öoo", "3. zoo"},
		language.Swedish,
		false,
		false,
	},
	{
		"Swedish text, reversed",
		[]string{"3. zoo", "1. foo", "2. öoo", "2. zoo"},
		[]string{"3. zoo", "2. öoo", "2. zoo", "1. foo"},
		language.Swedish,
		false,
		true,
	},
	{
		"mixed numbered and unnumbered",
		[]string{"10. x", "aloe", "27. bar", "love", "1. hello"},
		[]string{"1. hello", "10. x", "27. bar", "aloe", "love"},
		language.Und,
		false,
		false,
	},
	{
		"mixed numbered and unnumbered, reversed",
		[]string{"10. x", "aloe", "27. bar", "love", "1. hello"},
		[]string{"love", "aloe", "27. bar", "10. x", "1. hello"},
		language.Und,
		false,
		true,
	},
}

func Test_numberedTextSort(t *testing.T) {
	for _, test := range append(textSortTests, numberedTextSortTests...) {
		t.Run(test.name, func(t *testing.T) {
			d := detest.New(t)
			// If the test fails and we haven't cloned then we cannot print
			// out debugging info with the original and the (improperly
			// sorted) list.
			clone := cloneSlice(test.input)
			numberedTextSort(clone, test.locale, test.caseInsensitive, test.reverse)
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
		})
	}
}

func cloneSlice(orig []string) []string {
	new := make([]string, len(orig))
	for i, o := range orig {
		new[i] = o
	}
	return new
}
