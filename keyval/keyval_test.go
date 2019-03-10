package keyval_test

import (
	"encoding/json"
	"testing"

	"github.com/dalloriam/orc/keyval"
)

func TestNew(t *testing.T) {
	m := keyval.NewModule()

	if m == nil {
		t.Error("New() returned nil module")
	}
}

func TestModule_Name(t *testing.T) {
	m := &keyval.Module{}

	if m.Name() != "keyval" {
		t.Errorf("invalid name: %s", m.Name())
	}
}

func TestModule_Actions(t *testing.T) {
	m := &keyval.Module{}

	expected := []string{"get", "set", "del", "list"}
	actual := m.Actions()

	for i := 0; i < len(expected); i++ {
		if i >= len(actual) {
			t.Errorf("expected %d actions, got %d", len(expected), len(actual))
			return
		}

		if expected[i] != actual[i] {
			t.Errorf("expected actions[%d] to be %s, got %s", i, expected[i], actual[i])
		}
	}
}

func TestModule_Execute(t *testing.T) {
	type testCase struct {
		name string

		action string
		data   map[string]interface{}

		expectedParsedMap map[string]interface{}
		wantErr           bool
	}

	m := keyval.NewModule()

	cases := []testCase{
		{"fails when action unknown", "random", map[string]interface{}{"key": "hello"}, nil, true},
		{"fails when key not specified", "get", map[string]interface{}{}, nil, true},
		{"fails when key doesnt exist", "get", map[string]interface{}{"key": "hello"}, nil, true},
		{"sets keys correctly", "set", map[string]interface{}{"key": "hello", "val": "world"}, map[string]interface{}{"message": "OK"}, false},
		{"fails when setting empty", "set", map[string]interface{}{"key": "hello"}, nil, true},
		{"gets keys correctly", "get", map[string]interface{}{"key": "hello"}, map[string]interface{}{"value": "world"}, false},
		{"deletes keys correctly", "del", map[string]interface{}{"key": "hello"}, map[string]interface{}{"message": "OK"}, false},
		{"fails when key doesnt exist", "get", map[string]interface{}{"key": "hello"}, nil, true},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {

			out, err := m.Execute(tCase.action, tCase.data)

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected err: %v, got err=%v", tCase.wantErr, err)
				return
			}

			if out == nil {
				if tCase.expectedParsedMap != nil {
					t.Errorf("got nil output, expected %v", tCase.expectedParsedMap)
				}
				return
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(out, &parsed); err != nil {
				t.Errorf("module returned invalid JSON")
				return
			}

			for k, v := range tCase.expectedParsedMap {
				if acV, ok := parsed[k]; ok {
					if acV != v {
						t.Errorf("expected module[%s] to return %v, got %v", k, v, acV)
					}
				} else {
					t.Errorf("expected %s:%v to be in module", k, v)
					return
				}
			}
		})
	}
}
