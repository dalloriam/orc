package task_test

import (
	"errors"
	"testing"

	"github.com/dalloriam/orc/task"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.PanicLevel)
}

const (
	expectedControllerName = "task"
)

var expectedActions = []string{"start", "stop"}

type mockService struct {
	ShouldFail bool
}

func (m *mockService) Start() error {
	if m.ShouldFail {
		return errors.New("Something terrible happened")
	}
	return nil
}

func TestNewController(t *testing.T) {
	type testCase struct {
		name        string
		testDataDir string

		wantErr bool
	}

	cases := []testCase{
		{"simple case", "./testdata/simple_defs", false},
		{"non-existent dir", "./testdata/doesnt_exist", true},
		{"bad json", "./testdata/bad_json", true},
	}

	for _, tCase := range cases {
		controller, err := task.NewController(tCase.testDataDir, false)

		if tCase.wantErr {
			if err == nil {
				t.Errorf("expected error, got none")
				return
			}
		} else {
			if err != nil {
				t.Errorf("expected no error, got %s", err.Error())
				return
			}
			if controller == nil {
				t.Errorf("returned nil controller")
			}

			actualName := controller.Name()
			if actualName != expectedControllerName {
				t.Errorf("expected name=%s, got: %s", expectedControllerName, actualName)
			}

			actions := controller.Actions()
			for i := 0; i < len(expectedActions); i++ {
				if i >= len(actions)+1 {
					t.Errorf("too few actions")
					return
				}
				if expectedActions[i] != actions[i] {
					t.Errorf("expected actions=%s, got: %s", expectedActions[i], actions[i])
				}
			}
		}
	}
}
