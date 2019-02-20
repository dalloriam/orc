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

type mocktask struct {
	ShouldIsRunningFail bool
	ShouldStartFail     bool
	ShouldStopFail      bool
	ShouldCleanupFail   bool
	ShouldNextTasksFail bool

	CurrentlyRunning bool
	ChainedTasks     []string

	CallChain []string
}

func (m *mocktask) IsRunning() (bool, error) {
	if m.ShouldIsRunningFail {
		return false, errors.New("something terrible happened")
	}

	m.CallChain = append(m.CallChain, "is_running")

	return m.CurrentlyRunning, nil
}

func (m *mocktask) Start() error {
	if m.ShouldStartFail {
		return errors.New("Something terrible happened")
	}
	m.CallChain = append(m.CallChain, "start")
	m.CurrentlyRunning = true
	return nil
}

func (m *mocktask) Stop() error {
	if m.ShouldStopFail {
		return errors.New("something terrible happened")
	}
	m.CallChain = append(m.CallChain, "stop")
	m.CurrentlyRunning = false
	return nil
}

func (m *mocktask) Cleanup() error {
	if m.ShouldCleanupFail {
		return errors.New("something terrible happened")
	}
	m.CallChain = append(m.CallChain, "cleanup")
	m.CurrentlyRunning = false
	return nil
}

func (m *mocktask) NextTasks() ([]string, error) {
	if m.ShouldNextTasksFail {
		return nil, errors.New("something terrible happened")
	}
	m.CallChain = append(m.CallChain, "next_tasks")
	return m.ChainedTasks, nil
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

func TestController_Start(t *testing.T) {
	type testCase struct {
		name    string
		tasks   map[string]*mocktask
		wantErr bool

		taskToStart       string
		expectedCallChain []string
	}

	cases := []testCase{
		testCase{
			name: "all is normal",
			tasks: map[string]*mocktask{
				"hello": &mocktask{},
			},
			wantErr:           false,
			taskToStart:       "hello",
			expectedCallChain: []string{"is_running", "start"},
		},
		testCase{
			name: "already running",
			tasks: map[string]*mocktask{
				"hello": &mocktask{CurrentlyRunning: true},
			},
			wantErr:           false,
			taskToStart:       "hello",
			expectedCallChain: []string{"is_running"},
		},
		testCase{
			name: "is_running fails",
			tasks: map[string]*mocktask{
				"hello": &mocktask{ShouldIsRunningFail: true},
			},
			wantErr:           true,
			taskToStart:       "hello",
			expectedCallChain: []string{},
		},
		testCase{
			name: "start fails",
			tasks: map[string]*mocktask{
				"hello": &mocktask{ShouldStartFail: true},
			},
			wantErr:           true,
			taskToStart:       "hello",
			expectedCallChain: []string{"is_running"},
		},
		testCase{
			name:              "unknown task",
			tasks:             make(map[string]*mocktask),
			wantErr:           true,
			taskToStart:       "hello",
			expectedCallChain: []string{},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			c := &task.Controller{
				RunningTasks: make(map[string]chan bool),
			}

			for k, v := range tCase.tasks {
				c.AddTask(k, v)
			}

			err := c.Start(tCase.taskToStart)

			for i := 0; i < len(tCase.expectedCallChain); i++ {
				actualCallChain := tCase.tasks[tCase.taskToStart].CallChain
				if i >= len(actualCallChain) {
					t.Errorf("not enough calls in call chain, expected %v, got %v", tCase.expectedCallChain, actualCallChain)
					return
				}

				if tCase.expectedCallChain[i] != actualCallChain[i] {
					t.Errorf("expected callchain item %d to be %s, got %s", i, tCase.expectedCallChain[i], actualCallChain[i])
				}
			}

			if tCase.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %s: ", err.Error())
				return
			}

			tCase.tasks[tCase.taskToStart].Stop()

			// Wait until task done.
			for _ = range c.RunningTasks[tCase.taskToStart] {
				// TODO: This is super ugly, find better way to do this
			}

			newChain := tCase.tasks[tCase.taskToStart].CallChain
			if newChain[len(newChain)-1] != "cleanup" {
				t.Errorf("expected last item to be 'cleanup', got %s", newChain)
			}
		})
	}
}

func TestController_Stop(t *testing.T) {
	type testCase struct {
		name    string
		tasks   map[string]*mocktask
		wantErr bool

		taskToStop        string
		expectedCallChain []string
	}

	cases := []testCase{
		testCase{
			name: "all is normal",
			tasks: map[string]*mocktask{
				"hello": &mocktask{CurrentlyRunning: true},
			},
			wantErr:           false,
			taskToStop:        "hello",
			expectedCallChain: []string{"is_running", "stop"},
		},
		testCase{
			name: "already stopped",
			tasks: map[string]*mocktask{
				"hello": &mocktask{},
			},
			wantErr:           false,
			taskToStop:        "hello",
			expectedCallChain: []string{"is_running"},
		},
		testCase{
			name: "is_running fails",
			tasks: map[string]*mocktask{
				"hello": &mocktask{ShouldIsRunningFail: true},
			},
			wantErr:           true,
			taskToStop:        "hello",
			expectedCallChain: []string{},
		},
		testCase{
			name: "stop fails",
			tasks: map[string]*mocktask{
				"hello": &mocktask{ShouldStopFail: true, CurrentlyRunning: true},
			},
			wantErr:           true,
			taskToStop:        "hello",
			expectedCallChain: []string{"is_running"},
		},
		testCase{
			name:              "unknown task",
			tasks:             make(map[string]*mocktask),
			wantErr:           true,
			taskToStop:        "hello",
			expectedCallChain: []string{},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			c := &task.Controller{
				RunningTasks: make(map[string]chan bool),
			}

			for k, v := range tCase.tasks {
				c.AddTask(k, v)
			}

			err := c.Stop(tCase.taskToStop)

			for i := 0; i < len(tCase.expectedCallChain); i++ {
				actualCallChain := tCase.tasks[tCase.taskToStop].CallChain
				if i >= len(actualCallChain) {
					t.Errorf("not enough calls in call chain, expected %v, got %v", tCase.expectedCallChain, actualCallChain)
					return
				}

				if tCase.expectedCallChain[i] != actualCallChain[i] {
					t.Errorf("expected callchain item %d to be %s, got %s", i, tCase.expectedCallChain[i], actualCallChain[i])
				}
			}

			if tCase.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %s: ", err.Error())
				return
			}
		})
	}
}
