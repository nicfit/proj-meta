package meta

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewProjectInvalidName(t *testing.T) {
	var err error

	for _, name := range []string{"", "   ", "\n\t", "\t   "} {
		t.Run("name=<empty>", func(t *testing.T) {
			_, err = NewProject(name, "v1.0.0")
			assert.Error(t, err)
			assert.ErrorContains(t, err, "empty project name")
		})
	}

	for _, name := range []string{"G Unit", "The\tBeatles", "Marquee\nMoon", "Frank\rZappa"} {
		t.Run("name="+name, func(t *testing.T) {
			_, err = NewProject(name, "v1.0.0")
			assert.Error(t, err)
			assert.ErrorContains(t, err, "multi-word project name")
		})
	}
}

func TestNewProjectInvalidVersion(t *testing.T) {
	var err error

	for _, version := range []string{"", "   ", "\n\t", "\t   "} {
		t.Run("version=<empty>", func(t *testing.T) {
			_, err = NewProject("EHG", version)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "empty project version")
		})
	}

	for _, version := range []string{"LTS", "1", "1.0", "1.0.0", "version1.0.0", "va.b.c", "v1.0.0-beta!!"} {
		t.Run("version="+version, func(t *testing.T) {
			_, err = NewProject("EHG", version)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "invalid semantic version")
		})
	}

}

func TestNewProject(t *testing.T) {
	type testCase struct {
		testName        string
		name            string
		version         string
		expectedName    string
		expectedVersion string
	}

	tests := []testCase{
		{"Valid", "Television", "v5.4.3", "Television", "v5.4.3"},
		{"Valid with prerelease", "Radiohead", "v2.0.0-beta.1", "Radiohead", "v2.0.0-beta.1"},
		{"ValidWithNameTrim", " Television\t", "v5.4.3", "Television", "v5.4.3"},
		{"ValidWithVersion", " Television", "   v5.4.3\n", "Television", "v5.4.3"},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			p, err := NewProject(tc.name, tc.version)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedName, p.Name())
			assert.Equal(t, tc.expectedVersion, p.Version())
		})
	}
}

func TestProjectNewCobraCommand(t *testing.T) {
	p, err := NewProject("marquee-moon", "v19.7.7-dev")
	assert.NoError(t, err)

	cmd := p.NewVersionCobraCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Print program version", cmd.Short)

	type flagInfo struct {
		name      string
		shorthand string
	}
	for _, f := range []flagInfo{
		{"short", "s"},
		{"prerelease", ""},
		{"major", ""},
		{"major-minor", ""},
	} {
		t.Run(f.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(f.name)
			assert.NotNil(t, flag)
			assert.Equal(t, f.shorthand, flag.Shorthand)
		})
	}

	og_stdout := os.Stdout
	defer func() { os.Stdout = og_stdout }()

	type testCase struct {
		args           []string
		expectedOutput string
	}
	for _, tc := range []testCase{
		{[]string{"version"}, "marquee-moon v19.7.7-dev\n"},
		{[]string{"version", "--short"}, "v19.7.7-dev\n"},
		{[]string{"version", "--prerelease"}, "dev\n"},
		{[]string{"version", "--major"}, "19\n"},
		{[]string{"version", "--major-minor"}, "19.7\n"},
	} {
		r, w, err := os.Pipe()
		assert.NoError(t, err)
		os.Stdout = w

		main_cmd := &cobra.Command{
			Use: "integrity",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		main_cmd.AddCommand(p.NewVersionCobraCommand())

		t.Run("Execute:"+strings.Join(tc.args, ","), func(t *testing.T) {
			os.Args = []string{"integrity"}
			os.Args = append(os.Args, tc.args...)
			err := main_cmd.Execute()
			assert.NoError(t, err)
			assert.NoError(t, w.Close())

			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			assert.NoError(t, r.Close())
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}

func TestProjectNewCliCommand(t *testing.T) {
	p, err := NewProject("Maths+English", "v20.0.7-rc3")
	assert.NoError(t, err)

	cmd := p.NewVersionCliCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Name)
	assert.Equal(t, "Print program version", cmd.Usage)

	type flagInfo struct {
		name      string
		shorthand string
	}
	for _, f := range []flagInfo{
		{"short", "s"},
		{"prerelease", ""},
		{"major", ""},
		{"major-minor", ""},
	} {
		t.Run(f.name, func(t *testing.T) {
			found := false
			for i := range cmd.Flags {
				if slices.Contains(cmd.Flags[i].Names(), f.name) {
					found = true
					break
				}
			}
			assert.True(t, found)
		})
	}

	og_stdout := os.Stdout
	defer func() { os.Stdout = og_stdout }()

	type testCase struct {
		args           []string
		expectedOutput string
	}
	for _, tc := range []testCase{
		{[]string{"version"}, "Maths+English v20.0.7-rc3\n"},
		{[]string{"version", "--short"}, "v20.0.7-rc3\n"},
		{[]string{"version", "--prerelease"}, "rc3\n"},
		{[]string{"version", "--major"}, "20\n"},
		{[]string{"version", "--major-minor"}, "20.0\n"},
	} {
		r, w, err := os.Pipe()
		assert.NoError(t, err)
		os.Stdout = w

		main_cmd := &cobra.Command{
			Use: "face-value",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		main_cmd.AddCommand(p.NewVersionCobraCommand())

		t.Run("Execute:"+strings.Join(tc.args, ","), func(t *testing.T) {
			os.Args = []string{"face-value"}
			os.Args = append(os.Args, tc.args...)
			err := main_cmd.Execute()
			assert.NoError(t, err)
			assert.NoError(t, w.Close())

			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			assert.NoError(t, r.Close())
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}
