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

type Func func(*todotxt.Item) string

type columnDef struct {
	matcher   matcher
	name      func(Config, string) string
	extractor func(Config, string) Func
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
}

func findColumn(ctx Config, key string) (Column, error) {
	for _, cDef := range columns {
		if cDef.matcher.match(key) {
			return Column{cDef.name(ctx, key), cDef.extractor(ctx, key)}, nil
		}
	}
	return Column{}, fmt.Errorf("column %s not found. Available columns are: %v", key, availableColumns())
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
	name: func(cfg Config, key string) string {
		return strings.Split(key, ":")[1]
	},
	extractor: func(cfg Config, key string) Func {
		tagKey := strings.Split(key, ":")[1]
		return func(i *todotxt.Item) string {
			tagValues := i.Tags()[tagKey]
			return strings.Join(tagValues, ",")
		}
	},
}

var doneColumn = columnDef{
	matcher: staticMatch("done"),
	name:    staticName("Done"),
	extractor: staticColumn(func(i *todotxt.Item) string {
		if i.Done() {
			return "x"
		}
		return ""
	}),
}

var priorityColumn = columnDef{
	matcher: staticMatch("priority"),
	name:    staticName("Priority"),
	extractor: staticColumn(func(i *todotxt.Item) string {
		return i.Priority().String()
	}),
}

var creationColumn = columnDef{
	matcher: staticMatch("creation"),
	name:    staticName("Created On"),
	extractor: staticColumn(func(i *todotxt.Item) string {
		date := i.CreationDate()
		if date == nil {
			return ""
		}
		return date.Format(time.DateOnly)
	}),
}

var completionColumn = columnDef{
	matcher: staticMatch("completion"),
	name:    staticName("Completed On"),
	extractor: staticColumn(func(i *todotxt.Item) string {
		date := i.CompletionDate()
		if date == nil {
			return ""
		}
		return date.Format(time.DateOnly)
	}),
}

var projectsColumn = columnDef{
	matcher: staticMatch("projects"),
	name:    staticName("Projects"),
	extractor: staticColumn(func(i *todotxt.Item) string {
		projects := i.Projects()
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
	extractor: staticColumn(func(i *todotxt.Item) string {
		contexts := i.Contexts()
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
	extractor: func(cfg Config, key string) Func {
		return func(i *todotxt.Item) string {
			return strconv.Itoa(cfg.List.IndexOf(i))
		}
	},
}

var descriptionColumn = columnDef{
	matcher: regexMatch("description(\\([0-9]+\\))?"),
	name:    staticName("Description"),
	extractor: func(cfg Config, key string) Func {
		key = strings.TrimPrefix(key, "description")
		var width = math.MaxInt
		if len(key) > 0 {
			var err error
			width, err = strconv.Atoi(key[1 : len(key)-1])
			if err != nil {
				panic(err) // can not happen, because matcher ensures that there is a valid number
			}
		}
		return func(item *todotxt.Item) string {
			return runewidth.Truncate(item.CleanDescription(cfg.CleanProjects, cfg.CleanContexts, cfg.CleanTags), width, "...")
		}
	},
}

func regexMatch(columnRegex string) matcher {
	regex := regexp.MustCompile(columnRegex)
	return regexMatcher{regex}
}

func staticMatch(columnKey string) matcher {
	return staticMatcher{columnKey}
}

func staticColumn(f Func) func(Config, string) Func {
	return func(cfg Config, s string) Func {
		return f
	}
}

func staticName(columnKey string) func(Config, string) string {
	return func(cfg Config, s string) string {
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
