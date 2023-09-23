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

var detailsProjection = strings.Split(strings.ReplaceAll(qprojection.StarProjection, ",tags", ""), ",")

type List struct {
	list            *todotxt.List
	selection       []*todotxt.Item
	extractors      []qprojection.Column
	table           table.Model
	interactive     bool
	availableWidth  int
	availableHeight int
}

type RefreshListMsg struct {
	List       *todotxt.List
	Selection  []*todotxt.Item
	Projection qprojection.Config
}

func NewList(list *todotxt.List, selection []*todotxt.Item, interactive bool, pCfg qprojection.Config) (List, error) {
	width, height, err := term.GetSize(0)
	if err != nil {
		return List{}, err
	}
	l := List{
		list:            list,
		selection:       selection,
		interactive:     interactive,
		availableWidth:  width,
		availableHeight: height,
	}

	columnExtractors, err := qprojection.Compile(pCfg)
	if err != nil {
		return List{}, fmt.Errorf("invalid projection %v: %w", pCfg.ColumnNames, err)
	}
	l.extractors = columnExtractors
	l.table = table.New()
	return l.refreshTable(list, selection, pCfg), nil
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
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return l, tea.Quit
		}
	case RefreshListMsg:
		l = l.refreshTable(msg.List, msg.Selection, msg.Projection)
	case tea.WindowSizeMsg:
		l.availableWidth = msg.Width
		l.availableHeight = msg.Height
		l.updateSize()
	}
	m, cmd := l.table.Update(msg)
	l.table = m

	return l, cmd
}

func (l List) updateSize() {
	l.table.SetHeight(max(0, min(len(l.selection), l.availableHeight-len(detailsProjection)-3)))
}

func (l List) View() string {
	if len(l.selection) == 0 {
		return "no matches"
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
	lines := make([]string, 0, len(columns))
	for _, c := range columns {
		if projection := c.Projector(selectedItem); len(projection) > 0 {
			title := fmt.Sprintf("%s:\t", titleStyle.Render(c.Title))
			truncated := runewidth.Truncate(projection, l.availableWidth-runewidth.StringWidth(title)-3, "...")
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, title, truncated))
		}
	}
	writer.WriteString(strings.Join(lines, "\n") + "\n")
}

func (l List) refreshTable(list *todotxt.List, selection []*todotxt.Item, projection qprojection.Config) List {
	previous := l.itemAtCursor()
	l.extractors = qprojection.MustCompile(projection)
	l.list = list
	l.selection = selection
	rows, columns := l.mapToColumns()
	l.table.SetRows(nil)
	l.table.SetColumns(nil)
	l.table.SetColumns(columns)
	l.table.SetRows(rows)
	if l.interactive {
		l.updateSize()
		l.table.Focus()
		l.table.SetStyles(table.DefaultStyles())
	} else {
		l.table.SetHeight(len(rows))
		defaultStyles := table.DefaultStyles()
		defaultStyles.Selected = lipgloss.NewStyle()
		l.table.SetStyles(defaultStyles)
	}
	l.table.UpdateViewport()
	l.moveCursorToItem(previous)
	return l
}

func (l List) itemAtCursor() *todotxt.Item {
	if 0 <= l.table.Cursor() && l.table.Cursor() < len(l.selection) {
		return l.selection[l.table.Cursor()]
	}
	return nil
}

func (l List) moveCursorToItem(target *todotxt.Item) {
	var positioOfItem = -1
	if target != nil {
		positioOfItem = slices.IndexFunc(l.selection, func(i *todotxt.Item) bool {
			return i.Description() == target.Description()
		})
	}
	if positioOfItem == -1 {
		l.table.SetCursor(min(l.table.Cursor(), max(0, len(l.selection)-1)))
		return
	}
	l.table.SetCursor(positioOfItem)
}
