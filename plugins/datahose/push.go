package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

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

func (cmd *pushCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("payload information is required")
	}

	client, err := hose.NewClient()
	if err != nil {
		return err
	}

	var payload hose.Payload

	if err := json.Unmarshal([]byte(args[0]), &payload); err != nil {
		return err
	}

	if err := client.Push(payload); err != nil {
		return err
	}

	outBytes, err := json.Marshal(map[string]interface{}{
		"message": "OK",
	})
	if err != nil {
		return err
	}

	fmt.Println(string(outBytes))

	return nil
}
