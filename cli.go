package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	cliCommandName = "cli"
	cliCommandHelp = "Interact with the ORC server."
	cliCommandArgs = "MODULE ACTION [OPTIONS]"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type cliCommand struct {
	arguments stringSlice
}

func (cmd *cliCommand) Name() string      { return cliCommandName }
func (cmd *cliCommand) Args() string      { return cliCommandArgs }
func (cmd *cliCommand) ShortHelp() string { return cliCommandHelp }
func (cmd *cliCommand) LongHelp() string  { return cliCommandHelp }
func (cmd *cliCommand) Hidden() bool      { return false }
func (cmd *cliCommand) Register(fs *flag.FlagSet) {
	fs.Var(&cmd.arguments, "a", "Pass argument to the action")
	fs.Var(&cmd.arguments, "argument", "Pass argument to the action")
}

func (cmd *cliCommand) parseArgumentPairs(args []string) (map[string]string, error) {
	argumentPairs := make(map[string]string)

	for i := 0; i < len(args); i++ {
		splitted := strings.Split(args[i], "=")
		if len(splitted) != 2 {
			return nil, fmt.Errorf("invalid argument syntax: expected 1 '=', got %d", len(splitted)-1)
		}
		argumentPairs[splitted[0]] = splitted[1]
	}

	return argumentPairs, nil
}

func (cmd *cliCommand) formatURL(module, action string) string {
	return fmt.Sprintf("http://%s:%d/%s/%s", serverHost, serverPort, module, action)
}

func (cmd *cliCommand) sendCommand(module, action string, body map[string]string) (map[string]interface{}, error) {

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)

	resp, err := http.Post(cmd.formatURL(module, action), "application/json", buf)

	if err != nil {
		return nil, err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var structured map[string]interface{}
	if err := json.Unmarshal(respData, &structured); err != nil {
		return nil, err
	}

	return structured, nil
}

func (cmd *cliCommand) pprintResponse(response map[string]interface{}) error {
	out, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

func (cmd *cliCommand) Run(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return errors.New("Invalid syntax")
	}

	argPairs, err := cmd.parseArgumentPairs(cmd.arguments)
	if err != nil {
		return err
	}

	output, err := cmd.sendCommand(args[0], args[1], argPairs)

	if err != nil {
		return err
	}

	return cmd.pprintResponse(output)
}
