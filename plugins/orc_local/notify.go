package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
)

const (
	notCommandName = "notify"
	notCommandHelp = "Send a native notification in a platform-agnostic way."
	notCommandArgs = "PAYLOAD"
)

type notifyPayload struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type notifyResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type notifyCommand struct{}

func (cmd *notifyCommand) Name() string              { return notCommandName }
func (cmd *notifyCommand) Args() string              { return notCommandArgs }
func (cmd *notifyCommand) ShortHelp() string         { return notCommandHelp }
func (cmd *notifyCommand) LongHelp() string          { return notCommandHelp }
func (cmd *notifyCommand) Hidden() bool              { return false }
func (cmd *notifyCommand) Register(fs *flag.FlagSet) {}

func (cmd *notifyCommand) linuxNotify(payload *notifyPayload) error {
	cmdObj := exec.Command("notify-send", payload.Title, payload.Message)
	output, err := cmdObj.CombinedOutput()

	if err != nil {
		return fmt.Errorf("error sending notification: %s", output)
	}

	return nil
}

func (cmd *notifyCommand) darwinNotify(payload *notifyPayload) error {
	cmdObj := exec.Command("/usr/bin/osascript", "-e", fmt.Sprintf("display notification \"%s\" with title \"%s\"", payload.Message, payload.Title))
	output, err := cmdObj.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error sending notification: %s", output)
	}
	return nil
}

func (cmd *notifyCommand) notify(payload *notifyPayload) error {
	switch runtime.GOOS {
	case "linux":
		return cmd.linuxNotify(payload)
	case "darwin":
		return cmd.darwinNotify(payload)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (cmd *notifyCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("json payload required")
	}

	var payload notifyPayload
	if err := json.Unmarshal([]byte(args[0]), &payload); err != nil {
		return err
	}

	var resp notifyResponse

	if err := cmd.notify(&payload); err != nil {
		resp = notifyResponse{Error: err.Error()}
	} else {
		resp = notifyResponse{Message: "OK"}
	}

	outBytes, err := json.Marshal(&resp)
	if err != nil {
		return err
	}

	fmt.Println(string(outBytes))
	return nil
}
