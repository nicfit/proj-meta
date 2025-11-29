package meta

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/urfave/cli/v3"
	"golang.org/x/mod/semver"
)

type Project interface {
	Name() string
	Version() string
	Full() string

	UpdateCobraCommand(cmd *cobra.Command) *cobra.Command
	NewVersionCobraCommand() *cobra.Command

	UpdateCliCommand(cmd *cli.Command) *cli.Command
	NewVersionCliCommand() *cli.Command
}

func NewProjectMust(name string, version string) Project {
	if project, err := NewProject(name, version); err != nil {
		panic(err)
	} else {
		return project
	}
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

func (p *project) Full() string {
	return fmt.Sprintf("%s %s", p.name, p.version)
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

type flag struct {
	aliases     []string
	usage       string
	destination bool
}

type flagOpts map[string]*flag

func newFlags() *flagOpts {
	var flags = flagOpts{
		"short": {
			aliases: []string{"s"},
			usage:   "Output only the semantic version.",
		},
		"major": {
			usage: "Show the major version.",
		},
		"major-minor": {
			usage: "Show the major and minor version.",
		},
		"prerelease": {
			usage: "Show the prerelease version.",
		},
	}
	return &flags
}

func handleVersionFlags(opts flagOpts, project Project) {
	if !opts["short"].destination &&
		!opts["major"].destination &&
		!opts["major-minor"].destination &&
		!opts["prerelease"].destination {

		fmt.Println(project.Full())
	} else {
		if opts["short"].destination {
			fmt.Printf("%s\n", project.Version())
		}
		if opts["major"].destination {
			fmt.Println(semver.Major(project.Version())[1:])
		}
		if opts["major-minor"].destination {
			fmt.Println(semver.MajorMinor(project.Version())[1:])
		}

		if opts["prerelease"].destination {
			if p := semver.Prerelease(project.Version()); p != "" {
				fmt.Println(p[1:])
			}
		}
	}
}
