package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const StarProjection = "idx,done,completion,creation,projects,contexts,tags,description"

type columnExtractor struct {
	title     string
	extractor func(List, *todotxt.Item) string
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

func refreshList() tea.Msg {
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
	l.table = table.New(
		table.WithHeight(16),
		table.WithFocused(true),
	)
	return l, nil
}

func (l List) mapToColumns() ([]table.Row, []table.Column) {
	columns := make([]table.Column, 0, len(l.extractors))
	rows := make([]table.Row, len(l.selection))
	for _, c := range l.extractors {
		maxWidth := 0
		values := make([]string, 0, len(l.extractors))
		for _, i := range l.selection {
			val := c.extractor(l, i)
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
	fns := make([]columnExtractor, 0, len(projection))
	for _, p := range projection {
		switch {
		case strings.HasPrefix(p, "tag:"):
			tagKey := strings.Split(p, ":")[1]
			fns = append(fns, columnExtractor{tagKey, tagColumn(tagKey)})
		case p == "idx":
			fns = append(fns, columnExtractor{"Idx", idxColumn})
		case p == "done":
			fns = append(fns, columnExtractor{"Done", doneColumn})
		case p == "description":
			fns = append(fns, columnExtractor{"Description", descriptionColumn})
		case p == "creation":
			fns = append(fns, columnExtractor{"Created On", creationColumn})
		case p == "completion":
			fns = append(fns, columnExtractor{"Completed On", completionColumn})
		case p == "projects":
			fns = append(fns, columnExtractor{"Projects", projectsColumn})
		case p == "contexts":
			fns = append(fns, columnExtractor{"Contexts", contextsColumn})
		case p == "tags":
			fns = append(fns, l.allTagsExtractors()...)
		default:
			return nil, fmt.Errorf("unknown column in specification: %s", p)
		}
	}

	return fns, nil
}

func tagColumn(key string) func(List, *todotxt.Item) string {
	return func(l List, i *todotxt.Item) string {
		tagValues := i.Tags()[key]
		return strings.Join(tagValues, ",")
	}
}

func doneColumn(l List, i *todotxt.Item) string {
	if i.Done() {
		return "x"
	}
	return ""
}

func creationColumn(l List, i *todotxt.Item) string {
	date := i.CreationDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

func completionColumn(l List, i *todotxt.Item) string {
	date := i.CompletionDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

func projectsColumn(l List, i *todotxt.Item) string {
	projects := i.Projects()
	projectStrings := make([]string, 0, len(projects))
	for _, p := range projects {
		projectStrings = append(projectStrings, p.String())
	}
	return strings.Join(projectStrings, ",")
}

func contextsColumn(l List, i *todotxt.Item) string {
	contexts := i.Contexts()
	contextStrings := make([]string, 0, len(contexts))
	for _, p := range contexts {
		contextStrings = append(contextStrings, p.String())
	}
	return strings.Join(contextStrings, ",")
}

func idxColumn(l List, i *todotxt.Item) string {
	return strconv.Itoa(l.list.IndexOf(i))
}

func descriptionColumn(l List, i *todotxt.Item) string {
	return runewidth.Truncate(i.CleanDescription(l.CleanProjects, l.CleanContexts, l.CleanTags), 50, "...")
}

func (l List) allTagsExtractors() []columnExtractor {
	tagKeys := make(map[string]struct{})
	for _, i := range l.selection {
		tags := i.Tags()
		for k := range tags {
			tagKeys[k] = struct{}{}
		}
	}

	extractors := make([]columnExtractor, 0, len(tagKeys))
	for key := range tagKeys {
		extractors = append(extractors, columnExtractor{key, tagColumn(key)})
	}
	return extractors
}

func (l List) Init() tea.Cmd {
	if l.Interactive {
		return refreshList
	}
	return tea.Sequence(refreshList, tea.Quit)
}

func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return l, tea.Quit
		}
	case refreshListMsg:
		rows, columns := l.mapToColumns()
		l.table.SetColumns(columns)
		l.table.SetRows(rows)
	}
	m, cmd := l.table.Update(msg)
	l.table = m
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
