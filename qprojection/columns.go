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

const StarProjection = "idx,done,completion,creation,projects,contexts,tags,description"

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
	{idxMatcher, idxName, idxColumn},
	{tagMatcher, tagName, tagColumn},
	{doneMatcher, doneName, staticColumn(doneColumn)},
	{creationMatcher, creationName, staticColumn(creationColumn)},
	{completionMatcher, CompletionName, staticColumn(completionColumn)},
	{projectsMatcher, projectsName, staticColumn(projectsColumn)},
	{contextsMatcher, contextsName, staticColumn(contextsColumn)},
	{descriptionMatcher, descriptionName, descriptionColumn},
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

var tagMatcher = regexMatch("tag:.+")

func tagName(cfg Config, key string) string {
	return strings.Split(key, ":")[1]
}

func tagColumn(cfg Config, key string) Func {
	tagKey := strings.Split(key, ":")[1]
	return func(i *todotxt.Item) string {
		tagValues := i.Tags()[tagKey]
		return strings.Join(tagValues, ",")
	}
}

var doneMatcher = staticMatch("done")
var doneName = staticName("Done")

func doneColumn(i *todotxt.Item) string {
	if i.Done() {
		return "x"
	}
	return ""
}

var creationMatcher = staticMatch("creation")
var creationName = staticName("Created On")

func creationColumn(i *todotxt.Item) string {
	date := i.CreationDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

var completionMatcher = staticMatch("completion")
var CompletionName = staticName("Completed On")

func completionColumn(i *todotxt.Item) string {
	date := i.CompletionDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

var projectsMatcher = staticMatch("projects")
var projectsName = staticName("Projects")

func projectsColumn(i *todotxt.Item) string {
	projects := i.Projects()
	projectStrings := make([]string, 0, len(projects))
	for _, p := range projects {
		projectStrings = append(projectStrings, p.String())
	}
	return strings.Join(projectStrings, ",")
}

var contextsMatcher = staticMatch("contexts")
var contextsName = staticName("Contexts")

func contextsColumn(i *todotxt.Item) string {
	contexts := i.Contexts()
	contextStrings := make([]string, 0, len(contexts))
	for _, p := range contexts {
		contextStrings = append(contextStrings, p.String())
	}
	return strings.Join(contextStrings, ",")
}

var idxMatcher = staticMatch("idx")
var idxName = staticName("Idx")

func idxColumn(cfg Config, key string) Func {
	return func(i *todotxt.Item) string {
		return strconv.Itoa(cfg.List.IndexOf(i))
	}
}

var descriptionMatcher = regexMatch("description(\\([0-9]+\\))?")
var descriptionName = staticName("Description")

func descriptionColumn(cfg Config, key string) Func {
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
