package view

import (
	"fmt"
	"io"
	"strings"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var detailsProjection = strings.Split(strings.ReplaceAll(qprojection.StarProjection, ",tags", ""), ",")

type List struct {
	list        *todotxt.List
	selection   []*todotxt.Item
	extractors  []qprojection.Column
	table       table.Model
	interactive bool
}

type refreshListMsg struct{}

func RefreshList() tea.Msg {
	return refreshListMsg{}
}

func NewList(list *todotxt.List, selection []*todotxt.Item, interactive bool, pCfg qprojection.Config) (List, error) {
	l := List{
		list:        list,
		selection:   selection,
		interactive: interactive,
	}

	columnExtractors, err := qprojection.Compile(pCfg)
	if err != nil {
		return List{}, fmt.Errorf("invalid projection %v: %w", pCfg.ColumnNames, err)
	}
	l.extractors = columnExtractors
	l.table = table.New()
	return l.refreshTable(), nil
}

func (l List) mapToColumns() ([]table.Row, []table.Column) {
	columns := make([]table.Column, 0, len(l.extractors))
	rows := make([]table.Row, len(l.selection))
	for _, c := range l.extractors {
		maxWidth := 0
		values := make([]string, 0, len(l.extractors))
		for _, i := range l.selection {
			val := c.Projector(i)
			values = append(values, val)
			maxWidth = max(maxWidth, len(val))
		}
		if maxWidth == 0 {
			continue
		}

		columns = append(columns, table.Column{Title: c.Title, Width: max(maxWidth, len(c.Title))})
		for i, v := range values {
			rows[i] = append(rows[i], v)
		}
	}
	return rows, columns
}

func (l List) Init() tea.Cmd {
	return nil
}

func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return l, tea.Quit
		}
	case refreshListMsg:
		l = l.refreshTable()
	case tea.WindowSizeMsg:
		l.table.SetHeight(min(len(l.selection), msg.Height-len(detailsProjection)-3))
	}
	m, cmd := l.table.Update(msg)
	l.table = m

	if !l.interactive {
		return l, tea.Quit
	}
	return l, cmd
}

func (l List) View() string {
	builder := strings.Builder{}
	builder.WriteString(l.table.View())
	if l.interactive {
		builder.WriteString("\n\n")
		l.renderDetails(&builder)
	}
	return builder.String()
}

func (l List) renderDetails(writer io.StringWriter) {
	selectedItem := l.selection[l.table.Cursor()]
	detailsProjectionConfig := qprojection.Config{
		ColumnNames: detailsProjection,
		List:        l.list,
	}
	columns := qprojection.MustCompile(detailsProjectionConfig)
	var maxTitleWidth int
	for _, c := range columns {
		maxTitleWidth = max(maxTitleWidth, len(c.Title))
	}
	titleStyle := lipgloss.NewStyle().Width(maxTitleWidth).Align(lipgloss.Left)
	for _, c := range columns {
		writer.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, fmt.Sprintf("%s:\t", titleStyle.Render(c.Title)), c.Projector(selectedItem)))
		writer.WriteString("\n")
	}
}

func (l List) refreshTable() List {
	rows, columns := l.mapToColumns()
	l.table.SetColumns(columns)
	l.table.SetRows(rows)
	if l.interactive {
		l.table.Focus()
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
