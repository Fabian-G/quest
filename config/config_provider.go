package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/spf13/viper"
)

type Key = string

const (
	TodoFileKey    Key = "todo-file"
	DoneFileKey    Key = "done-file"
	KeepBackupsKey Key = "backup"
	InteractiveKey Key = "interactive"
	EditorKey      Key = "editor"
	UnknownTagsKey Key = "unknown-tags"
	ViewsKey       Key = "view"
	MacrosKey      Key = "macro"
	TagsKey        Key = "tags"
	QuestScoreKey  Key = "quest-score"
	NowFuncKey     Key = "now-func-for-testing"
	ClearOnDone    Key = "clear-on-done"
)

func buildConfig(file string) (*viper.Viper, error) {
	v := viper.New()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(fmt.Errorf("could not determine users home directory: %w", err))
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if len(configHome) == 0 {
		configHome = path.Join(homeDir, ".config")
	}

	v.SetConfigType("toml")
	if file != "" {
		v.SetConfigFile(file)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(path.Join(configHome, "quest"))
		v.AddConfigPath("$HOME/.quest/")
	}
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
	v.SetDefault(TodoFileKey, path.Join(dataHome, "quest/todo.txt"))
	v.SetDefault(DoneFileKey, path.Join(dataHome, "quest/done.txt"))
	v.SetDefault(KeepBackupsKey, 5)
	v.SetDefault(EditorKey, getDefaultEditor())
	v.SetDefault(UnknownTagsKey, true)
	v.SetDefault(MacrosKey, []any{})
	v.SetDefault(ViewsKey, []any{})
	v.SetDefault(TagsKey, make(map[string]string))
	v.SetDefault(InteractiveKey, false)
	v.SetDefault(QuestScoreKey+".urgency-tag", "due")
	v.SetDefault(QuestScoreKey+".urgency-begin", 90)
	v.SetDefault(QuestScoreKey+".min-priority", "E")
	v.SetDefault(ClearOnDone, nil)
}

func getDefaultEditor() string {
	var possibleEditors []string = []string{os.Getenv("EDITOR"), "nano", "nvim", "vim", "vi", "emacs", "notepad.exe"}
	for _, editor := range possibleEditors {
		if p, err := exec.LookPath(editor); err == nil {
			return p
		}
	}
	return ""
}

func nowFunc(v *viper.Viper) func() time.Time {
	now := v.Get(NowFuncKey)
	if now == nil {
		return func() time.Time {
			return time.Now()
		}
	}
	return now.(func() time.Time)
}
