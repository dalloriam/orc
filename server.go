package main

import (
	"context"
	"flag"
	"os"
	"os/user"
	"path"

	"github.com/dalloriam/orc/interfaces"
	log "github.com/sirupsen/logrus"
)

const (
	serverCommandName = "server"
	serverCommandArgs = "[--docker-defs /path/to/docker/defs/directory] [--plugin-dir /path/to/plugin/dir]"
	serverCommandHelp = "Starts the ORC server."

	defaultDockerPathSuffix = ".config/dalloriam/orc/docker"
	defaultPluginDirSuffix = ".config/dalloriam/orc/plugins"
)

func getHomeDir() (string, error){
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func createDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}


type serverCommand struct {
	dockerDefsDir string
	pluginsDir string
}

func (cmd *serverCommand) Name() string {return serverCommandName}
func (cmd *serverCommand) Args() string {return serverCommandArgs}
func (cmd *serverCommand) ShortHelp() string {return serverCommandHelp}
func (cmd *serverCommand) LongHelp() string {return serverCommandHelp}
func (cmd *serverCommand) Hidden() bool { return false   }

func (cmd *serverCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.dockerDefsDir, "docker_defs_path", "", "Path to docker definitions directory. (defaults to ~/.config/dalloriam/orc/docker)")
	fs.StringVar(&cmd.pluginsDir, "plugins_dir", "", "Path to the plugins directory. (defaults to ~/.config/dalloriam/orc/plugins)")
}

func (cmd *serverCommand) Run(ctx context.Context, args []string) error {
	if cmd.dockerDefsDir == "" {
		homeDir, err := getHomeDir()
		if err != nil {
			return err
		}

		cmd.dockerDefsDir = path.Join(homeDir, defaultDockerPathSuffix)
	}

	if cmd.pluginsDir == "" {
		homeDir, err := getHomeDir()
		if err != nil {
			return err
		}
		cmd.pluginsDir = path.Join(homeDir, defaultPluginDirSuffix)
	}

	if err := createDirIfNotExists(cmd.dockerDefsDir); err != nil {
		return err
	}

	if err := createDirIfNotExists(cmd.pluginsDir); err != nil {
		return err
	}

	o, err := New(cmd.dockerDefsDir, cmd.pluginsDir, interfaces.HandleWithHTTP)

	if err != nil {
		return err
	}

	log.Fatal(o.Serve("0.0.0.0", 8080))

	return nil
}