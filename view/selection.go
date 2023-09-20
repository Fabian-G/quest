package view

import (
	"fmt"
	"io"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type selectionKeyMap struct {
	All     key.Binding
	Select  key.Binding
	None    key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func (s selectionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{s.All, s.None, s.Select, s.Confirm, s.Cancel}
}
func (s selectionKeyMap) FullHelp() [][]key.Binding {
	return nil
}

var defaultSelectionKeyMap = selectionKeyMap{
	All: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "All"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("␣", "Select"),
	),
	None: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "None"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("⏎", "Confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "Cancel"),
	),
}

type Selection struct {
	list      list.Model
	help      help.Model
	Cancelled bool
}

func NewSelection(choices []*todotxt.Item) Selection {
	l := list.New(toListItem(choices), listItemDelegate{list.NewDefaultItemStyles()}, 0, 0)
	l.SetShowHelp(false)
	l.Title = "Confirm Selection:"
	h := help.New()
	return Selection{
		list: l,
		help: h,
	}
}

type listItem struct {
	item     *todotxt.Item
	selected bool
}

func (i listItem) FilterValue() string {
	return i.item.Description()
}

func toListItem(items []*todotxt.Item) []list.Item {
	dItems := make([]list.Item, 0, len(items))
	for _, i := range items {
		dItems = append(dItems, list.Item(listItem{i, false}))
	}

	return dItems
}

type listItemDelegate struct {
	styles list.DefaultItemStyles
}

func (d listItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	txtItem := item.(listItem)
	var selectionMarker string = ">"
	if !txtItem.selected {
		selectionMarker = " "
	}
	line := fmt.Sprintf("%s %s", selectionMarker, txtItem.item.Description())
	if m.Index() == index {
		line = d.styles.SelectedTitle.Render(line)
	} else {
		line = d.styles.NormalTitle.Render(line)
	}
	fmt.Fprint(w, line)
}

func (d listItemDelegate) Height() int {
	return 1
}

func (d listItemDelegate) Spacing() int {
	return 0
}

func (d listItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (s Selection) Selection() []*todotxt.Item {
	listItems := s.list.Items()
	selection := make([]*todotxt.Item, 0, len(listItems))
	for _, item := range listItems {
		lItem := item.(listItem)
		if lItem.selected {
			selection = append(selection, lItem.item)
		}
	}
	return selection
}

func (s Selection) Init() tea.Cmd {
	return nil
}

func (s Selection) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetWidth(msg.Width)
		s.list.SetHeight(min(msg.Height-1, len(s.list.Items())+6))
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultSelectionKeyMap.All):
			s = s.selectAll()
		case key.Matches(msg, defaultSelectionKeyMap.None):
			s = s.selectNone()
		case key.Matches(msg, defaultSelectionKeyMap.Select):
			s = s.toggleCurrent()
		case key.Matches(msg, defaultSelectionKeyMap.Cancel):
			s.Cancelled = true
			return s, tea.Quit
		case key.Matches(msg, defaultSelectionKeyMap.Confirm):
			return s, tea.Quit
		}
	}
	var lcmd, hcmd tea.Cmd
	s.list, lcmd = s.list.Update(msg)
	s.help, hcmd = s.help.Update(msg)
	return s, tea.Batch(lcmd, hcmd)
}

func (s Selection) View() string {
	list := s.list.View()
	help := s.help.View(defaultSelectionKeyMap)
	return lipgloss.JoinVertical(lipgloss.Left, list, help)
}

func (s Selection) selectAll() Selection {
	items := s.list.Items()
	for idx := range items {
		lItem := items[idx].(listItem)
		lItem.selected = true
		items[idx] = lItem
	}
	s.list.SetItems(items)
	return s
}

func (s Selection) selectNone() Selection {
	items := s.list.Items()
	for idx := range items {
		lItem := items[idx].(listItem)
		lItem.selected = false
		items[idx] = lItem
	}
	s.list.SetItems(items)
	return s
}

func (s Selection) toggleCurrent() Selection {
	items := s.list.Items()
	lItem := items[s.list.Index()].(listItem)
	lItem.selected = !lItem.selected
	items[s.list.Index()] = lItem
	s.list.SetItems(items)
	return s
}
