package orc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
)

type CommandType string

const (
	Shell   CommandType = "shell"
	Network CommandType = "network"
)

type Command struct {
	Type      CommandType `json:"type"`
	Command   string      `json:"command"`
	Arguments []string    `json:"arguments"`
	Block     bool        `json:"block"`
}

func (c Command) Execute(userArguments map[string]interface{}) ([]byte, error) {
	dumpedBytes, err := json.Marshal(userArguments)
	if err != nil {
		return nil, err
	}

	totalArguments := append(c.Arguments, string(dumpedBytes))

	if c.Type == Shell {
		cmd := exec.Command(c.Command, totalArguments...)
		if c.Block {
			return cmd.Output()
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

func (c Command) getHTTPHandler(actionName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[ORC] - Running action %s\n", actionName)

		userData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var parsedData map[string]interface{}

		if len(userData) != 0 {
			if err := json.Unmarshal(userData, &parsedData); err != nil {
				outMsg, err := json.Marshal(map[string]string{"error": err.Error()})
				if err != nil {
					panic(err)
				}
				w.Write(outMsg)
				return
			}
		}

		outputBytes, err := c.Execute(parsedData)

		if err != nil {
			outMsg, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				panic(err)
			}
			w.Write(outMsg)
			return
		}

		outMsg, err := json.Marshal(map[string]string{"message": "OK", "output": string(outputBytes)})
		if err != nil {
			panic(err)
		}
		w.Write(outMsg)
	}
}
