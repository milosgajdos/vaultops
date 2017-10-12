package main

import (
	"fmt"
	"os"

	"github.com/milosgajdos83/vaultops/command"
	"github.com/mitchellh/cli"
)

const (
	// cli version
	version = "0.1"
	// cli name
	cliName = "vaultops"
)

func main() {
	// create new cli with given version and name
	c := cli.NewCLI(cliName, version)
	c.Args = os.Args[1:]
	c.Commands = Commands()
	// run a command
	exitStatus, err := c.Run()
	if err != nil {
		fmt.Println(err)
	}
	// exit with poroper exit status
	os.Exit(exitStatus)
}

// Commands initializes all clie csubdommands
func Commands() map[string]cli.CommandFactory {
	// meta command params are inherited by almost every command
	meta := &command.Meta{
		// we will use color UI
		UI: &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			Ui: &cli.PrefixedUi{
				InfoPrefix:  "[ info ] ",
				ErrorPrefix: "[ errpr ] ",
				WarnPrefix:  "[ warn ] ",
				Ui: &cli.BasicUi{
					Writer:      os.Stdout,
					ErrorWriter: os.Stderr,
				},
			},
		},
	}
	// return commands factory
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
		"mount": func() (cli.Command, error) {
			return &command.MountCommand{
				Meta: *meta,
			}, nil
		},
		"backend": func() (cli.Command, error) {
			return &command.BackendCommand{
				Meta: *meta,
			}, nil
		},
		"policy": func() (cli.Command, error) {
			return &command.PolicyCommand{
				Meta: *meta,
			}, nil
		},
		"setup": func() (cli.Command, error) {
			return &command.SetupCommand{
				Meta: *meta,
			}, nil
		},
	}
}
