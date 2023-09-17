package config

import (
	"github.com/Fabian-G/quest/query"
)

func buildQueryCompiler() *query.Compiler {
	registerMacros()
	return &query.Compiler{
		TagTypes: TagTypes(),
	}
}
