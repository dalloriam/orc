package main

import (
	"context"
	"flag"

	"github.com/dalloriam/orc/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "orc"
	p.Description = "Personal service orchestrator & runtime."

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	p.Commands = []cli.Command{
		&serverCommand{},
	}

	p.FlagSet = flag.NewFlagSet("orc", flag.ExitOnError)

	p.Before = func(ctx context.Context) error {
		// Set the log level.
		logrus.SetLevel(logrus.DebugLevel)

		return nil
	}

	p.Run()
}
