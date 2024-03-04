package di

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
	return EditorFunc(func(fileName string) error {
		wdBackup, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := os.Chdir(path.Dir(fileName)); err != nil {
			return err
		}
		defer func() {
			_ = os.Chdir(wdBackup)
		}()
		cmd := exec.Command(command, path.Base(fileName))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("editor command failed: %w", err)
		}
		return nil
	})
}
