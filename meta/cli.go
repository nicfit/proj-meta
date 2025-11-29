package meta

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (p *project) UpdateCliCommand(cmd *cli.Command) (versionCmd *cli.Command) {
	v := p.NewVersionCliCommand()
	cmd.Version = p.Version()
	cmd.Commands = append(cmd.Commands, v)
	return v
}

func (p *project) NewVersionCliCommand() *cli.Command {
	var opts = newFlags()

	cmd := &cli.Command{
		Name:  "version",
		Usage: "Print program version",
		Action: func(context.Context, *cli.Command) error {
			handleVersionFlags(*opts, p)
			return nil
		},
	}

	for name, opt := range *opts {
		f := cli.BoolFlag{
			Name:        name,
			Aliases:     opt.aliases,
			Usage:       opt.usage,
			Destination: &(opt.destination),
		}
		cmd.Flags = append(cmd.Flags, &f)
	}

	return cmd
}
