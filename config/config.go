package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

const (
	TodoFile = "todo-file"
	IdxOrder = "index-order"
)

func init() {
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

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(path.Join(configHome, "quest"))
	viper.AddConfigPath("$HOME/.quest/")
	viper.SetDefault(IdxOrder, "+done,-creation,+description")
	viper.SetDefault(TodoFile, path.Join(dataHome, "quest/todo.txt"))

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return
		}
		log.Fatal(err)
	}
	registerMacros()
}
