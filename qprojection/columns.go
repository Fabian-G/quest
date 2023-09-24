package qprojection

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/mattn/go-runewidth"
)

const StarProjection = "idx,done,priority,completion,creation,projects,contexts,tags,description"

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
	idxColumn,
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
			tagValues := i.Tags()[tagKey]
			return strings.Join(tagValues, ",")
		}
	},
}

var doneColumn = columnDef{
	matcher: staticMatch("done"),
	name:    staticName("Done"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		if item.Done() {
			return "x"
		}
		return ""
	}),
}

var priorityColumn = columnDef{
	matcher: staticMatch("priority"),
	name:    staticName("Priority"),
	extractor: staticColumn(func(p Projector, list *todotxt.List, item *todotxt.Item) string {
		return item.Priority().String()
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
		return date.Format(time.DateOnly)
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
		return date.Format(time.DateOnly)
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
		return strings.Join(projectStrings, ",")
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
		return strings.Join(contextStrings, ",")
	}),
}

var idxColumn = columnDef{
	matcher: staticMatch("idx"),
	name:    staticName("Idx"),
	extractor: func(key string) exFunc {
		return func(p Projector, list *todotxt.List, item *todotxt.Item) string {
			return strconv.Itoa(list.IndexOf(item))
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
			return runewidth.Truncate(item.CleanDescription(p.expandClean(l)), width, "...")
		}
	},
}

var questScoreColumn = columnDef{
	matcher: staticMatch("score"),
	name:    staticName("Score"),
	extractor: staticColumn(func(p Projector, l *todotxt.List, i *todotxt.Item) string {
		result := p.ScoreCalc.ScoreOf(i)
		var score string
		if result.Score >= 10 {
			score = fmt.Sprintf("%.0f.", result.Score)
		} else {
			score = fmt.Sprintf("%.1f", result.Score)
		}
		urgentFlag := "U"
		importantFlag := "I"
		if !result.IsImportant() {
			importantFlag = " "
		}
		if !result.IsUrgent() {
			urgentFlag = " "
		}
		return fmt.Sprintf("%s %s%s", score, urgentFlag, importantFlag)
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
