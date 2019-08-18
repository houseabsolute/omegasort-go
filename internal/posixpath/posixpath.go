// Copied from the goma/server project. If I try to use this as an import then
// the go module system thinks I need _all_ the deps for the entire project,
// as opposed to just what's needed for winpath and posixpath.

// Copyright 2017 The Goma Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package posixpath handles posix-path (Unix style; slash separeted path).
package posixpath

import (
	"fmt"
	"strings"
)

// FilePath provides posix filepath.
type FilePath struct{}

func (FilePath) IsAbs(path string) bool     { return IsAbs(path) }
func (FilePath) Base(path string) string    { return Base(path) }
func (FilePath) Dir(path string) string     { return Dir(path) }
func (FilePath) Join(elem ...string) string { return Join(elem...) }

func (FilePath) Rel(basepath, targpath string) (string, error) {
	return Rel(basepath, targpath)
}

func (FilePath) Clean(path string) string       { return Clean(path) }
func (FilePath) SplitElem(path string) []string { return SplitElem(path) }
func (FilePath) PathSep() string                { return "/" }

// IsAbs returns true if fname is absolute path.
func IsAbs(fname string) bool {
	return strings.HasPrefix(fname, "/")
}

// Base returns the last element of fname.
// If fname is empty, or ends with path separator, Base returns ".".
// If fname is "/" only, Base returns "/".
func Base(fname string) string {
	elem := SplitElem(fname)
	if len(elem) == 0 {
		return "."
	}
	base := elem[len(elem)-1]
	return base
}

// Dir returns all but the last element of path, typically the path's
// directory.  Differnt from filepath.Dir, it won't clean ".." in the result.
// If path is empty, Dir returns ".".
func Dir(fname string) string {
	elem := SplitElem(fname)
	if len(elem) == 0 {
		return "."
	}
	if len(elem) == 1 {
		if elem[0] == "/" {
			return "/"
		}
		return "."
	}
	elem = elem[:len(elem)-1]
	return joinElem(elem)
}

// Join joins any number of path elements into a single path, adding a "/"
// if necessary.  Different from filepath.Join, it won't clean ".." in
// the result. All empty strings are ignored.
func Join(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}
	var paths []string
	for _, e := range elem {
		if e == "" {
			continue
		}
		paths = append(paths, e)
	}
	elem = SplitElem(strings.Join(paths, "/"))
	if len(elem) > 0 && elem[len(elem)-1] == "." {
		elem = elem[:len(elem)-1]
	}
	if len(elem) == 0 {
		return ""
	}
	return joinElem(elem)
}

// Rel returns a relative path that is lexically equivalent to targpath when
// joined to basepath with an intervening separator.
func Rel(basepath, targpath string) (string, error) {
	if IsAbs(basepath) != IsAbs(targpath) {
		return "", fmt.Errorf("Rel: can't make %s relative to %s", targpath, basepath)
	}

	base := SplitElem(basepath)
	base = cleanElem(base)
	targ := SplitElem(targpath)
	targ = cleanElem(targ)

	// Find first different elem at i. (base[:i] equals targ[:i])
	bl := len(base)
	tl := len(targ)
	var i int
	for i < bl && i < tl {
		if base[i] != targ[i] {
			break
		}
		i++
	}
	var elem []string
	if i != bl {
		// Base elements left.
		alldotdot := func(elem []string) bool {
			for _, e := range elem {
				if e != ".." {
					return false
				}
			}
			return true
		}
		if alldotdot(base[i:]) {
			return "", fmt.Errorf("Rel: can't make %s relative to %s", targpath, basepath)
		}
		n := bl - i
		for i := 0; i < n; i++ {
			elem = append(elem, "..")
		}
	}
	elem = append(elem, targ[i:]...)
	return joinElem(elem), nil
}

// Clean returns the shortest path name equivalent to path by purely lexical processing.
func Clean(path string) string {
	elems := SplitElem(path)
	elems = cleanElem(elems)
	return joinElem(elems)
}

// SplitElem splits path into element, separated by "/".
// If fname is absolute path, first element is "/".
// If fname ends with "/" or "/.", last element is ".".
// Empty string, "/" or "." won't be appeared in other elements.
func SplitElem(fname string) []string {
	if fname == "" {
		return nil
	}
	if strings.Repeat("/", len(fname)) == fname {
		return []string{"/"}
	}
	if fname == "." {
		return []string{"."}
	}
	var elem []string
	if strings.HasPrefix(fname, "/") {
		elem = append(elem, "/")
	}
	for _, e := range strings.Split(fname, "/") {
		if e == "" || e == "." {
			continue
		}
		elem = append(elem, e)
	}
	if strings.HasSuffix(fname, "/") || strings.HasSuffix(fname, "/.") {
		elem = append(elem, ".")
	}
	return elem
}

func joinElem(elem []string) string {
	if len(elem) == 0 {
		return "."
	}
	if len(elem) == 1 {
		return elem[0]
	}
	if elem[0] == "/" {
		elem[0] = ""
	}
	return strings.Join(elem, "/")
}

func cleanElem(elem []string) []string {
	if len(elem) > 0 && elem[len(elem)-1] == "." {
		elem = elem[:len(elem)-1]
	}
	var r []string
	for _, e := range elem {
		if e == ".." {
			if len(r) == 1 && r[0] == "/" {
				continue
			}
			if len(r) > 0 && r[len(r)-1] != ".." {
				r = r[:len(r)-1]
				continue
			}
		}
		r = append(r, e)
	}
	return r
}
