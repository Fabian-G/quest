package qprojection

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
)

type Projector struct {
	Clean []string
}

func (p Projector) Verify(projection []string, list *todotxt.List) error {
	realProjection := p.expandAliasColumns(projection, list)
	for _, columnId := range realProjection {
		_, err := p.findColumn(columnId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Projector) MustProject(projection []string, list *todotxt.List) ([]string, [][]string) {
	header, data, err := p.Project(projection, list)
	if err != nil {
		panic(err)
	}

	return header, data
}

func (p Projector) Project(projection []string, list *todotxt.List) ([]string, [][]string, error) {
	headers, extractors, err := p.compile(projection, list)
	if err != nil {
		return nil, nil, err
	}
	tasks := list.Tasks()
	data := make([][]string, 0, len(tasks))
	for _, i := range tasks {
		data = append(data, p.projectItem(list, i, extractors))
	}
	return headers, data, nil
}

func (p Projector) MustProjectItem(projection []string, list *todotxt.List, item *todotxt.Item) ([]string, []string) {
	header, data, err := p.ProjectItem(projection, list, item)
	if err != nil {
		panic(err)
	}

	return header, data
}

func (p Projector) ProjectItem(projection []string, list *todotxt.List, item *todotxt.Item) ([]string, []string, error) {
	headers, extractors, err := p.compile(projection, list)
	if err != nil {
		return nil, nil, err
	}
	return headers, p.projectItem(list, item, extractors), nil
}

func (p Projector) compile(projection []string, list *todotxt.List) ([]string, []exFunc, error) {
	realProjection := p.expandAliasColumns(projection, list)
	headers := make([]string, 0, len(realProjection))
	extractors := make([]exFunc, 0, len(realProjection))
	for _, columnId := range realProjection {
		c, err := p.findColumn(columnId)
		if err != nil {
			return nil, nil, err
		}
		headers = append(headers, c.name(columnId))
		extractors = append(extractors, c.extractor(columnId))
	}
	return headers, extractors, nil
}

func (p Projector) projectItem(list *todotxt.List, item *todotxt.Item, columns []exFunc) []string {
	data := make([]string, 0, len(columns))
	for _, c := range columns {
		data = append(data, c(p, list, item))
	}
	return data
}

func (p Projector) findColumn(key string) (columnDef, error) {
	for _, cDef := range columns {
		if cDef.matcher.match(key) {
			return cDef, nil
		}
	}
	return columnDef{}, fmt.Errorf("column %s not found. Available columns are: %v", key, availableColumns())
}

func (p Projector) expandAliasColumns(projection []string, list *todotxt.List) []string {
	realProjection := make([]string, 0, len(projection))
	for _, p := range projection {
		switch p {
		case "tags":
			tagKeys := list.AllTags()
			for _, key := range tagKeys {
				realProjection = append(realProjection, fmt.Sprintf("tag:%s", key))
			}
		default:
			realProjection = append(realProjection, p)
		}
	}
	return realProjection
}

func (p Projector) expandClean(list *todotxt.List) (proj []todotxt.Project, ctx []todotxt.Context, tags []string) {
	for _, c := range p.Clean {
		c := strings.TrimSpace(c)
		switch {
		case c == "+ALL":
			proj = append(proj, list.AllProjects()...)
		case c == "@ALL":
			ctx = append(ctx, list.AllContexts()...)
		case c == "ALL":
			tags = append(tags, list.AllTags()...)
		case strings.HasPrefix(c, "@"):
			ctx = append(ctx, todotxt.Context(c[1:]))
		case strings.HasPrefix(c, "+"):
			proj = append(proj, todotxt.Project(c[1:]))
		case len(c) == 0:
			continue
		default:
			tags = append(tags, c)
		}
	}
	return
}
