package plugins

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

	PluginDir string `json:"plugin_dir,omitempty"`
}

// Execute executes a shell command and returns the output.
func (c Command) Execute(userArguments map[string]interface{}) (map[string]interface{}, error) {

	if c.Type == Shell {
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
		cmd := exec.Command(c.Command, totalArguments...)

		if c.PluginDir != "" {
			cmd.Dir = c.PluginDir
		}

		if c.Block {
			outBytes, err := cmd.CombinedOutput()
			if err != nil {
				return nil, err
			}
			var responseData map[string]interface{}
			if err := json.Unmarshal(outBytes, &responseData); err != nil {
				return nil, err
			}
			return responseData, nil
		}
		return nil, cmd.Start()

	} else if c.Type == Network {
		// TODO: Handle error.
		bodyBytes, _ := json.Marshal(userArguments)

		resp, err := http.Post(c.Command, "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}

		// TODO: Handle error.
		respBody, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			return nil, errors.New(string(respBody))
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal(respBody, &responseData); err != nil {
			return nil, err
		}

		return responseData, nil

	} else {
		return nil, errors.New("not implemented")
	}
}
