// +build !windows

package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/houseabsolute/detest/pkg/detest"
)

const binary = "../omegasort"

func TestMain(m *testing.M) {
	c := exec.Command("go", "build")
	// The integration tests working directory will be the directory
	// containing this test file.
	c.Dir = ".."
	err := c.Run()
	if err != nil {
		panic(fmt.Errorf("error running go build: %w", err))
	}

	os.Exit(m.Run())
}

func TestAllCases(t *testing.T) {
	td := t.TempDir()

	err := filepath.Walk("cases", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".test") {
			t.Run(path, func(t *testing.T) { runOneTest(t, td, path) })
		}
		return nil
	})
	d := detest.New(t)
	d.Require(d.Is(err, nil, "no error walking directory for cases"))
}

func TestCheckUnique(t *testing.T) {
	config := config{
		Sort:   "text",
		Unique: true,
		Check:  true,
	}
	td := t.TempDir()

	type test struct {
		name       string
		content    string
		expectFail bool
	}
	tests := []test{
		{
			name: "unique and sorted",
			content: `
a
b
c
d
e
f
`,
			expectFail: false,
		},
		{
			name: "unique and unsorted",
			content: `
c
d
a
b
e
f
`,
			expectFail: true,
		},
		{
			name: "not unique and sorted",
			content: `
a
b
c
c
d
e
f
f
`,
			expectFail: true,
		},
		{
			name: "not unique and sorted",
			content: `
c
c
d
e
a
b
c
c
d
e
f
f
`,
			expectFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := detest.New(t)

			tf := filepath.Join(td, strings.ReplaceAll(test.name, " ", "-"))
			err := ioutil.WriteFile(tf, []byte(test.content), 0755)
			d.Require(d.Is(err, nil, "no error writing to %s", tf))

			out, err := runOmegasort(d, config, tf)
			if test.expectFail {
				d.IsNot(err, nil, "got an error when running omegasort")
				var exitErr *exec.ExitError
				if d.Is(errors.As(err, &exitErr), true, "error is an *exec.ExitError") {
					d.Is(exitErr.ExitCode(), 1, "exit code is 1")
				}
			} else {
				d.Is("", out, "no output when running omegasort")
				d.Is(err, nil, "no error running omegasort")
			}
		})
	}
}

type config struct {
	Sort            string `json:"sort"`
	Locale          string `json:"locale"`
	Unique          bool   `json:"unique"`
	CaseInsensitive bool   `json:"case_insensitive"`
	Reverse         bool   `json:"reverse"`
	Windows         bool   `json:"windows"`
	Check           bool
}

func runOneTest(t *testing.T, td, path string) {
	d := detest.New(t)

	s := readFile(d, path)

	test := strings.Split(s, "----")
	fc, err := d.Func(func(test []string) (bool, string) {
		ok := len(test) >= 2 && len(test) <= 3
		expl := ""
		if !ok {
			expl = "expected 2 or 3 items in the test file"
		}
		return ok, expl
	})
	d.Require(d.Is(err, nil, "error from d.Func"))
	if err != nil {
		t.Fatal(err)
	}
	if !d.Passes(test, fc, "matched optional config, test input, and expected output in test file content") {
		return
	}

	c := config{
		Sort: "text",
	}
	var input, expect string

	if len(test) == 3 {
		err = json.Unmarshal([]byte(test[0]), &c)
		d.Is(err, nil, "no error unmarshaling test config")
		input = test[1]
		expect = test[2]
	} else {
		input = test[0]
		expect = test[1]
	}
	expect = strings.TrimSpace(expect)

	tf := filepath.Join(td, filepath.Base(path))
	err = ioutil.WriteFile(tf, []byte(input), 0755)
	d.Require(d.Is(err, nil, fmt.Sprintf("no error writing to %s", tf)))

	out, err := runOmegasort(d, c, tf)
	d.Is("", out, "no output when running omegasort on %s", tf)
	d.Require(d.Is(err, nil, "no error running omegasort on %s"), tf)

	sorted := strings.TrimSpace(readFile(d, tf))
	d.Is(sorted, expect, "got the expected sorted output")
}

func readFile(d *detest.D, path string) string {
	f, err := os.Open(path)
	d.Require(d.Is(err, nil, "no error opening %s", path))
	b, err := ioutil.ReadAll(f)
	d.Require(d.Is(err, nil, "no error reading %s", path))

	return string(b)
}

func (c *config) args() []string {
	args := []string{"--sort", c.Sort}
	if c.Locale != "" {
		args = append(args, "--locale", c.Locale)
	}
	if c.Unique {
		args = append(args, "--unique")
	}
	if c.CaseInsensitive {
		args = append(args, "--case-insensitive")
	}
	if c.Reverse {
		args = append(args, "--reverse")
	}
	if c.Windows {
		args = append(args, "--windows")
	}
	if c.Check {
		args = append(args, "--check")
	} else {
		args = append(args, "--in-place")
	}

	return args
}

func runOmegasort(d *detest.D, c config, file string) (string, error) {
	args := c.args()
	args = append(args, file)
	cmd := exec.Command(binary, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
