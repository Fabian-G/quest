package config

import "github.com/spf13/viper"

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$XDG_CONFIG_HOME/quest/")
	viper.AddConfigPath("$HOME/.config/quest/")
	viper.AddConfigPath("$HOME/.quest/")
	viper.SetDefault("default.sortOrder", "+done,+creation,+description")
}
