package view

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

var detailsProjection = slices.DeleteFunc(slices.Clone(qprojection.StarProjection), func(s string) bool { return s == "tags" })

type List struct {
	list            *todotxt.List
	selection       []*todotxt.Item
	projection      []string
	projector       qprojection.Projector
	table           table.Model
	interactive     bool
	availableWidth  int
	availableHeight int
}

type RefreshListMsg struct {
	List       *todotxt.List
	Selection  []*todotxt.Item
	Projection []string
}

func NewList(proj qprojection.Projector, interactive bool) (List, error) {
	width, height, err := term.GetSize(0)
	if err != nil {
		return List{}, err
	}
	l := List{
		projector:       proj,
		interactive:     interactive,
		availableWidth:  width,
		availableHeight: height,
	}

	l.table = table.New()
	return l, nil
}

func (l List) mapToColumns() ([]table.Row, []table.Column) {
	headings, data := l.projector.MustProject(l.projection, l.list, l.selection) // It is the callers job to verify the projection
	if len(headings) == 0 || len(data) == 0 {
		return nil, nil
	}
	columns := make([]table.Column, 0, len(headings))
	rows := make([]table.Row, len(data))
	for i, h := range headings {
		maxWidth := 0
		values := make([]string, 0, len(data))
		for _, val := range data {
			values = append(values, val[i])
			maxWidth = max(maxWidth, len(val[i]))
		}
		if maxWidth == 0 {
			continue
		}

		columns = append(columns, table.Column{Title: h, Width: max(maxWidth, len(h))})
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
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return l, tea.Quit
		}
	case RefreshListMsg:
		l = l.refreshTable(msg.List, msg.Selection, msg.Projection)
	case tea.WindowSizeMsg:
		l.availableWidth = msg.Width
		l.availableHeight = msg.Height
		l = l.updateSize()
	}
	m, cmd := l.table.Update(msg)
	l.table = m

	return l, cmd
}

func (l List) updateSize() List {
	l.table.SetHeight(max(0, min(len(l.selection), l.availableHeight-len(detailsProjection)-3)))
	return l
}

func (l List) View() string {
	if len(l.selection) == 0 {
		return "no matches\n"
	}
	builder := strings.Builder{}
	builder.WriteString(l.table.View())
	builder.WriteString("\n")
	if l.interactive {
		builder.WriteString("\n")
		l.renderDetails(&builder)
	}
	return builder.String()
}

func (l List) renderDetails(writer io.StringWriter) {
	selectedItem := l.itemAtCursor()
	if selectedItem == nil {
		return
	}
	headings, data := l.projector.MustProject(detailsProjection, l.list, []*todotxt.Item{selectedItem})
	var maxTitleWidth int
	for _, c := range headings {
		maxTitleWidth = max(maxTitleWidth, len(c))
	}
	titleStyle := lipgloss.NewStyle().Width(maxTitleWidth).Align(lipgloss.Left)
	lines := make([]string, 0, len(data))
	for i, c := range headings {
		if len(data[0][i]) > 0 {
			title := fmt.Sprintf("%s:\t", titleStyle.Render(c))
			truncated := runewidth.Truncate(data[0][i], l.availableWidth-runewidth.StringWidth(title)-3, "...")
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, title, truncated))
		}
	}
	writer.WriteString(strings.Join(lines, "\n") + "\n")
}

func (l List) refreshTable(list *todotxt.List, selection []*todotxt.Item, projection []string) List {
	previous := l.itemAtCursor()
	l.list = list
	l.selection = selection
	l.projection = projection
	rows, columns := l.mapToColumns()
	l.table.SetRows(nil)
	l.table.SetColumns(nil)
	l.table.SetColumns(columns)
	l.table.SetRows(rows)
	if l.interactive {
		l = l.updateSize()
		l = l.moveCursorToItem(previous)
		l.table.Focus()
		l.table.SetStyles(table.DefaultStyles())
	} else {
		l.table.SetHeight(len(rows))
		defaultStyles := table.DefaultStyles()
		defaultStyles.Selected = lipgloss.NewStyle()
		l.table.SetStyles(defaultStyles)
	}
	return l
}

func (l List) itemAtCursor() *todotxt.Item {
	if 0 <= l.table.Cursor() && l.table.Cursor() < len(l.selection) {
		return l.selection[l.table.Cursor()]
	}
	return nil
}

func (l List) moveCursorToItem(target *todotxt.Item) List {
	var positionOfItem = -1
	if target != nil {
		positionOfItem = slices.IndexFunc(l.selection, func(i *todotxt.Item) bool {
			return i.Description() == target.Description()
		})
	}
	if positionOfItem == -1 {
		l.table.SetCursor(min(l.table.Cursor(), max(0, len(l.selection)-1)))
		return l
	}
	l.table.SetCursor(positionOfItem)
	return l
}
