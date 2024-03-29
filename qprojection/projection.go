package qprojection

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/lipgloss"
)

type ColorFunc func(*todotxt.List, *todotxt.Item) *lipgloss.Color

type Projector struct {
	Clean         []string
	ScoreCalc     qscore.Calculator
	HumanizedTags []string
	TagTypes      map[string]qselect.DType
	TagColors     map[string]ColorFunc
	LineColors    ColorFunc
	colorOverride *lipgloss.Color
	defaultColor  lipgloss.Color
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

func (p Projector) MustProject(projection []string, list *todotxt.List, selection []*todotxt.Item) ([]string, [][]string, [][]lipgloss.Style) {
	header, data, styles, err := p.Project(projection, list, selection)
	if err != nil {
		panic(err)
	}

	return header, data, styles
}

func (p Projector) Project(projection []string, list *todotxt.List, selection []*todotxt.Item) ([]string, [][]string, [][]lipgloss.Style, error) {
	headers, extractors, err := p.compile(projection, list)
	if err != nil {
		return nil, nil, nil, err
	}
	data := make([][]string, 0, len(selection))
	styles := make([][]lipgloss.Style, 0, len(selection))
	for _, i := range selection {
		p.colorOverride = p.LineColors(list, i)
		line, lineStyles := p.projectItem(list, i, extractors)
		data = append(data, line)
		styles = append(styles, lineStyles)
	}
	return headers, data, styles, nil
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

func (p Projector) projectItem(list *todotxt.List, item *todotxt.Item, columns []exFunc) ([]string, []lipgloss.Style) {
	data := make([]string, 0, len(columns))
	styles := make([]lipgloss.Style, 0, len(columns))
	for _, c := range columns {
		val, style := c(p, list, item)
		data = append(data, val)
		if p.colorOverride != nil {
			styles = append(styles, lipgloss.NewStyle().Foreground(p.colorOverride))
		} else {
			styles = append(styles, lipgloss.NewStyle().Foreground(style))
		}
	}
	return data, styles
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

func (p Projector) expandClean(item *todotxt.Item) (proj []todotxt.Project, ctx []todotxt.Context, tags []string) {
	return ExpandCleanExpression(item, p.Clean)
}

func ExpandCleanExpression(item *todotxt.Item, clean []string) (proj []todotxt.Project, ctx []todotxt.Context, tags []string) {
	for _, c := range clean {
		c := strings.TrimSpace(c)
		switch {
		case c == "+ALL":
			proj = append(proj, item.Projects()...)
		case c == "@ALL":
			ctx = append(ctx, item.Contexts()...)
		case c == "ALL":
			tags = append(tags, item.Tags().Keys()...)
		case strings.HasPrefix(c, "@"):
			ctx = append(ctx, todotxt.Context(c[1:]))
		case strings.HasPrefix(c, "+"):
			proj = append(proj, todotxt.Project(c[1:]))
		case len(c) == 0:
			continue
		default:
			tags = append(tags, strings.TrimPrefix(c, "tag:"))
		}
	}
	return
}
