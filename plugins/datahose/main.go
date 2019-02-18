package main

import (
	"context"
	"flag"

	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "datahose"
	p.Description = "Datahose plugin for ORC"

	p.Commands = []cli.Command{
		&manifestCommand{},
	}

	p.FlagSet = flag.NewFlagSet("orc", flag.ExitOnError)

	p.Before = func(ctx context.Context) error {
		// Set the log level.
		logrus.SetLevel(logrus.DebugLevel)

		return nil
	}

	p.Run()
}
