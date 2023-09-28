package todotxt

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"
)

var DefaultJsonEncoder = JsonEncoder{}

type JsonEncoder struct {
}

type jsonItem struct {
	Line             int       `json:"line",omitempty`
	Done             bool      `json:"done,omitempty"`
	Priority         string    `json:"priority,omitempty"`
	Tags             Tags      `json:"tags,omitempty"`
	Contexts         []Context `json:"contexts,omitempty"`
	Projects         []Project `json:"projects,omitempty"`
	Creation         string    `json:"creation,omitempty"`
	Completion       string    `json:"completion,omitempty"`
	Description      string    `json:"description,omitempty"`
	CleanDescription string    `json:"clean_description,omitempty"`
}

func (f JsonEncoder) Encode(w io.Writer, list *List, tasks []*Item) error {
	out := bufio.NewWriter(w)
	jsonItems := make([]jsonItem, 0, len(tasks))
	for _, t := range tasks {
		projects, contexts, tags := t.Projects(), t.Contexts(), t.Tags()
		jsonItems = append(jsonItems, jsonItem{
			Line:             list.LineOf(t),
			Done:             t.Done(),
			Priority:         strings.Trim(t.prio.String(), "()"),
			Tags:             tags,
			Contexts:         contexts,
			Projects:         projects,
			Creation:         f.formatOrEmpty(t.CreationDate()),
			Completion:       f.formatOrEmpty(t.CompletionDate()),
			Description:      t.Description(),
			CleanDescription: t.CleanDescription(t.Projects(), t.Contexts(), tags.Keys()),
		})
	}
	json.NewEncoder(out).Encode(jsonItems)
	return out.Flush()
}

func (f JsonEncoder) formatOrEmpty(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format(time.DateOnly)
}
