package view

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

type Matcher interface {
	Match(string) bool
	fmt.Stringer
}

type ExtractionFunc func(ExtractionCtx) string

type ExtractionCtx struct {
	List          *todotxt.List
	Item          *todotxt.Item
	CleanTags     []string
	CleanProjects []todotxt.Project
	CleanContexts []todotxt.Context
}

type ColumnDef struct {
	matcher   Matcher
	name      func(string) string
	extractor func(string) ExtractionFunc
}

var columns = []ColumnDef{
	{idxMatcher, idxName, staticColumn(idxColumn)},
	{tagMatcher, tagName, tagColumn},
	{doneMatcher, doneName, staticColumn(doneColumn)},
	{creationMatcher, creationName, staticColumn(creationColumn)},
	{completionMatcher, CompletionName, staticColumn(completionColumn)},
	{projectsMatcher, projectsName, staticColumn(projectsColumn)},
	{contextsMatcher, contextsName, staticColumn(contextsColumn)},
	{descriptionMatcher, descriptionName, descriptionColumn},
}

func findColumn(key string) (string, ExtractionFunc) {
	for _, cDef := range columns {
		if cDef.matcher.Match(key) {
			return cDef.name(key), cDef.extractor(key)
		}
	}
	return "", nil
}

func availableColumns() []string {
	availableColumns := make([]string, 0, len(columns))
	for _, c := range columns {
		availableColumns = append(availableColumns, c.matcher.String())
	}
	return availableColumns
}

var tagMatcher = regexMatch("tag:.+")

func tagName(key string) string {
	return strings.Split(key, ":")[1]
}

func tagColumn(key string) ExtractionFunc {
	tagKey := strings.Split(key, ":")[1]
	return func(ctx ExtractionCtx) string {
		tagValues := ctx.Item.Tags()[tagKey]
		return strings.Join(tagValues, ",")
	}
}

var doneMatcher = staticMatch("done")
var doneName = staticName("Done")

func doneColumn(ctx ExtractionCtx) string {
	if ctx.Item.Done() {
		return "x"
	}
	return ""
}

var creationMatcher = staticMatch("creation")
var creationName = staticName("Created On")

func creationColumn(ctx ExtractionCtx) string {
	date := ctx.Item.CreationDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

var completionMatcher = staticMatch("completion")
var CompletionName = staticName("Completed On")

func completionColumn(ctx ExtractionCtx) string {
	date := ctx.Item.CompletionDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

var projectsMatcher = staticMatch("projects")
var projectsName = staticName("Projects")

func projectsColumn(ctx ExtractionCtx) string {
	projects := ctx.Item.Projects()
	projectStrings := make([]string, 0, len(projects))
	for _, p := range projects {
		projectStrings = append(projectStrings, p.String())
	}
	return strings.Join(projectStrings, ",")
}

var contextsMatcher = staticMatch("contexts")
var contextsName = staticName("Contexts")

func contextsColumn(ctx ExtractionCtx) string {
	contexts := ctx.Item.Contexts()
	contextStrings := make([]string, 0, len(contexts))
	for _, p := range contexts {
		contextStrings = append(contextStrings, p.String())
	}
	return strings.Join(contextStrings, ",")
}

var idxMatcher = staticMatch("idx")
var idxName = staticName("Idx")

func idxColumn(ctx ExtractionCtx) string {
	return strconv.Itoa(ctx.List.IndexOf(ctx.Item))
}

var descriptionMatcher = regexMatch("description(\\([0-9]+\\))?")
var descriptionName = staticName("Description")

func descriptionColumn(key string) ExtractionFunc {
	key = strings.TrimPrefix(key, "description")
	var width = math.MaxInt
	if len(key) > 0 {
		var err error
		width, err = strconv.Atoi(key[1 : len(key)-1])
		if err != nil {
			panic(err) // can not happen, because matcher ensures that there is a valid number
		}
	}
	return func(ctx ExtractionCtx) string {

		return runewidth.Truncate(ctx.Item.CleanDescription(ctx.CleanProjects, ctx.CleanContexts, ctx.CleanTags), width, "...")
	}
}

func regexMatch(columnRegex string) Matcher {
	regex := regexp.MustCompile(columnRegex)
	return regexMatcher{regex}
}

func staticMatch(columnKey string) Matcher {
	return staticMatcher{columnKey}
}

func staticColumn(f ExtractionFunc) func(string) ExtractionFunc {
	return func(s string) ExtractionFunc {
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

func (m regexMatcher) Match(key string) bool {
	return m.regex.MatchString(key)
}

func (m regexMatcher) String() string {
	return m.regex.String()
}

type staticMatcher struct {
	key string
}

func (m staticMatcher) Match(key string) bool {
	return m.key == key
}

func (m staticMatcher) String() string {
	return m.key
}
