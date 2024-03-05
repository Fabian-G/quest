package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	version       = "dev"
	commit        = "none"
	date          = "unknown"
	builtBy       = "unknown"
	questAsciiArt = `
████████▄   ███    █▄     ▄████████    ▄████████     ███     
███    ███  ███    ███   ███    ███   ███    ███ ▀█████████▄ 
███    ███  ███    ███   ███    █▀    ███    █▀     ▀███▀▀██ 
███    ███  ███    ███  ▄███▄▄▄       ███            ███   ▀ 
███    ███  ███    ███ ▀▀███▀▀▀     ▀███████████     ███     
███    ███  ███    ███   ███    █▄           ███     ███     
███  ▀ ███  ███    ███   ███    ███    ▄█    ███     ███     
 ▀██████▀▄█ ████████▀    ██████████  ▄████████▀     ▄████▀   
                                                             
`
)

type versionCommand struct{}

func newVersionCommand() *versionCommand {
	return &versionCommand{}
}

func (v *versionCommand) command() *cobra.Command {
	var versionCommand = &cobra.Command{
		Use:     "version",
		Short:   "Outputs the installed version of the programme",
		GroupID: "global-cmd",
		RunE:    v.version,
	}
	return versionCommand
}

func (v *versionCommand) version(cmd *cobra.Command, args []string) error {
	fmt.Println(questAsciiArt)
	fmt.Println("quest: todo.txt CLI for (almost) any workflow.")
	fmt.Println("https://fabian-g.github.io/quest/")
	fmt.Println()
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("BuildDate: %s\n", date)
	fmt.Printf("BuiltBy: %s\n", builtBy)
	timewarrior, err := exec.LookPath("timew")
	var timewVersion string
	if err != nil {
		timewarrior = "n/a"
	} else {
		cmdOut, _ := exec.Command(timewarrior, "--version").Output()
		timewVersion = string(bytes.TrimSpace(cmdOut))
	}
	fmt.Printf("Timewarrior: %s (v%s)\n", timewarrior, timewVersion)
	return nil
}
