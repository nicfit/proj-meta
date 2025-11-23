package meta

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type Project interface {
	Name() string
	Version() string

	UpdateCobraCommand(cmd *cobra.Command)
	NewVersionCobraCommand() *cobra.Command
}

func NewProject(name string, version string) (Project, error) {
	var err error

	if name, err = validateProjectName(name); err != nil {
		return nil, err
	}

	if version, err = validateProjectVersion(version); err != nil {
		return nil, err
	}

	return &project{
		version: version,
		name:    name,
	}, nil
}

func (p *project) Name() string {
	return p.name
}

func (p *project) Version() string {
	return p.version
}

func (p *project) NewVersionCobraCommand() *cobra.Command {
	var (
		show_short       bool
		show_major       bool
		show_major_minor bool
		show_pre         bool
	)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print program version",
		Run: func(cmd *cobra.Command, args []string) {
			if show_short {
				fmt.Printf("%s\n", p.version)
			} else if show_major {
				fmt.Println(semver.Major(p.version)[1:])
			} else if show_major_minor {
				fmt.Println(semver.MajorMinor(p.version)[1:])
			} else if show_pre {
				if p := semver.Prerelease(p.version); p != "" {
					fmt.Println(p[1:])
				}
			} else {
				fmt.Printf("%s %s\n", p.name, p.version)
			}
		},
	}

	cmd.Flags().BoolVarP(&show_short, "short", "s", false, "Output only the semantic version.")
	cmd.Flags().BoolVar(&show_major, "major", false, "Show the major version.")
	cmd.Flags().BoolVar(&show_major_minor, "major-minor", false, "Show the major and minor version.")
	cmd.Flags().BoolVar(&show_pre, "pre-release", false, "Show the prerelease version.")
	cmd.MarkFlagsMutuallyExclusive("short", "major", "major-minor", "pre-release")

	return cmd
}

func (p *project) UpdateCobraCommand(cmd *cobra.Command) {
	cmd.Version = p.version
	cmd.AddCommand(p.NewVersionCobraCommand())
}

type project struct {
	name    string
	version string
}

func validateProjectName(name string) (string, error) {
	name = strings.Trim(name, cutset)
	if name == "" {
		return name, errors.New("empty project name")
	} else if strings.ContainsAny(name, cutset) {
		return name, errors.New("multi-word project name")
	}

	return name, nil
}

func validateProjectVersion(version string) (string, error) {
	version = strings.Trim(version, cutset)
	if version == "" {
		return version, errors.New("empty project version")
	}

	if !semver.IsValid(version) {
		return version, errors.New("invalid semantic version")
	}
	version = semver.Canonical(version)

	return version, nil
}

const cutset = "\t\n\r "
