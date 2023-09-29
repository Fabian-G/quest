package di

import (
	"fmt"
	"os"
	"os/exec"
)

type Editor interface {
	Edit(path string) error
}

type EditorFunc func(path string) error

func (e EditorFunc) Edit(path string) error {
	return e(path)
}

func buildEditor(c Config) Editor {
	command := c.Editor
	return EditorFunc(func(path string) error {
		cmd := exec.Command(command, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("editor command failed: %w", err)
		}
		return nil
	})
}
