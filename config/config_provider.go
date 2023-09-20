package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Key = string

const (
	TodoFile    Key = "todo-file"
	IdxOrder    Key = "index-order"
	KeepBackups Key = "backup"
	Interactive Key = "interactive"
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

	dataHome := os.Getenv("XDG_DATA_HOME")
	if len(dataHome) == 0 {
		dataHome = path.Join(homeDir, ".local/share")
	}

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(path.Join(configHome, "quest"))
	v.AddConfigPath("$HOME/.quest/")
	v.SetDefault(IdxOrder, "+done,-creation,+description")
	v.SetDefault(TodoFile, path.Join(dataHome, "quest/todo.txt"))
	v.SetDefault(KeepBackups, 5)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return v, nil
		}
		return nil, err
	}
	return v, nil
}
