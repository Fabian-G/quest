package config

import (
	"github.com/Fabian-G/quest/query"
)

func buildQueryCompiler() *query.Compiler {
	return &query.Compiler{
		TagTypes: TagTypes(),
	}
}
