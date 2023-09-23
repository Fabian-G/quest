package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/spf13/viper"
)

type Key = string

const (
	TodoFile    Key = "todo-file"
	DoneFile    Key = "done-file"
	IdxOrder    Key = "index-order"
	KeepBackups Key = "backup"
	Interactive Key = "interactive"
	Editor      Key = "editor"
)

func buildConfig() (*viper.Viper, error) {
	v := viper.New()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(fmt.Errorf("could not determine users home directory: %w", err))
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if len(configHome) == 0 {
		configHome = path.Join(homeDir, ".config")
	}

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(path.Join(configHome, "quest"))
	v.AddConfigPath("$HOME/.quest/")
	setTopLevelDefaults(v, homeDir)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return v, nil
		}
		return nil, err
	}
	return v, nil
}

func setTopLevelDefaults(v *viper.Viper, homeDir string) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if len(dataHome) == 0 {
		dataHome = path.Join(homeDir, ".local/share")
	}
	v.SetDefault(IdxOrder, "+done,-creation,+description")
	v.SetDefault(TodoFile, path.Join(dataHome, "quest/todo.txt"))
	v.SetDefault(DoneFile, path.Join(dataHome, "quest/done.txt"))
	v.SetDefault(KeepBackups, 5)
	v.SetDefault(Editor, getDefaultEditor())
}

func getDefaultEditor() string {
	var possibleEditors []string = []string{os.Getenv("EDITOR")}
	switch runtime.GOOS {
	case "windows":
		possibleEditors = append(possibleEditors, "notepad.exe")
	case "linux":
		possibleEditors = append(possibleEditors, "nano", "nvim", "vim", "vi", "emacs", "ed")
	}
	for _, editor := range possibleEditors {
		if p, err := exec.LookPath(editor); err == nil {
			return p
		}
	}
	return ""
}
