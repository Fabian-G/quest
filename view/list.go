package view

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Fabian-G/quest/query"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type List struct {
	list           *todotxt.List
	query          query.Func
	table          table.Model
	interactive    bool
	rowIndexToItem map[int]*todotxt.Item
}

func NewList(list *todotxt.List, queryString string, interactive bool) (List, error) {
	qFunc, err := query.Compile(queryString, query.FOL)
	if err != nil {
		return List{}, err
	}
	columns := []table.Column{
		{
			Title: "Idx",
			Width: 3,
		},
		{
			Title: "Done",
			Width: 4,
		},

		{
			Title: "Creation",
			Width: 10,
		},
		{
			Title: "Projects",
			Width: 8,
		},
		{
			Title: "Contexts",
			Width: 8,
		},
		{
			Title: "Description",
			Width: 30,
		},
	}

	l := List{
		list:        list,
		query:       qFunc,
		interactive: interactive,
	}
	l.rowIndexToItem = make(map[int]*todotxt.Item)
	items := qFunc.Filter(list)
	rows := make([]table.Row, 0, len(items))
	for i, item := range items {
		rows = append(rows, l.rowFromItem(item))
		l.rowIndexToItem[i] = item
	}

	var height int
	var styles = table.DefaultStyles()
	var cursor int
	if l.interactive {
		height = 10
		cursor = 0
	} else {
		height = len(rows) + 1
		styles.Selected = styles.Cell
		cursor = math.MaxInt
	}

	l.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(height),
		table.WithFocused(l.interactive),
		table.WithStyles(styles),
	)
	l.table.SetCursor(cursor)

	return l, nil
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
	selectedItem, ok := l.rowIndexToItem[l.table.Cursor()]
	if !ok || !l.interactive {
		return builder.String()
	}
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

func (l List) rowFromItem(i *todotxt.Item) table.Row {
	columns := make([]string, 0, 7)
	columns = append(columns, fmt.Sprintf("%d", l.list.IndexOf(i)))
	if i.Done() {
		columns = append(columns, "x")
	} else {
		columns = append(columns, " ")
	}
	if i.CreationDate() != nil {
		columns = append(columns, i.CreationDate().Format(time.DateOnly))
	} else {
		columns = append(columns, "")
	}
	projects := make([]string, 0)
	for _, p := range i.Projects() {
		projects = append(projects, p.String())
	}
	columns = append(columns, strings.Join(projects, ","))
	contexts := make([]string, 0)
	for _, c := range i.Contexts() {
		contexts = append(contexts, c.String())
	}
	columns = append(columns, strings.Join(contexts, ","))
	columns = append(columns, i.Description())
	return columns
}
