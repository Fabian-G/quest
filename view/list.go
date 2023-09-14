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

var columnIds = []string{
	idxColumnId,
	doneColumnId,
	descriptionColumnId,
	creationColumnId,
	completionColumnId,
	projectsColumnId,
	contextsColumnId,
	tagsColumnId,
}

const (
	idxColumnId         = "idx"
	doneColumnId        = "done"
	descriptionColumnId = "description"
	creationColumnId    = "creation"
	completionColumnId  = "completion"
	projectsColumnId    = "projects"
	contextsColumnId    = "contexts"
	tagsColumnId        = "tags"
)

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
		case p == idxColumnId:
			fns = append(fns, columnExtractor{"Idx", idxColumn})
		case p == doneColumnId:
			fns = append(fns, columnExtractor{"Done", doneColumn})
		case p == descriptionColumnId:
			fns = append(fns, columnExtractor{"Description", descriptionColumn})
		case p == creationColumnId:
			fns = append(fns, columnExtractor{"Created On", creationColumn})
		case p == completionColumnId:
			fns = append(fns, columnExtractor{"Completed On", completionColumn})
		case p == projectsColumnId:
			fns = append(fns, columnExtractor{"Projects", projectsColumn})
		case p == contextsColumnId:
			fns = append(fns, columnExtractor{"Contexts", contextsColumn})
		case p == tagsColumnId:
			fns = append(fns, l.allTagsExtractors()...)
		default:
			return nil, fmt.Errorf("unknown column: %s\nAvailable columns are: %v", p, columnIds)
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
