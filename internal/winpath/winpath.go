// Copied from the goma/server project. If I try to use this as an import then
// the go module system thinks I need _all_ the deps for the entire project,
// as opposed to just what's needed for winpath and posixpath.

// Copyright 2017 The Goma Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package winpath handles windows-path (backslash separated path).
// It also accepts slash as path separator.
package winpath

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/houseabsolute/omegasort/internal/posixpath"
)

var (
	absPathPattern = regexp.MustCompile(`^[A-Za-z]:[/\\].*`)
	drivePattern   = regexp.MustCompile(`^([A-Za-z]:)(.*)`)
)

// FilePath provides win filepath.
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
func (FilePath) PathSep() string                { return `\` }

// IsAbs returns true if fname is absolute path.
func IsAbs(fname string) bool {
	return absPathPattern.MatchString(fname)
}

// Base returns the last element of fname.
// If fname is empty, or ends with path separator, Base returns ".".
// If fname is `\` only, Base returns `\`.
func Base(fname string) string {
	if fname == "" {
		return "."
	}
	fname = strings.Replace(fname, `\`, "/", -1)
	_, path := splitDrive(fname)
	if path == "" {
		return "."
	}
	base := posixpath.Base(path)
	return strings.Replace(base, "/", `\`, -1)
}

// Dir returns all but the last element of path, typically the path's
// directory.  Differnt from filepath.Dir, it won't clean ".." in the result.
// If path is empty, Dir returns ".".
func Dir(fname string) string {
	if fname == "" {
		return "."
	}
	fname = strings.Replace(fname, `\`, "/", -1)
	drive, path := splitDrive(fname)
	if path == "" {
		return drive
	}
	dirname := posixpath.Dir(path)
	return drive + strings.Replace(dirname, "/", `\`, -1)
}

// Join joins any number of path elements into a single path, adding a "/"
// if necessary.  Different from filepath.Join, it won't clean ".." in
// the result. All empty strings are ignored.
func Join(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}
	// copy
	elem = append([]string{}, elem...)
	for i := range elem {
		elem[i] = strings.Replace(elem[i], `\`, "/", -1)
	}
	var drive string
	drive, elem[0] = splitDrive(elem[0])
	path := posixpath.Join(elem...)
	return drive + strings.Replace(path, "/", `\`, -1)
}

// Rel returns a relative path that is lexically equivalent to targpath when
// joined to basepath with an intervening separator.
// TODO: case insensitive match.
func Rel(basepath, targpath string) (string, error) {
	if IsAbs(basepath) != IsAbs(targpath) {
		return "", fmt.Errorf("Rel: can't make %s relative to %s", targpath, basepath)
	}
	bdrive, bpath := splitDrive(basepath)
	tdrive, tpath := splitDrive(targpath)
	if bdrive != tdrive {
		return "", fmt.Errorf("Rel: can't make %s relative to %s", targpath, basepath)
	}
	bpath = strings.Replace(bpath, `\`, "/", -1)
	tpath = strings.Replace(tpath, `\`, "/", -1)
	rpath, err := posixpath.Rel(bpath, tpath)
	if err != nil {
		return "", err
	}
	return strings.Replace(rpath, "/", `\`, -1), nil
}

// Clean returns the shortest path name equivalent to path by purely lexical processing.
func Clean(path string) string {
	drive, path := splitDrive(path)
	path = strings.Replace(path, `\`, "/", -1)
	path = posixpath.Clean(path)

	return drive + strings.Replace(path, "/", `\`, -1)
}

// SplitElem splits path into element, separated by `\` or "/".
// If fname is absolute path, first element is `\` or `<drive>:\`,
// otherwise if fname has drive, first element is `<drive>:`.
// If fname ends with `\` or `\.`, last element is ".".
// Empty string, "/", `\` or "." won't be appeared in other elements.
func SplitElem(fname string) []string {
	drive, path := splitDrive(fname)
	path = strings.Replace(path, `\`, "/", -1)
	elems := posixpath.SplitElem(path)
	if len(elems) > 0 && elems[0] == "/" {
		elems[0] = drive + `\`
	} else if drive != "" {
		elems = append([]string{drive}, elems...)
	}
	return elems
}

func splitDrive(fname string) (string, string) {
	m := drivePattern.FindStringSubmatch(fname)
	if m != nil {
		return m[1], m[2]
	}
	return "", fname
}
