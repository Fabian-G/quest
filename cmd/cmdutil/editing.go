package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func AskRetry(cmd *cobra.Command, err error) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "your changes are invalid: %s\n", err)
	fmt.Fprint(cmd.OutOrStdout(), "Retry? (Y/n) ")
	var answer string
	fmt.Fscanln(cmd.InOrStdin(), &answer)
	return strings.ToLower(answer) != "n"
}

func StartEditor(editorCmd string, path string) error {
	cmd := exec.Command(editorCmd, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor command failed: %w", err)
	}
	return nil
}
