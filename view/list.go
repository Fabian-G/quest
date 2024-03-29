package view

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var interactiveStyles = table.Styles{
	Selected: lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("3")),
	Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
	Cell:     lipgloss.NewStyle().Padding(0, 1),
}
var nonInteractiveStyles = table.Styles{
	Selected: lipgloss.NewStyle(),
	Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
	Cell:     lipgloss.NewStyle().Padding(0, 1),
}
var detailsProjection = slices.DeleteFunc(slices.Clone(qprojection.StarProjection), func(s string) bool { return s == "tags" })

type List struct {
	list            *todotxt.List
	repo            *todotxt.Repo
	selection       []*todotxt.Item
	projection      []string
	projector       qprojection.Projector
	getTasks        func(*todotxt.List) []*todotxt.Item
	table           table.Model
	interactive     bool
	availableWidth  int
	availableHeight int
}

type RefreshListMsg struct {
	List *todotxt.List
}

func NewList(repo *todotxt.Repo, proj qprojection.Projector, projection []string, getTasks func(*todotxt.List) []*todotxt.Item, interactive bool) List {
	l := List{
		repo:        repo,
		projector:   proj,
		projection:  projection,
		getTasks:    getTasks,
		interactive: interactive,
	}

	l.table = table.New()
	return l
}

func (l List) Run(initial *todotxt.List) error {
	model, _ := l.Update(RefreshListMsg{List: initial})
	l = model.(List)
	switch l.interactive {
	case true:
		programme := tea.NewProgram(l)
		data, end, err := l.repo.Watch()
		if err != nil {
			return err
		}
		defer end()
		go func() {
			for update := range data {
				newList, err := update()
				if err != nil {
					continue
				}
				programme.Send(RefreshListMsg{List: newList})
			}
		}()
		if _, err := programme.Run(); err != nil {
			return err
		}
	default:
		fmt.Print(l.View())
	}
	return nil
}

func (l List) mapToColumns() ([]table.Row, []table.Column, func(table.Model, string, table.CellPosition) string) {
	headings, data, styles := l.projector.MustProject(l.projection, l.list, l.selection) // It is the callers job to verify the projection
	if len(headings) == 0 || len(data) == 0 {
		return nil, nil, func(m table.Model, s string, cp table.CellPosition) string { return s }
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
			styles = deleteColumn(styles, len(columns))
			continue
		}

		columns = append(columns, table.Column{Title: h, Width: max(maxWidth, len(h))})
		for i, v := range values {
			rows[i] = append(rows[i], v)
		}
	}
	return rows, columns, func(m table.Model, s string, cp table.CellPosition) string {
		style := styles[cp.RowID][cp.Column]
		val := style.Padding(0, 1).Render(s)
		return strings.ReplaceAll(val, "\x1b[0m", "\x1b[39m")
	}
}

func deleteColumn(styles [][]lipgloss.Style, idx int) [][]lipgloss.Style {
	for i := 0; i < len(styles); i++ {
		styles[i] = slices.Delete(styles[i], idx, idx+1)
	}
	return styles
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
		l = l.refreshTable(msg.List)
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

func (l List) renderDetails(writer *strings.Builder) {
	selectedItem := l.itemAtCursor()
	if selectedItem == nil {
		return
	}
	headings, data, _ := l.projector.MustProject(detailsProjection, l.list, []*todotxt.Item{selectedItem})
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

func (l List) refreshTable(list *todotxt.List) List {
	previous := l.itemAtCursor()
	l.list = list
	l.selection = l.getTasks(list)
	rows, columns, renderCell := l.mapToColumns()
	l.table.SetStyles(l.styles(renderCell))
	l.table.SetRows(nil)
	l.table.SetColumns(nil)
	l.table.SetColumns(columns)
	l.table.SetRows(rows)
	if l.interactive {
		l = l.updateSize()
		l = l.moveCursorToItem(previous)
		l.table.Focus()
	} else {
		l.table.SetHeight(len(rows))
	}
	return l
}

func (l List) styles(renderCell func(table.Model, string, table.CellPosition) string) table.Styles {
	if l.interactive {
		styles := interactiveStyles
		styles.RenderCell = renderCell
		return styles
	}
	styles := nonInteractiveStyles
	styles.RenderCell = renderCell
	return styles
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
