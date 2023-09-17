package view

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const StarProjection = "idx,done,completion,creation,projects,contexts,tags,description"

type columnExtractor struct {
	title     string
	extractor ExtractionFunc
}

type List struct {
	list          *todotxt.List
	selection     []*todotxt.Item
	extractors    []columnExtractor
	table         table.Model
	Interactive   bool
	CleanProjects []todotxt.Project
	CleanContexts []todotxt.Context
	CleanTags     []string
}

type refreshListMsg struct{}

func RefreshList() tea.Msg {
	return refreshListMsg{}
}

func NewList(list *todotxt.List, selection []*todotxt.Item, projection []string) (List, error) {
	l := List{
		list:      list,
		selection: selection,
	}

	columnExtractors, err := l.columnExtractors(projection)
	if err != nil {
		return List{}, fmt.Errorf("invalid projection %v: %w", projection, err)
	}
	l.extractors = columnExtractors
	l.table = table.New()
	return l, nil
}

func (l List) mapToColumns() ([]table.Row, []table.Column) {
	ctx := ExtractionCtx{
		List:          l.list,
		CleanTags:     l.CleanTags,
		CleanProjects: l.CleanProjects,
		CleanContexts: l.CleanContexts,
	}
	columns := make([]table.Column, 0, len(l.extractors))
	rows := make([]table.Row, len(l.selection))
	for _, c := range l.extractors {
		maxWidth := 0
		values := make([]string, 0, len(l.extractors))
		for _, i := range l.selection {
			ctx.Item = i
			val := c.extractor(ctx)
			values = append(values, val)
			maxWidth = max(maxWidth, len(val))
		}
		if maxWidth == 0 {
			continue
		}

		columns = append(columns, table.Column{Title: c.title, Width: max(maxWidth, len(c.title))})
		for i, v := range values {
			rows[i] = append(rows[i], v)
		}
	}
	return rows, columns
}

func (l List) columnExtractors(projection []string) ([]columnExtractor, error) {
	projection = l.expandAliasColumns(projection)
	fns := make([]columnExtractor, 0, len(projection))
	for _, p := range projection {
		name, extractor := findColumn(p)
		if extractor == nil {
			return nil, fmt.Errorf("unknown column: %s\nAvailable columns are: %v", p, availableColumns())
		}
		fns = append(fns, columnExtractor{name, extractor})
	}

	return fns, nil
}

func (l List) expandAliasColumns(projection []string) []string {
	realProjection := make([]string, 0, len(projection))
	for _, p := range projection {
		switch p {
		case "tags":
			tagKeys := make(map[string]struct{})
			for _, i := range l.selection {
				tags := i.Tags()
				for k := range tags {
					tagKeys[k] = struct{}{}
				}
			}
			for key := range tagKeys {
				realProjection = append(realProjection, fmt.Sprintf("tag:%s", key))
			}
		default:
			realProjection = append(realProjection, p)
		}
	}
	return realProjection
}

func (l List) Init() tea.Cmd {
	return RefreshList
}

func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return l, tea.Quit
		}
	case refreshListMsg:
		l = l.refreshTable()
	}
	m, cmd := l.table.Update(msg)
	l.table = m

	if !l.Interactive {
		return l, tea.Quit
	}
	return l, cmd
}

func (l List) View() string {
	builder := strings.Builder{}
	builder.WriteString(l.table.View())
	if !l.Interactive {
		return builder.String()
	}
	selectedItem := l.selection[l.table.Cursor()]
	builder.WriteString("\n")
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Index:\t\t%d\n", l.list.IndexOf(selectedItem)))
	if selectedItem.Done() {
		builder.WriteString("Status:\t\tdone\n")
	} else {
		builder.WriteString("Status:\t\tpending\n")
	}
	projects := make([]string, 0)
	for _, p := range selectedItem.Projects() {
		projects = append(projects, p.String())
	}
	builder.WriteString("Projects:\t")
	builder.WriteString(strings.Join(projects, ", "))
	builder.WriteString("\n")
	contexts := make([]string, 0)
	for _, c := range selectedItem.Contexts() {
		contexts = append(contexts, c.String())
	}
	builder.WriteString("Contexts:\t")
	builder.WriteString(strings.Join(contexts, ", "))
	builder.WriteString("\n")
	builder.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, "Description:\t", selectedItem.Description()))

	return builder.String()
}

func (l List) refreshTable() List {
	rows, columns := l.mapToColumns()
	l.table.SetColumns(columns)
	l.table.SetRows(rows)
	if l.Interactive {
		l.table.Focus()
		l.table.SetHeight(16)
		l.table.SetStyles(table.DefaultStyles())
	} else {
		l.table.SetHeight(len(rows))
		defaultStyles := table.DefaultStyles()
		defaultStyles.Selected = lipgloss.NewStyle()
		l.table.SetStyles(defaultStyles)
	}
	l.table.UpdateViewport()
	return l
}
