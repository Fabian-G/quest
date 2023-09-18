package qprojection

import (
	"fmt"

	"github.com/Fabian-G/quest/todotxt"
)

type Config struct {
	ColumnNames   []string
	List          *todotxt.List
	CleanTags     []string
	CleanProjects []todotxt.Project
	CleanContexts []todotxt.Context
}

type Column struct {
	Title     string
	Projector Func
}

func Compile(cfg Config) ([]Column, error) {
	projectionColumns := expandAliasColumns(cfg, cfg.ColumnNames)

	columns := make([]Column, 0, len(projectionColumns))
	for _, p := range projectionColumns {
		column, err := findColumn(cfg, p)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

func expandAliasColumns(cfg Config, projection []string) []string {
	realProjection := make([]string, 0, len(projection))
	for _, p := range projection {
		switch p {
		case "tags":
			tagKeys := cfg.List.AllTags()
			for key := range tagKeys {
				realProjection = append(realProjection, fmt.Sprintf("tag:%s", key))
			}
		default:
			realProjection = append(realProjection, p)
		}
	}
	return realProjection
}
