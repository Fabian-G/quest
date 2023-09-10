package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type columnExtractor struct {
	title     string
	extractor func(*todotxt.Item) string
}

type List struct {
	list        *todotxt.List
	selection   []*todotxt.Item
	table       table.Model
	interactive bool
}

func NewList(list *todotxt.List, selection []*todotxt.Item, projection []string, interactive bool) (List, error) {
	l := List{
		list:        list,
		selection:   selection,
		interactive: interactive,
	}

	columnExtractors, err := l.columnExtractors(projection)
	if err != nil {
		return List{}, fmt.Errorf("invalid projection %v: %w", projection, err)
	}

	rows, columns := l.mapToColumns(columnExtractors)

	l.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(16),
		table.WithFocused(l.interactive),
	)

	return l, nil
}

func (l List) mapToColumns(extractors []columnExtractor) ([]table.Row, []table.Column) {
	columns := make([]table.Column, 0, len(extractors))
	rows := make([]table.Row, len(l.selection))
	for _, c := range extractors {
		maxWidth := 0
		values := make([]string, 0, len(extractors))
		for _, i := range l.selection {
			val := c.extractor(i)
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
		switch p {
		case "idx":
			fns = append(fns, columnExtractor{"Idx", l.idxColumn})
		case "description":
			fns = append(fns, columnExtractor{"Description", l.descriptionColumn})
		default:
			return nil, fmt.Errorf("unknown column in specification: %s", p)
		}
	}

	return fns, nil
}

func (l List) idxColumn(i *todotxt.Item) string {
	return strconv.Itoa(l.list.IndexOf(i))
}

func (l List) descriptionColumn(i *todotxt.Item) string {
	return runewidth.Truncate(i.Description(), 50, "...")
}

func (l List) Init() tea.Cmd {
	if l.interactive {
		return nil
	}
	return tea.Quit
}

func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return l, tea.Quit
		}
	}
	m, cmd := l.table.Update(msg)
	l.table = m
	return l, cmd
}

func (l List) View() string {
	builder := strings.Builder{}
	builder.WriteString(l.table.View())
	if !l.interactive {
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
