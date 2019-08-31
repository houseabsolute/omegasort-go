# What is this?

Omegasort if a text file sorting tool that aims to be the last sorting tool you'll need.

I wrote this because I like to keep various types of files in a sorted order (gitignore files, lists of spelling stopwords, etc.). I wanted a tool I could call as part of my commit hooks and CI (using [precious](https://github.com/houseabsolute/precious).

## usage: `omegasort [<flags>] [<file>]`

### Flags:

| Short | Long | Description |
| ----- | ---- | ----------- |
| `-h`  | `--help` | Show context-sensitive help (also try `--help-long` and `--help-man`). |
| | `--version` | Show application version. |
| `-s` | `--sort=SORT` | The type of sorting to use. See below for options. |
| `-l` | `--locale=""` | The locale to use for sorting. If this is not specified the sorting is in codepoint order. |
| `-c` | `--case-insensitive` | Sort case-insensitively. Note that many locales always do this so if you specify a locale you may get case-insensitive output regardless of this flag. |
| `-r` | `--reverse` | Sort in reverse order. |
| | `--windows` | Parse paths as Windows paths for path sort. |
| `-i` | `--in-place` | Modify the file in place instead of making a backup. |
| | `--stdout` | Print the sorted output to stdout instead of making a new file. |
| | `--check` | Check that the file is sorted instead of sorting it. If it is not sorted the exit status will be 2. |
| | `--debug` | Print out debugging info while running. |
| | `--docs` | Print out extended sorting documentation. |

### Args:

* `[<file>]`  The file to sort.

## Sorting Options:

* text - sort the file as text according to the specified locale
* numbered-text - sort the file assuming that each line starts with a numeric prefix, then fall back to sorting by text according to the specified locale
* datetime-text - sort the file assuming that each line starts with a date or datetime prefix, then fall back to sorting by text according to the specified locale
* path - sort the file assuming that each line is a path, sorted so that deeper paths come after shorter
* ip - sort the file assuming that each line is an IP address
* network - sort the file assuming that each line is a network in CIDR form

### Text

This sorts each line of the file as text without any special parsing. The exact
sorting is determined by the `--locale,` `--case-insensitive,` and `--reverse` flags.
See below for details on how locales work.

### Numbered Text

This assumes that each line of the file starts with a numeric value, optionally
followed by non-numeric text.

Lines should not have any leading space before the number. The number can
either be an integer (including 0) or a simple float (no scientific notation).

The lines will be sorted numerically first. If two lines have the same number
they will be sorted by text as above.

Lines without numbers always sort after lines with numbers.

This sorting method accepts the `--locale,` `--case-insensitive,` and `--reverse`
flags.

### Path Sort

Each line is treated as a path.

The paths are sorted by the following rules:

* Absolute paths come before relative.
* Paths are sorted by depth before sorting by the path content, so /z comes
before /a/a.
* If you pass the `--windows` flag, then paths with drive letters are sorted
based on the drive letter first. Paths with drive letters sort before paths
without them.

This sorting method accepts the `--locale,` `--case-insensitive,` and `--reverse`
flags in addition to the `--windows` flag.

### Datetime Sort

This sorting method assumes that each line starts with a date or datetime,
without any space in it. That means datetimes need to be in a format like
"2019-08-27T19:13:16".

Lines should not have any leading space before the datetime.

This sorting method accepts the `--locale,` `--case-insensitive,` and `--reverse`
flags.

### IP Sort

This method assumes that each line is an IPv4 or IPv6 address (not a network).

The sorting method is the same as if each line were the corresponding integer
for the address.

This sorting method accepts the `--reverse` flag.

### Network Sort

This method assumes that each line is an IPv4 or IPv6 network in CIDR notation.

If there are two networks with the same base address they are sorted with the
larger network first (so 1.1.1.0/24 comes before 1.1.1.0/28).

This sorting method accepts the `--reverse` flag.

## Build Status

[![Build Status](https://travis-ci.com/houseabsolute/omegasort.svg?branch=master)](https://travis-ci.com/houseabsolute/omegasort)
