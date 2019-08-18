// Copyright 2017 The Goma Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package posixpath

import (
	"reflect"
	"testing"
)

func TestIsAbs(t *testing.T) {
	for _, tc := range []struct {
		path  string
		isAbs bool
	}{
		{"", false},
		{"/", true},
		{"/usr/bin/gcc", true},
		{"..", false},
		{"/a/../bb", true},
		{".", false},
		{"./", false},
		{"lala", false},
	} {
		got := IsAbs(tc.path)
		if got != tc.isAbs {
			t.Errorf("IsAbs(%q)=%t; want=%t", tc.path, got, tc.isAbs)
		}
	}
}

func TestBase(t *testing.T) {
	for _, tc := range []struct {
		path, result string
	}{
		{"", "."},
		{".", "."},
		{"/.", "."},
		{"/", "/"},
		{"////", "/"},
		{"x/", "."}, // different from filepath.Base
		{"abc", "abc"},
		{"abc/def", "def"},
		{"a/b/.x", ".x"},
		{"a/b/c.", "c."},
		{"a/b/c.x", "c.x"},
	} {
		got := Base(tc.path)
		if got != tc.result {
			t.Errorf("Base(%q)=%q; want=%q", tc.path, got, tc.result)
		}
	}
}

func TestDir(t *testing.T) {
	for _, tc := range []struct {
		path, result string
	}{
		{"", "."},
		{".", "."},
		{"/.", "/"},
		{"/", "/"},
		{"////", "/"},
		{"/foo", "/"},
		{"x/", "x"},
		{"abc", "."},
		{"abc/def", "abc"},
		{"a/b/.x", "a/b"},
		{"a/b/c.", "a/b"},
		{"a/b/c.x", "a/b"},
	} {
		got := Dir(tc.path)
		if got != tc.result {
			t.Errorf("Dir(%q)=%q; want=%q", tc.path, got, tc.result)
		}
	}
}

func TestJoin(t *testing.T) {
	for _, tc := range []struct {
		elem []string
		path string
	}{
		// zero parameters.
		{nil, ""},

		// one parameter
		{[]string{""}, ""},
		{[]string{"/"}, "/"},
		{[]string{"a"}, "a"},

		// two parameters
		{[]string{"a", "b"}, "a/b"},
		{[]string{"a", ""}, "a"},
		{[]string{"", "b"}, "b"},
		{[]string{"/", "a"}, "/a"},
		{[]string{"/", "a/b"}, "/a/b"},
		{[]string{"/", ""}, "/"},
		{[]string{"//", "a"}, "/a"},
		{[]string{"/a", "b"}, "/a/b"},
		{[]string{"a/", "b"}, "a/b"},
		{[]string{"a/", ""}, "a"},
		{[]string{"", ""}, ""},
		// different from filepath.Join, it won't clean /../.
		{[]string{"a", "../b"}, "a/../b"},

		// tree parameters
		{[]string{"/", "a", "b"}, "/a/b"},
	} {
		got := Join(tc.elem...)
		if got != tc.path {
			t.Errorf("Join(%q)=%q; want=%q", tc.elem, got, tc.path)
		}
	}
}

func TestRel(t *testing.T) {
	for _, tc := range []struct {
		root, path, want string
	}{
		{"a/b", "a/b", "."},
		{"a/b/.", "a/b", "."},
		{"a/b", "a/b/.", "."},
		{"./a/b", "a/b", "."},
		{"a/b", "./a/b", "."},
		{"ab/cd", "ab/cde", "../cde"},
		{"ab/cd", "ab/c", "../c"},
		{"a/b", "a/b/c/d", "c/d"},
		{"a/b", "a/b/../c", "../c"},
		{"a/b/../c", "a/b", "../b"},
		{"a/b/c", "a/c/d", "../../c/d"},
		{"a/b", "c/d", "../../c/d"},
		{"a/b/c/d", "a/b", "../.."},
		{"a/b/c/d", "a/b/", "../.."},
		{"a/b/c/d/", "a/b", "../.."},
		{"a/b/c/d/", "a/b/", "../.."},
		{"../../a/b", "../../a/b/c/d", "c/d"},
		{"/a/b", "/a/b", "."},
		{"/a/b/.", "/a/b", "."},
		{"/a/b", "/a/b/.", "."},
		{"/ab/cd", "/ab/cde", "../cde"},
		{"/ab/cd", "/ab/c", "../c"},
		{"/a/b", "/a/b/c/d", "c/d"},
		{"/a/b", "/a/b/../c", "../c"},
		{"/a/b/../c", "/a/b", "../b"},
		{"/a/b/c", "/a/c/d", "../../c/d"},
		{"/a/b", "/c/d", "../../c/d"},
		{"/a/b/c/d", "/a/b", "../.."},
		{"/a/b/c/d", "/a/b/", "../.."},
		{"/a/b/c/d/", "/a/b", "../.."},
		{"/a/b/c/d/", "/a/b/", "../.."},
		{"/../../a/b", "/../../a/b/c/d", "c/d"},
		{".", "a/b", "a/b"},
		{".", "..", ".."},
	} {
		got, err := Rel(tc.root, tc.path)
		if err != nil || got != tc.want {
			t.Errorf("Rel(%q, %q)=%q, %v; want %q, nil", tc.root, tc.path, got, err, tc.want)
		}
	}

	// can't do purely lexically
	for _, tc := range []struct {
		root, path string
	}{
		{"..", "."},
		{"..", "a"},
		{"../..", ".."},
		{"a", "/a"},
		{"/a", "a"},
	} {
		got, err := Rel(tc.root, tc.path)
		if err == nil {
			t.Errorf("Rel(%q, %q)=%q, nil; want error", tc.root, tc.path, got)
		}
	}
}

func TestClean(t *testing.T) {
	for _, tc := range []struct {
		path, result string
	}{
		// https://golang.org/src/path/filepath/path_test.go

		// Already clean
		{"abc", "abc"},
		{"abc/def", "abc/def"},
		{"a/b/c", "a/b/c"},
		{".", "."},
		{"..", ".."},
		{"../..", "../.."},
		{"../../abc", "../../abc"},
		{"/abc", "/abc"},
		{"/", "/"},

		// Empty is current dir
		{"", "."},

		// Remove trailing slash
		{"abc/", "abc"},
		{"abc/def/", "abc/def"},
		{"a/b/c/", "a/b/c"},
		{"./", "."},
		{"../", ".."},
		{"../../", "../.."},
		{"/abc/", "/abc"},

		// Remove doubled slash
		{"abc//def//ghi", "abc/def/ghi"},
		{"//abc", "/abc"},
		{"///abc", "/abc"},
		{"//abc//", "/abc"},
		{"abc//", "abc"},

		// Remove . elements
		{"abc/./def", "abc/def"},
		{"/./abc/def", "/abc/def"},
		{"abc/.", "abc"},

		// Remove .. elements
		{"abc/def/ghi/../jkl", "abc/def/jkl"},
		{"abc/def/../ghi/../jkl", "abc/jkl"},
		{"abc/def/..", "abc"},
		{"abc/def/../..", "."},
		{"/abc/def/../..", "/"},
		{"abc/def/../../..", ".."},
		{"/abc/def/../../..", "/"},
		{"abc/def/../../../ghi/jkl/../../../mno", "../../mno"},
		{"/../abc", "/abc"},

		// Combinations
		{"abc/./../def", "def"},
		{"abc//./../def", "def"},
		{"abc/../../././../def", "../../def"},
	} {
		got := Clean(tc.path)
		if got != tc.result {
			t.Errorf("Clean(%q)=%q; want=%q", tc.path, got, tc.result)
		}
	}
}

func TestSplitElem(t *testing.T) {
	for _, tc := range []struct {
		path string
		elem []string
	}{
		{"", nil},

		{"/", []string{"/"}},
		{"a", []string{"a"}},
		{".", []string{"."}},

		{"/.", []string{"/", "."}},
		{"////", []string{"/"}},
		{"////.", []string{"/", "."}},
		{".////", []string{"."}},

		{"a/b", []string{"a", "b"}},
		{"/a/b", []string{"/", "a", "b"}},
		{"a/b/.", []string{"a", "b", "."}},
		{"a/b/", []string{"a", "b", "."}},

		{"a//b", []string{"a", "b"}},
		{"//a", []string{"/", "a"}},
		{"a//", []string{"a", "."}},
		{"a//.", []string{"a", "."}},
		{"a/./b", []string{"a", "b"}},
		{"a/././b", []string{"a", "b"}},

		{"a/../b", []string{"a", "..", "b"}},
	} {
		got := SplitElem(tc.path)
		if !reflect.DeepEqual(got, tc.elem) {
			t.Errorf("SplitElem(%q)=%q; want=%q", tc.path, got, tc.elem)
		}
	}
}
