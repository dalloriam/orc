package main

import (
	"context"
	"flag"

	"github.com/dalloriam/orc/plugins/datahose/hose"
)

const (
	pushCommandName = "push"
	pushCommandHelp = "Pushes an event to the datahose."
	pushCommandArgs = "KEY BODY"
)

type pushCommand struct{}

func (cmd *pushCommand) Name() string              { return pushCommandName }
func (cmd *pushCommand) Args() string              { return pushCommandArgs }
func (cmd *pushCommand) ShortHelp() string         { return pushCommandHelp }
func (cmd *pushCommand) LongHelp() string          { return pushCommandHelp }
func (cmd *pushCommand) Hidden() bool              { return true }
func (cmd *pushCommand) Register(fs *flag.FlagSet) {}

func (cmd *pushCommand) Run(ctx context.Context, args []byte) error {
	client, err := hose.NewClient()
	if err != nil {
		return err
	}
}
