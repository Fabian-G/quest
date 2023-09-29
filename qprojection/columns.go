package qprojection

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var StarProjection = []string{
	"line", "done", "priority", "completion", "creation", "projects", "contexts", "tags", "description",
}

type matcher interface {
	match(string) bool
	fmt.Stringer
}

type exFunc func(Projector, *todotxt.List, *todotxt.Item) string

type columnDef struct {
	matcher   matcher
	name      func(string) string
	extractor func(string) exFunc
}

var columns = []columnDef{
	lineColumn,
	tagColumn,
	doneColumn,
	priorityColumn,
	creationColumn,
	completionColumn,
	projectsColumn,
	contextsColumn,
	descriptionColumn,
	questScoreColumn,
}

func availableColumns() []string {
	availableColumns := make([]string, 0, len(columns))
	for _, c := range columns {
		availableColumns = append(availableColumns, c.matcher.String())
	}
	return availableColumns
}

var tagColumn = columnDef{
	matcher: regexMatch("tag:.+"),
	name: func(key string) string {
		return strings.Split(key, ":")[1]
	},
	extractor: func(key string) exFunc {
		tagKey := strings.Split(key, ":")[1]
		return func(p Projector, l *todotxt.List, i *todotxt.Item) string {
			var tagValues []string
			if p.TagTypes[tagKey] == qselect.QDate && slices.Contains(p.HumanizedTags, tagKey) {
				for _, v := range i.Tags()[tagKey] {
					d, err := time.Parse(time.DateOnly, v)
					if err != nil {
						tagValues = append(tagValues, v)
						continue
					}
					tagValues = append(tagValues, humanTime(d))
				}
			} else {
				tagValues = i.Tags()[tagKey]
			}
			var color lipgloss.Color
			if f, ok := p.TagColors[tagKey]; ok {
				if c := f(l, i); c != nil {
					color = *c
				}
			}
			return p.colorize(color, strings.Join(tagValues, ","))
		}
	},
}

var doneColumn = columnDef{
	matcher: staticMatch("done"),
	name:    staticName("Done"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		if item.Done() {
			return p.colorize(p.defaultColor, "x")
		}
		return ""
	}),
}

var priorityColumn = columnDef{
	matcher: staticMatch("priority"),
	name:    staticName("Priority"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		prio := item.Priority().String()
		switch item.Priority() {
		case todotxt.PrioA:
			return p.colorize(lipgloss.Color("1"), prio)
		case todotxt.PrioB:
			return p.colorize(lipgloss.Color("2"), prio)
		case todotxt.PrioC:
			return p.colorize(lipgloss.Color("3"), prio)
		default:
			return p.colorize(p.defaultColor, prio)
		}
	}),
}

var creationColumn = columnDef{
	matcher: staticMatch("creation"),
	name:    staticName("Created On"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		date := item.CreationDate()
		if date == nil {
			return ""
		}
		return p.colorize(p.defaultColor, date.Format(time.DateOnly))
	}),
}

var completionColumn = columnDef{
	matcher: staticMatch("completion"),
	name:    staticName("Completed On"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		date := item.CompletionDate()
		if date == nil {
			return ""
		}
		return p.colorize(p.defaultColor, date.Format(time.DateOnly))
	}),
}

var projectsColumn = columnDef{
	matcher: staticMatch("projects"),
	name:    staticName("Projects"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		projects := item.Projects()
		projectStrings := make([]string, 0, len(projects))
		for _, p := range projects {
			projectStrings = append(projectStrings, p.String())
		}
		return p.colorize(p.defaultColor, strings.Join(projectStrings, ","))
	}),
}

var contextsColumn = columnDef{
	matcher: staticMatch("contexts"),
	name:    staticName("Contexts"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		contexts := item.Contexts()
		contextStrings := make([]string, 0, len(contexts))
		for _, p := range contexts {
			contextStrings = append(contextStrings, p.String())
		}
		return p.colorize(p.defaultColor, strings.Join(contextStrings, ","))
	}),
}

var lineColumn = columnDef{
	matcher: staticMatch("line"),
	name:    staticName("#"),
	extractor: func(key string) exFunc {
		return func(p Projector, list *todotxt.List, item *todotxt.Item) string {
			return p.colorize(p.defaultColor, strconv.Itoa(list.LineOf(item)))
		}
	},
}

var descriptionColumn = columnDef{
	matcher: regexMatch("description(\\([0-9]+\\))?"),
	name:    staticName("Description"),
	extractor: func(key string) exFunc {
		key = strings.TrimPrefix(key, "description")
		var width = math.MaxInt
		if len(key) > 0 {
			var err error
			width, err = strconv.Atoi(key[1 : len(key)-1])
			if err != nil {
				panic(err) // can not happen, because matcher ensures that there is a valid number
			}
		}
		return func(p Projector, l *todotxt.List, item *todotxt.Item) string {
			return p.colorize(p.defaultColor, runewidth.Truncate(item.CleanDescription(p.expandClean(l)), width, "..."))
		}
	},
}

var questScoreColumn = columnDef{
	matcher: staticMatch("score"),
	name:    staticName("Score"),
	extractor: staticColumn(func(p Projector, l *todotxt.List, i *todotxt.Item) string {
		result := p.ScoreCalc.ScoreOf(i)
		var score string
		switch {
		case result.Score >= 10:
			score = fmt.Sprintf("%.0f.", result.Score)
		default:
			score = fmt.Sprintf("%.1f", result.Score)
		}
		color := p.defaultColor
		switch {
		case result.IsImportant() && result.IsUrgent():
			color = lipgloss.Color("2")
		case result.IsImportant():
			color = lipgloss.Color("6")
		case result.IsUrgent():
			color = lipgloss.Color("1")
		}
		return p.colorize(color, score)
	}),
}

func regexMatch(columnRegex string) matcher {
	regex := regexp.MustCompile(columnRegex)
	return regexMatcher{regex}
}

func staticMatch(columnKey string) matcher {
	return staticMatcher{columnKey}
}

func staticColumn(f exFunc) func(string) exFunc {
	return func(key string) exFunc {
		return f
	}
}

func staticName(columnKey string) func(string) string {
	return func(s string) string {
		return columnKey
	}
}

type regexMatcher struct {
	regex *regexp.Regexp
}

func (m regexMatcher) match(key string) bool {
	return m.regex.MatchString(key)
}

func (m regexMatcher) String() string {
	return m.regex.String()
}

type staticMatcher struct {
	key string
}

func (m staticMatcher) match(key string) bool {
	return m.key == key
}

func (m staticMatcher) String() string {
	return m.key
}
