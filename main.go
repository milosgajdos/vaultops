package main

import (
	"fmt"
	"os"

	"github.com/milosgajdos/vaultops/command"
	"github.com/mitchellh/cli"
)

const (
	// cli version
	version = "0.1"
	// cli name
	cliName = "vaultops"
)

func main() {
	c := cli.NewCLI(cliName, version)
	c.Args = os.Args[1:]
	c.Commands = Commands()

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(exitStatus)
}

// Commands initializes all cli subcommands
func Commands() map[string]cli.CommandFactory {
	meta := &command.Meta{
		UI: &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			Ui: &cli.PrefixedUi{
				InfoPrefix:  "[INFO] ",
				WarnPrefix:  "[WARN] ",
				ErrorPrefix: "[ERROR] ",
				Ui: &cli.BasicUi{
					Writer:      os.Stdout,
					ErrorWriter: os.Stderr,
				},
			},
		},
	}

	return map[string]cli.CommandFactory{
		"init": func() (cli.Command, error) {
			return &command.InitCommand{
				Meta: *meta,
			}, nil
		},
		"unseal": func() (cli.Command, error) {
			return &command.UnsealCommand{
				Meta: *meta,
			}, nil
		},
	}
}
