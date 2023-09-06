package config

import "github.com/Fabian-G/quest/todotxt"

type Di struct {
	repo *todotxt.Repo
}

func (d *Di) TodoTxtRepo() *todotxt.Repo {
	if d.repo == nil {
		d.repo = BuildTodoTxtRepo()
	}
	return d.repo
}
