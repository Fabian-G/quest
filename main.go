package main

import (
	"os"
	"strings"

	"github.com/Fabian-G/quest/cmd"
	"github.com/Fabian-G/quest/config"
)

func main() {
	args := os.Args[1:]
	cmd.Execute(&config.Di{
		ConfigFile: configFromArgs(args),
	}, args)
}

func configFromArgs(args []string) string {
	// Unfortunately we have to grab the --config flag before running cobra, because
	// the construction of the views need the config, but at construction time the
	// flags are not parsed yet.
	for i, arg := range args {
		if arg == "--" {
			return ""
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.Split(arg, "=")[1]
		}
		if strings.HasPrefix(arg, "--config") && i < len(args)-1 {
			return args[i+1]
		}
	}
	return ""
}
