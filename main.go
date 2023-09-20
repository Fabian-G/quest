package main

import (
	"github.com/Fabian-G/quest/cmd"
	"github.com/Fabian-G/quest/config"
)

func main() {
	cmd.Execute(&config.Di{})
}
