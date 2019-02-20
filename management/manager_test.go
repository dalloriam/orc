package management_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/dalloriam/orc/management"
)

const (
	actionActionsAvailable = "actions_available"
	moduleName             = "manage"
)

func TestNewModule(t *testing.T) {
	mod := management.NewModule()

	if mod == nil {
		t.Errorf("returned nil module")
	}
}

func TestModule_Name(t *testing.T) {
	mod := &management.Module{}
	actual := mod.Name()
	if actual != moduleName {
		t.Errorf("expected module name to be %s, got %s", moduleName, actual)
	}
}

func TestModule_Actions(t *testing.T) {
	mod := &management.Module{}

	expectedActions := []string{actionActionsAvailable}
	actualActions := mod.Actions()

	for i := 0; i < len(expectedActions); i++ {
		if i >= len(actualActions) {
			t.Errorf("not enough actions returned. got %d, expected %d", len(actualActions), len(expectedActions))
		}

		if expectedActions[i] != actualActions[i] {
			t.Errorf("expected action at idx=%d to be %s, got %s instead", i, expectedActions[i], actualActions[i])
		}
	}
}

func TestModule_Execute(t *testing.T) {
	type testCase struct {
		name       string
		actionName string

		registeredActions map[string]string

		expectedOutput map[string][]string
		wantErr        bool
	}

	cases := []testCase{
		{
			name:           "unknown action",
			actionName:     "woah",
			expectedOutput: nil,
			wantErr:        true,
		},
		{
			name:           "actions_available, none defined",
			actionName:     actionActionsAvailable,
			expectedOutput: make(map[string][]string),
			wantErr:        false,
		},
		{
			name:       "actions_available, some defined",
			actionName: actionActionsAvailable,
			expectedOutput: map[string][]string{
				"hello": []string{"there"},
			},
			registeredActions: map[string]string{
				"hello": "there",
			},
			wantErr: false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			// Initialize management module.
			mod := management.NewModule()

			// Insert actions expected by the test.
			if tCase.registeredActions != nil {
				for k, v := range tCase.registeredActions {
					mod.RegisterAction(k, v)
				}
			}

			actualBytes, err := mod.Execute(tCase.actionName, nil)

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected err==nil: %v, got err=%v instead", tCase.wantErr, err)
				return
			}

			var actualUnmarshaled map[string][]string
			if actualBytes != nil {
				if err := json.Unmarshal(actualBytes, &actualUnmarshaled); err != nil {
					t.Errorf("unexpected exception: execute returned invalid JSON: %s", actualBytes)
					return
				}

				for k, v := range tCase.expectedOutput {
					actual := actualUnmarshaled[k]

					if !reflect.DeepEqual(v, actual) {
						t.Errorf("expected output [%v], got [%v] instead", v, actual)
						return

					}
				}
			} else {
				if tCase.expectedOutput != nil {
					t.Errorf("expected nil output, got %v", actualBytes)
					return
				}
			}
		})
	}
}
