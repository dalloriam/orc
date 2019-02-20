package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

const (
	moduleName = "task"

	// MaintenanceLoopFrequencyMs is the frequency at which we check for changes in task status.
	MaintenanceLoopFrequencyMs = 500
)

// Controller defines available docker interactions
type Controller struct {
	defsDirectory string
	tasks         map[string]taskDef

	RunningTasks map[string]chan bool

	shouldInitializeTasks bool
}

// NewController loads the task definitions and returns a new controller.
func NewController(definitionsDirectory string, initializeTasks bool) (*Controller, error) {
	cont := &Controller{
		defsDirectory:         definitionsDirectory,
		RunningTasks:          make(map[string]chan bool),
		shouldInitializeTasks: initializeTasks,
	}
	if err := cont.loadTasks(); err != nil {
		return nil, err
	}

	logrus.Infof("%s module loaded successfully", moduleName)
	return cont, nil
}

// Name returns the name of the module.
func (c *Controller) Name() string {
	return moduleName
}

// Actions returns the actions defined by the module
func (c *Controller) Actions() []string {
	return []string{"start", "stop", "running"}
}

// AddTask adds the task to the controller.
func (c *Controller) AddTask(name string, t taskDef) {
	if c.tasks == nil {
		c.tasks = make(map[string]taskDef)
	}
	c.tasks[name] = t
}

func (c *Controller) loadTasks() error {
	ctxLog := logrus.WithFields(logrus.Fields{
		"module": moduleName,
	})

	files, err := ioutil.ReadDir(c.defsDirectory)
	if err != nil {
		return fmt.Errorf("invalid task directory: %s", c.defsDirectory)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		taskPath := path.Join(c.defsDirectory, f.Name())

		data, err := ioutil.ReadFile(taskPath)
		if err != nil {
			return err
		}

		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return err
		}

		c.AddTask(task.Name, &task)

		if c.shouldInitializeTasks {
			if err := task.Initialize(); err != nil {
				return err
			}
		} else {
			ctxLog.Warn("skipping task initialization as requested in controller configuration.")
		}
		ctxLog.Infof("task loaded successfully: %s", task.Name)

		isRunning, err := task.IsRunning()
		if err != nil {
			return err
		}

		if isRunning {
			logrus.Infof("hooking into already running task: %s", task.Name)
			go c.manageLifecycle(task.Name, &task)
		}
	}

	return nil
}

// Execute executes an action.
func (c *Controller) Execute(actionName string, data map[string]interface{}) ([]byte, error) {
	switch actionName {
	case "start":
		var args StartPayload
		if err := mapstructure.Decode(data, &args); err != nil {
			return nil, err
		}
		if err := c.Start(args.TaskName); err != nil {
			return nil, err
		}
	case "stop":
		var args StartPayload
		if err := mapstructure.Decode(data, &args); err != nil {
			return nil, err
		}
		if err := c.Stop(args.TaskName); err != nil {
			return nil, err
		}
	case "running":
		return json.Marshal(map[string]interface{}{
			"message": "OK",
			"tasks":   c.getRunningTasks(),
		})
	default:
		return nil, fmt.Errorf("unknown action: %s", actionName)
	}
	return json.Marshal(map[string]interface{}{"message": "OK"})
}

func (c *Controller) getRunningTasks() []string {
	tasks := []string{}
	for k := range c.RunningTasks {
		tasks = append(tasks, k)
	}
	return tasks
}

// Manages the lifecycle (status & cleanup) of a running task
func (c *Controller) manageLifecycle(name string, task taskDef) {
	if _, ok := c.RunningTasks[name]; ok {
		// If this is true, task is already managed by another goroutine.
		return
	}

	outChan := make(chan bool)
	c.RunningTasks[name] = outChan

	go func() {
		ctxLog := logrus.WithFields(logrus.Fields{
			"module": moduleName,
			"task":   name,
		})

		// No matter how we exit, cleanup must be performed.
		defer func() {
			close(outChan)
			delete(c.RunningTasks, name)

			if err := task.Cleanup(); err != nil {
				ctxLog.Errorf("error cleaning up [%s]: %s", name, err.Error())
			}
		}()

		// Assume the current service is running
		isRunning := true
		var err error

		for isRunning {
			isRunning, err = task.IsRunning()
			if err != nil {
				ctxLog.Errorf("error fetching status: %s", err.Error())
				return
			}

			time.Sleep(time.Duration(MaintenanceLoopFrequencyMs * time.Millisecond))
		}

		ctxLog.Info("task complete")

		nextTasks, err := task.NextTasks()
		if err != nil {
			ctxLog.Errorf("error fetching next tasks: %s", err.Error())
		}

		for _, taskName := range nextTasks {
			if err := c.Start(taskName); err != nil {
				ctxLog.Errorf("error starting connex task [%s]: %s", taskName, err.Error())
			}
		}
	}()
}

// Start runs the container as task.
func (c *Controller) Start(taskName string) error {
	if task, ok := c.tasks[taskName]; ok {
		// Start the task from the definition
		isRunning, err := task.IsRunning()
		if err != nil {
			return err
		}

		if !isRunning {
			if err := task.Start(); err != nil {
				return err
			}
		} else {
			logrus.Infof("task [%s] is already running", taskName)
		}

		// Run the task
		c.manageLifecycle(taskName, task)
		return nil
	}

	return fmt.Errorf("unknown task: %s", taskName)
}

// Stop stops a task.
func (c *Controller) Stop(taskName string) error {
	if task, ok := c.tasks[taskName]; ok {
		isRunning, err := task.IsRunning()
		if err != nil {
			return err
		}

		if !isRunning {
			logrus.Infof("task [%s] is not running", taskName)
			return nil
		}

		return task.Stop()
	}

	return fmt.Errorf("unknown task: %s", taskName)
}
