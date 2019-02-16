package main

import (
	"context"
	"flag"

	"github.com/dalloriam/orc/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
)

type inputData struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "local"
	p.Description = "Local Host Interface Plugin for ORC"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	p.Commands = []cli.Command{
		&notifyCommand{},
	}

	p.FlagSet = flag.NewFlagSet("orc", flag.ExitOnError)

	p.Before = func(ctx context.Context) error {
		// Set the log level.
		logrus.SetLevel(logrus.DebugLevel)

		return nil
	}

	p.Run()
}
