package orc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os/exec"
)

// CommandType regroups the supported command types.
type CommandType string

// Different supported command types.
const (
	Shell   CommandType = "shell"
	Network CommandType = "network"
)

// Command represents a command supported by a plugin.
type Command struct {
	Type      CommandType `json:"type"`
	Command   string      `json:"command"`
	Arguments []string    `json:"arguments"`
	Block     bool        `json:"block"`
}

// Execute executes a shell command and returns the output.
func (c Command) Execute(userArguments map[string]interface{}) ([]byte, error) {
	var totalArguments []string
	if userArguments != nil {
		dumpedBytes, err := json.Marshal(userArguments)
		if err != nil {
			return nil, err
		}
		totalArguments = append(c.Arguments, string(dumpedBytes))
	} else {
		totalArguments = c.Arguments
	}

	if c.Type == Shell {
		cmd := exec.Command(c.Command, totalArguments...)
		if c.Block {
			return cmd.CombinedOutput()
		}
		return nil, cmd.Start()

	} else if c.Type == Network {
		body := map[string]interface{}{
			"arguments": totalArguments,
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}

		resp, err := http.Post(c.Command, "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != 200 {
			return nil, errors.New(string(respBody))
		}

		return respBody, nil

	} else {
		return nil, errors.New("not implemented")
	}
}
