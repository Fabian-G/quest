package config

import (
	"github.com/Fabian-G/quest/query"
	"github.com/Fabian-G/quest/todotxt"
)

type Di struct {
	repo     *todotxt.Repo
	compiler *query.Compiler
}

func (d *Di) TodoTxtRepo() *todotxt.Repo {
	if d.repo == nil {
		d.repo = buildTodoTxtRepo()
	}
	return d.repo
}

func (d *Di) QueryCompiler() *query.Compiler {
	if d.compiler == nil {
		d.compiler = buildQueryCompiler()
	}
	return d.compiler
}
