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

type matcher interface {
	Match(string) bool
	fmt.Stringer
}

type extractionFunc func(extractionCtx) string

type extractionCtx struct {
	list          *todotxt.List
	item          *todotxt.Item
	cleanTags     []string
	cleanProjects []todotxt.Project
	cleanContexts []todotxt.Context
}

type columnDef struct {
	matcher   matcher
	name      func(string) string
	extractor func(string) extractionFunc
}

var columns = []columnDef{
	{idxMatcher, idxName, staticColumn(idxColumn)},
	{tagMatcher, tagName, tagColumn},
	{doneMatcher, doneName, staticColumn(doneColumn)},
	{creationMatcher, creationName, staticColumn(creationColumn)},
	{completionMatcher, CompletionName, staticColumn(completionColumn)},
	{projectsMatcher, projectsName, staticColumn(projectsColumn)},
	{contextsMatcher, contextsName, staticColumn(contextsColumn)},
	{descriptionMatcher, descriptionName, descriptionColumn},
}

func findColumn(key string) (string, extractionFunc) {
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

func tagColumn(key string) extractionFunc {
	tagKey := strings.Split(key, ":")[1]
	return func(ctx extractionCtx) string {
		tagValues := ctx.item.Tags()[tagKey]
		return strings.Join(tagValues, ",")
	}
}

var doneMatcher = staticMatch("done")
var doneName = staticName("Done")

func doneColumn(ctx extractionCtx) string {
	if ctx.item.Done() {
		return "x"
	}
	return ""
}

var creationMatcher = staticMatch("creation")
var creationName = staticName("Created On")

func creationColumn(ctx extractionCtx) string {
	date := ctx.item.CreationDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

var completionMatcher = staticMatch("completion")
var CompletionName = staticName("Completed On")

func completionColumn(ctx extractionCtx) string {
	date := ctx.item.CompletionDate()
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}

var projectsMatcher = staticMatch("projects")
var projectsName = staticName("Projects")

func projectsColumn(ctx extractionCtx) string {
	projects := ctx.item.Projects()
	projectStrings := make([]string, 0, len(projects))
	for _, p := range projects {
		projectStrings = append(projectStrings, p.String())
	}
	return strings.Join(projectStrings, ",")
}

var contextsMatcher = staticMatch("contexts")
var contextsName = staticName("Contexts")

func contextsColumn(ctx extractionCtx) string {
	contexts := ctx.item.Contexts()
	contextStrings := make([]string, 0, len(contexts))
	for _, p := range contexts {
		contextStrings = append(contextStrings, p.String())
	}
	return strings.Join(contextStrings, ",")
}

var idxMatcher = staticMatch("idx")
var idxName = staticName("Idx")

func idxColumn(ctx extractionCtx) string {
	return strconv.Itoa(ctx.list.IndexOf(ctx.item))
}

var descriptionMatcher = regexMatch("description(\\([0-9]+\\))?")
var descriptionName = staticName("Description")

func descriptionColumn(key string) extractionFunc {
	key = strings.TrimPrefix(key, "description")
	var width = math.MaxInt
	if len(key) > 0 {
		var err error
		width, err = strconv.Atoi(key[1 : len(key)-1])
		if err != nil {
			panic(err) // can not happen, because matcher ensures that there is a valid number
		}
	}
	return func(ctx extractionCtx) string {

		return runewidth.Truncate(ctx.item.CleanDescription(ctx.cleanProjects, ctx.cleanContexts, ctx.cleanTags), width, "...")
	}
}

func regexMatch(columnRegex string) matcher {
	regex := regexp.MustCompile(columnRegex)
	return regexMatcher{regex}
}

func staticMatch(columnKey string) matcher {
	return staticMatcher{columnKey}
}

func staticColumn(f extractionFunc) func(string) extractionFunc {
	return func(s string) extractionFunc {
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
