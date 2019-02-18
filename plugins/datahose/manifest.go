package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/dalloriam/orc/plugins"
)

const (
	manCommandName = "manifest"
	manCommandHelp = "Returns the plugin manifest"
	manCommandArgs = ""
)

type manifestCommand struct{}

func (cmd *manifestCommand) Name() string              { return manCommandName }
func (cmd *manifestCommand) Args() string              { return manCommandArgs }
func (cmd *manifestCommand) ShortHelp() string         { return manCommandHelp }
func (cmd *manifestCommand) LongHelp() string          { return manCommandHelp }
func (cmd *manifestCommand) Hidden() bool              { return true }
func (cmd *manifestCommand) Register(fs *flag.FlagSet) {}

func (cmd *manifestCommand) Run(ctx context.Context, args []string) error {
	p := plugins.PluginManifest{
		PluginName: "datahose",
		ActionMap: map[string]plugins.Command{
			"notify": plugins.Command{
				Type:      "shell",
				Command:   "./datahose",
				Arguments: []string{"push"},
				Block:     true,
			},
		},
	}
	outBytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	fmt.Println(string(outBytes))
	return nil
}
